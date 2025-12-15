package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"encoding/json"

	"hotel-booking-system/package/events"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleMessage(message []byte, topic kafka.TopicPartition, cn int) error {
	var event events.BookingCreatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		logrus.Errorf("Failed to parse event JSON: %v", err)
		return nil
	}
	logrus.Infof("Processing booking ID: %d for email: %s", event.BookingID, event.UserEmail)
	data := map[string]interface{}{
		"to_addr":  event.UserEmail,
		"subject":  "Подтверждение бронирования",
		"template": "hello_email",
		"vars": map[string]string{
			"UserName":  event.UserName,
			"BookingID": fmt.Sprintf("%d", event.BookingID),
			"UserEmail": event.UserEmail,
			"Amount":    fmt.Sprintf("%.2f", event.Amount),
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
