package pubsubmanager

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func get_info(id string, ch <-chan interface{}) {
	for m := range ch {
		fmt.Printf("id %s get info: %v \n", id, m)
	}
}
func pub_info(cancels ...CloseListenerFunc) {
	for _, m := range []int{0, 1, 2, 3, 4, 5, 6} {
		PubSub.Publish(m)
	}
	for _, cancel := range cancels {
		cancel()
	}
}

func TestSubFirst(t *testing.T) {
	PubSub.AddChannel("1")
	subch1, closech1, cancel1, err1 := PubSub.RegistListener("1", 0)
	if err1 != nil {
		assert.FailNow(t, err1.Error(), "RegistListener get error")
	}
	subch2, closech2, cancel2, err2 := PubSub.RegistListener("1", 0)
	if err1 != nil {
		assert.FailNow(t, err2.Error(), "RegistListener get error")
	}

	go get_info("1", subch1)
	go get_info("2", subch2)
	go pub_info(cancel1, cancel2)
	<-closech1
	<-closech2
	t.Log("ok")
}
