package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleMessage(message []byte, topic kafka.TopicPartition, cn int) error {
	logrus.Infof("Consumer #%d, Message from kafka with offset %d '%s' on partition %d", cn, topic.Offset, string(message), topic.Partition)
	data := map[string]interface{}{
		"to_addr":  string(message),
		"subject":  "Teen Titans GO",
		"template": "hello_email",
		"vars": map[string]string{
			"Name": "User",
		},
	}

	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(
		"http://localhost:8090/html_email",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", body)
	return nil
}
