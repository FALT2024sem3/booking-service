package handler

import (
	"encoding/json"
	"fmt"

	"hotel-booking-system/internal/notification"
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

	reqBody := notification.EmailWithTemplateRequestBody{
		ToAddr:   event.UserEmail,
		Subject:  "Подтверждение бронирования",
		Template: "hello_email",
		Vars: map[string]string{
			"UserName":  event.UserName,
			"BookingID": fmt.Sprintf("%d", event.BookingID),
			"UserEmail": event.UserEmail,
			"Amount":    fmt.Sprintf("%.2f", event.Amount),
		},
	}

	if err := notification.SendEmailLogic(reqBody); err != nil {
		logrus.Errorf("Failed to send email: %v", err)
		return nil
	}

	logrus.Info("Email sent successfully via Kafka handler")
	return nil
}
