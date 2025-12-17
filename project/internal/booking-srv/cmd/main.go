package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"hotel-booking-system/internal/booking-srv/repository"
	"hotel-booking-system/internal/booking-srv/server"
	"hotel-booking-system/internal/booking-srv/stg"
	"hotel-booking-system/internal/kafka"
	db "hotel-booking-system/internal/package/database"
	hotelv1 "hotel-booking-system/package/proto/fast/stable"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, relying on environment variables")
	}

	hotelAddr := os.Getenv("HOTEL_SERVICE_ADDR")
	if hotelAddr == "" {
		hotelAddr = "localhost:50051"
	}
	conn, err := grpc.NewClient(hotelAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("Did not connect to hotel service: %v", err)
	}
	defer conn.Close()
	hotelClient := hotelv1.NewHotelServiceClient(conn)

	dbConfig := db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}
	bookingDB, err := db.Connect(dbConfig)
	if err != nil {
		logrus.Fatalf("Error connecting to booking database: %v", err)
	}
	defer bookingDB.Close()

	kafkaAddr := os.Getenv("KAFKA_BROKERS")
	if kafkaAddr == "" {
		kafkaAddr = "localhost:9091,localhost:9092,localhost:9093"
	}
	producer, err := kafka.NewProducer(strings.Split(kafkaAddr, ","))
	if err != nil {
		logrus.Fatalf("Failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	repo := repository.NewRepository(bookingDB)

	storage := stg.NewStorage(repo, hotelClient, producer)

	bookingServer := server.NewBookingServer(storage)
	bookingServer.SetServer()

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = ":8080"
	}

	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: bookingServer.Mux,
	}

	go func() {
		logrus.Infof("Starting booking server on %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logrus.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("Error during server shutdown: %v", err)
	} else {
		logrus.Info("Server gracefully stopped")
	}
}
