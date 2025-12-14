package main

import (
	"context"
	"fmt"
	"hotel-booking-system/internal/booking-srv/server"
	"hotel-booking-system/internal/booking-srv/stg"
	"hotel-booking-system/internal/database"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

// Загрузка конфигурации из YAML файла
func LoadConfig(filename string) (*database.DBConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config database.DBConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

func main() {
	config, err := LoadConfig("../../../configs/database.yaml")

	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	booking_bd, err := database.ConnectBookingDB(*config)

	if err != nil {
		fmt.Printf("Error loading booking_database: %v\n", err)
		os.Exit(1)
	}

	local_storage := stg.NewStorage(booking_bd)
	bookingServer := server.NewBookingServer(local_storage)
	bookingServer.SetServer()
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: bookingServer.Mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Starting server on http://localhost:8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting the server: %v\n", err)
		}
	}()

	<-stop
	fmt.Println("\nReceived shutdown signal...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("Error during server shutdown: %v\n", err)
	} else {
		fmt.Println("Server gracefully stopped")
	}
}
