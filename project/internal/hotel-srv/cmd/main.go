package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"hotel-booking-system/internal/hotel-srv/repository"
	"hotel-booking-system/internal/hotel-srv/server"
	"hotel-booking-system/internal/hotel-srv/stg"
	db "hotel-booking-system/internal/package/database"
	hotelv1 "hotel-booking-system/package/proto/fast/stable"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	dbConfig := db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}

	hotelDB, err := db.Connect(dbConfig)
	if err != nil {
		logrus.Fatalf("Error connecting to hotel database: %v", err)
	}
	defer hotelDB.Close()
	repo := repository.NewRepository(hotelDB)
	storage := stg.NewStorage(repo)
	hotelServer := server.NewHotelServer(storage)
	hotelServer.SetServer()
	go func() {
		port := os.Getenv("HTTP_PORT")
		if port == "" {
			port = ":8081"
		}
		logrus.Infof("Starting hotel HTTP server on %s", port)
		if err := http.ListenAndServe(port, hotelServer.Mux); err != nil {
			logrus.Errorf("HTTP server error: %v", err)
		}
	}()
	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logrus.Fatalf("Failed to listen on gRPC: %v", err)
	}
	grpcServer := grpc.NewServer()
	hotelv1.RegisterHotelServiceServer(grpcServer, hotelServer)
	go func() {
		logrus.Info("Starting hotel gRPC server on :50051")
		if err := grpcServer.Serve(grpcListener); err != nil {
			logrus.Errorf("gRPC server error: %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logrus.Info("Shutting down...")
	grpcServer.GracefulStop()
}
