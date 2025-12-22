package main

import (
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"hotel-booking-system/internal/handler"
	"hotel-booking-system/internal/kafka"
	"hotel-booking-system/internal/notification"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	go func() {
		addr := ":8090"
		mux := http.NewServeMux()
		mux.HandleFunc("/html_email", notification.HTMLTemplateEmailHandler)
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		logrus.Infof("Starting notification HTTP server on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			logrus.Errorf("HTTP server failed: %v", err)
		}
	}()

	kafkaAddr := os.Getenv("KAFKA_BROKERS")
	if kafkaAddr == "" {
		kafkaAddr = "localhost:9091,localhost:9092,localhost:9093"
	}
	brokers := strings.Split(kafkaAddr, ",")
	topic := "booking-created"
	groupID := "notification-service-group"

	notificationHandler := handler.NewHandler()

	for i := 1; i <= 3; i++ {
		consumer, err := kafka.NewConsumer(notificationHandler, brokers, topic, groupID, i)
		if err != nil {
			logrus.Fatalf("Failed to create consumer %d: %v", i, err)
		}
		go consumer.Start()
	}

	logrus.Info("Notification service started (Kafka consumers + HTTP)")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logrus.Info("Shutting down notification service...")
}
