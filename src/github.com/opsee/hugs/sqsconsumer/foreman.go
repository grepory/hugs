package sqsconsumer

import (
	"math"
	"strconv"
	"time"

	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/store"
)

// As in Job Site
type Site struct {
	QueueUrl           string
	DBUrl              string
	WorkerPool         chan chan ForemanCommand
	Store              *store.Postgres
	CurrentWorkerCount *int64
	siteSQS            *sqs.SQS
}

// Get approximate number of messages remaining in queue
func (site *Site) GetJobsRemaining() (int, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: []*string{aws.String("ApproximateNumberOfMessages")},
		QueueUrl:       aws.String(site.QueueUrl),
	}
	out, err := site.siteSQS.GetQueueAttributes(input)
	if err != nil {
		return -1, err
	}
	messagesString := aws.StringValue(out.Attributes["ApproximateNumberOfMessages"])
	logrus.WithFields(logrus.Fields{"Event": "GetJobsRemaining"}).Info("Approximately ", messagesString, " messages remaining in the queue")

	count, err := strconv.Atoi(messagesString)
	if err != nil {
		return -1, err
	}

	return count, nil
}

type ForemanCommand int

const (
	Quit ForemanCommand = 0
)

// Foreman manages # and behavior of workers at Site
type Foreman struct {
	Site                  *Site
	CurrentWorkerCount    int64
	TargetWorkerCount     int64
	MaxWorkerCount        int64
	MinWorkerCount        int64
	OptimalWorkEstimate   int64
	PreviousWorkEstimates [6]int
	UpdateFreqSec         int
	CommandChan           chan ForemanCommand
}

func NewForeman(updateFreqSec int, optimalWorkEstimate int64, maxWorkers int64, minWorkers int64, sqsURL string, dbURL string) *Foreman {
	sqs := sqs.New(config.GetConfig().AWSSession)
	foreman := &Foreman{
		Site: &Site{
			QueueUrl:           sqsURL,
			DBUrl:              dbURL,
			CurrentWorkerCount: aws.Int64(minWorkers),
			WorkerPool:         make(chan chan ForemanCommand, maxWorkers),
			siteSQS:            sqs,
		},
		TargetWorkerCount:     minWorkers,
		CurrentWorkerCount:    minWorkers,
		MaxWorkerCount:        maxWorkers,
		MinWorkerCount:        minWorkers,
		OptimalWorkEstimate:   optimalWorkEstimate,
		PreviousWorkEstimates: [6]int{},
		UpdateFreqSec:         updateFreqSec, // we assume that all tasks will finish within 1 minute
		CommandChan:           make(chan ForemanCommand, maxWorkers),
	}

	return foreman
}

// Ensure that the computed target worker count is >= MinWorkerCount and <= MaxWorkerCount
func (foreman *Foreman) safeTargetWorkerCount(count int64) int64 {
	if count < foreman.MinWorkerCount {
		count = foreman.MinWorkerCount
	}
	if count > foreman.MaxWorkerCount {
		count = foreman.MaxWorkerCount
	}
	return count
}

// put a lower bound on estimated work to prevent us from killing off too many workers (optimal is 1k)
func (foreman *Foreman) safeWorkEstimate(count int) int {
	if int64(count) < foreman.OptimalWorkEstimate {
		// TODO(dan) fix this casting mess
		count = int(foreman.OptimalWorkEstimate)
	}
	return count
}

// calculate target worker count from estimated load
func (foreman *Foreman) ComputeWorkerTarget(load int64) int64 {
	est := int64(math.Pow(math.Log(float64(load)), 2.5))
	return foreman.safeTargetWorkerCount(est)
}

// Estimate current work and set target worker count
func (foreman *Foreman) EstimateWork() {
	count, err := foreman.Site.GetJobsRemaining()
	if err != nil {
		logrus.WithFields(logrus.Fields{"Event": "EstimateWork", "Error": err}).Warn("Couldn't get SQS Queue Attributes")
	}
	count = foreman.safeWorkEstimate(count)
	sma := foreman.computeJobCountSMA(count)
	foreman.TargetWorkerCount = foreman.ComputeWorkerTarget(int64(sma))
	logrus.Info("Foreman: Load SMA: ", sma, ". Currently ", foreman.CurrentWorkerCount, " workers. Target is ", foreman.TargetWorkerCount, " workers.")
}

