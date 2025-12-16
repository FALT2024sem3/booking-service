package config

import (
    "fmt"
    "os"
    "hotel-booking-system/internal/database"
    
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server struct {
        HTTPPort string `yaml:"http_port"`
        GRPCPort string `yaml:"grpc_port"`
    } `yaml:"hotel_service"`
    Database database.DBConfig `yaml:"-"`
}

func LoadConfig(filename string, dbConfig database.DBConfig) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config struct {
        HotelService struct {
            HTTPPort string `yaml:"http_port"`
            GRPCPort string `yaml:"grpc_port"`
        } `yaml:"hotel_service"`
    }
    
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }
    
    cfg := &Config{
        Server: config.HotelService,
        Database: dbConfig,
    }
    
    return cfg, nil
}