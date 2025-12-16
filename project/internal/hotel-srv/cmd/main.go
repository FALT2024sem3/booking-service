package main

import (
	"context"
	"fmt"
	"hotel-booking-system/internal/database"
	"hotel-booking-system/internal/hotel-srv/server"
	"hotel-booking-system/internal/hotel-srv/stg"
	hotelv1 "hotel-booking-system/package/proto/fast/stable"

	// "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Загрузка конфигурации из YAML файла
func LoadConfig(filename string) (*database.DBConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config struct {
		Databases struct {
			Hotel database.DBConfig `yaml:"hotel"`
		} `yaml:"databases"`
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config.Databases.Hotel, nil
}

func main() {
	// Настройка логирования
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../../../.env")
	}

	// Загрузка конфигурации
	configPath := "configs/database.yaml"
	config, err := LoadConfig(configPath)
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}

	// Подключение к базе данных
	hotelDB, err := database.ConnectHotelDB(*config)
	if err != nil {
		logrus.Fatalf("Error connecting to hotel database: %v", err)
		os.Exit(1)
	}
	defer hotelDB.Close()

	// Создание storage
	local_storage := stg.NewStorage(hotelDB)

	// Создание сервера
	hotelServer := server.NewHotelServer(local_storage)
	hotelServer.SetServer()

	// Запуск HTTP сервера
	httpServer := &http.Server{
		Addr:    ":8081",
		Handler: hotelServer.Mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logrus.Info("Starting hotel HTTP server on http://localhost:8081")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Error starting the server: %v", err)
		}
	}()

	// Запуск gRPC сервера для внутреннего взаимодействия с booking-srv
	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logrus.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	grpcServer := grpc.NewServer()
	hotelv1.RegisterHotelServiceServer(grpcServer, hotelServer)

	go func() {
		logrus.Info("Starting hotel gRPC server on :50051")
		if err := grpcServer.Serve(grpcListener); err != nil {
			logrus.Errorf("gRPC server error: %v", err)
		}
	}()

	<-stop
	logrus.Info("\nReceived shutdown signal...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Остановка gRPC сервера
	grpcServer.GracefulStop()

	// Остановка HTTP сервера
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("Error during server shutdown: %v", err)
	} else {
		logrus.Info("Server gracefully stopped")
	}
}
