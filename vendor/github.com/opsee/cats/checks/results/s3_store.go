package results

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gogo/protobuf/proto"
	"github.com/opsee/basic/schema"
	log "github.com/opsee/logrus"
)

// S3Store stores CheckResult objects in S3 by ResultId (check_id:bastion_id).
type S3Store struct {
	BucketName string
	S3Client   *s3.S3
}

// GetResultByCheckId gets the latest CheckResult for a Check from persistent storage.
func (s *S3Store) GetResultByCheckId(bastionId, checkId string) (result *schema.CheckResult, err error) {
	resultPath := fmt.Sprintf("%s/%s/latest.pb", checkId, bastionId)
	log.Infof("fetching result from s3://%s/%s", s.BucketName, resultPath)

	getObjResp, err := s.S3Client.GetObject(&s3.GetObjectInput{
		Bucket:              aws.String(s.BucketName),
		Key:                 aws.String(resultPath),
		ResponseContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(getObjResp.Body)
	getObjResp.Body.Close()
	if err != nil {
		return nil, err
	}

	result = &schema.CheckResult{}

	if err := proto.Unmarshal(bodyBytes, result); err != nil {
		return nil, err
	}

	return result, nil
}

// PutResult puts a CheckResult to persistent storage.
func (s *S3Store) PutResult(result *schema.CheckResult) error {
	resultPath := fmt.Sprintf("%s/%s/latest.pb", result.CheckId, result.BastionId)

	resultBytes, err := proto.Marshal(result)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(resultBytes)

	_, err = s.S3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(resultPath),
		Body:   reader,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *S3Store) GetCheckSnapshot(transitionId int64, checkId string) (check *schema.Check, err error) {
	snapshotPath := fmt.Sprintf("%s/snapshots/%d.pb", checkId, transitionId)

	getObjResp, err := s.S3Client.GetObject(&s3.GetObjectInput{
		Bucket:              aws.String(s.BucketName),
		Key:                 aws.String(snapshotPath),
		ResponseContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(getObjResp.Body)
	getObjResp.Body.Close()
	if err != nil {
		return nil, err
	}

	check = &schema.Check{}

	if err := proto.Unmarshal(bodyBytes, check); err != nil {
		return nil, err
	}

	return check, nil
}

func (s *S3Store) PutCheckSnapshot(transitionId int64, check *schema.Check) error {
	snapshotPath := fmt.Sprintf("%s/snapshots/%d.pb", check.Id, transitionId)
	checkBytes, err := proto.Marshal(check)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(checkBytes)

	_, err = s.S3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(snapshotPath),
		Body:   reader,
	})
	if err != nil {
		return err
	}

	return nil
}
