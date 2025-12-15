package main

import (
	"context"
	"fmt"
	"hotel-booking-system/internal/booking-srv/server"
	"hotel-booking-system/internal/booking-srv/stg"
	"hotel-booking-system/internal/database"
	"hotel-booking-system/internal/handler"
	"hotel-booking-system/internal/kafka"
	"hotel-booking-system/internal/notification"
	hotelv1 "hotel-booking-system/package/proto/fast/stable"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	topic         = "my-topic"
	consumerGroup = "my-consumer-group"
)

var address = []string{"localhost:9091", "localhost:9092", "localhost:9093"}

// Загрузка конфигурации из YAML файла
func LoadConfig(filename string) (*database.DBConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config struct {
		Databases struct {
			Booking database.DBConfig `yaml:"booking"`
		} `yaml:"databases"`
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config.Databases.Booking, nil
}

func main() {

	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../../../.env")
	}

	configPath := "configs/database.yaml"
	config, err := LoadConfig(configPath)

	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to hotel service: %v", err)
	}
	defer conn.Close()

	hotelClient := hotelv1.NewHotelServiceClient(conn)

	booking_bd, err := database.ConnectBookingDB(*config)

	if err != nil {
		fmt.Printf("Error loading booking_database: %v\n", err)
		os.Exit(1)
	}

	local_storage := stg.NewStorage(booking_bd, hotelClient)
	bookingServer := server.NewBookingServer(local_storage)
	bookingServer.SetServer()
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: bookingServer.Mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Starting server_1 on http://localhost:8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting the server: %v\n", err)
		}
	}()

	go func() {
		addr := ":8090"
		mux := http.NewServeMux()
		mux.HandleFunc("/html_email", notification.HTMLTemplateEmailHandler)

		log.Printf("Starting server_2 on http://localhost %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("Error starting notification server: %v", err)
		}
	}()

	notificationHandler := handler.NewHandler()
	consumer1, err := kafka.NewConsumer(notificationHandler, address, topic, consumerGroup, 1)
	if err != nil {
		logrus.Fatal(err)
	}
	consumer2, err := kafka.NewConsumer(notificationHandler, address, topic, consumerGroup, 2)
	if err != nil {
		logrus.Fatal(err)
	}
	consumer3, err := kafka.NewConsumer(notificationHandler, address, topic, consumerGroup, 3)
	if err != nil {
		logrus.Fatal(err)
	}

	go consumer1.Start()
	go consumer2.Start()
	go consumer3.Start()

	<-stop
	fmt.Println("\nReceived shutdown signal...")

	// Stop consumers
	_ = consumer1.Stop()
	_ = consumer2.Stop()
	_ = consumer3.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("Error during server shutdown: %v\n", err)
	} else {
		fmt.Println("Server gracefully stopped")
	}
}
