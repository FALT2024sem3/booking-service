package main

import (
	"fmt"

	"github.com/sirupsen/logrus"

	k "hotel-booking-system/internal/kafka"
)

const (
	topic = "my-topic"
)

var address = []string{"localhost:9091", "localhost:9092", "localhost:9093"}

func main() {
	fmt.Println("Starting producer...")
	p, err := k.NewProducer(address)
	if err != nil {
		logrus.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		msg := fmt.Sprintf("0955452@mail.ru")
		if err = p.Produce(msg, topic); err != nil {
			logrus.Error(err)
		}
	}
	fmt.Println("Finished producing messages")
}