// Initialize sma for past n estimates to current estimated work
func (foreman *Foreman) initWorkEstimateHistory() {
	count, err := foreman.Site.GetJobsRemaining()
	if err != nil {
		logrus.WithFields(logrus.Fields{"Event": "EstimateWork", "Error": err}).Warn("Couldn't get SQS Queue Attributes")
	}
	count = foreman.safeWorkEstimate(count)
	for i := 0; i < len(foreman.PreviousWorkEstimates); i++ {
		foreman.PreviousWorkEstimates[i] = count
	}
}

//  Add or remove workers based on current estimated workload
func (foreman *Foreman) AdjustWorkerCount() {
	foreman.CurrentWorkerCount = atomic.LoadInt64(foreman.Site.CurrentWorkerCount)
	diff := foreman.TargetWorkerCount - foreman.CurrentWorkerCount
	if diff > 0 {
		var i int64
		for i = 0; i < diff; i++ {
			if foreman.CurrentWorkerCount < foreman.MaxWorkerCount {
				foreman.recruitWorker()
			} else {
				logrus.Warn("Foreman: Hit worker ceiling.")
				break
			}
		}
	} else {
		var i int64
		for i = 0; i < diff*-1; i++ {
			if foreman.CurrentWorkerCount > foreman.MinWorkerCount {
				foreman.discardWorker()
			} else {
				logrus.Warn("Foreman: Hit worker floor.")
				break
			}
		}
	}
}

// compute simple moving average for job counts
func (foreman *Foreman) computeJobCountSMA(currentEstimate int) int {
	avg := 0.0
	len := len(foreman.PreviousWorkEstimates)
	for i := 1; i < len; i++ {
		foreman.PreviousWorkEstimates[i-1] = foreman.PreviousWorkEstimates[i]
		avg += float64(foreman.PreviousWorkEstimates[i-1])
	}
	foreman.PreviousWorkEstimates[len-1] = currentEstimate

	avg += float64(foreman.PreviousWorkEstimates[len-1])
	avg /= float64(len)

	return int(avg)
}

// Get rid of a worker
func (foreman *Foreman) discardWorker() {
	go func() {
		foreman.CommandChan <- Quit
	}()
	//TODO(dan) this should be done in the go func and threadsafe
}

// Get a new worker
func (foreman *Foreman) recruitWorker() {
	worker := NewWorker(foreman.Site)
	worker.Start()
}

// send out commands to workers in Site's WorkerPool
func (foreman *Foreman) issueCommands() {
	for {
		select {
		case command := <-foreman.CommandChan:
			go func() {
				workerChan := <-foreman.Site.WorkerPool
				workerChan <- command
			}()
		}
	}
}

// Start managed worker pool
func (foreman *Foreman) Start() {
	foreman.InitWorkers()
	go foreman.issueCommands()
	foreman.initWorkEstimateHistory()

	for {
		ta := time.Now()
		logrus.Info("Foreman: Start cycle @", ta)
		foreman.EstimateWork()
		foreman.AdjustWorkerCount()
		elapsed := time.Since(ta)

		wait := time.Duration(foreman.UpdateFreqSec)*time.Second - elapsed

		if wait < time.Duration(0)*time.Second {
			logrus.WithFields(logrus.Fields{"Event": "Foreman Wait"}).Warn("Foreman couldn't finish tasks in allotted time!")
			continue
		}

		time.Sleep(wait) // wait the rest of the minute
		logrus.Info("Foreman: Finished cycle @", time.Now())
	}
}

func (foreman *Foreman) InitWorkers() {
	// starting n number of workers
	var i int64
	for i = 0; i < foreman.CurrentWorkerCount; i++ {
		worker := NewWorker(foreman.Site)
		worker.Start()
	}
}
