package closeio

import (
	"log"
	"os"
	"testing"
)

func TestStatusList(t *testing.T) {
	key := os.Getenv("CLOSEIO_KEY")
	closeAPI := New(key)
	statuses, err := closeAPI.Statuses()
	if err != nil {
		log.Println(err)
	}
	log.Println(statuses)
}
