package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	ServerAddress  string
	DBAddress      string
	AccrualAddress string
	SecretKey      string
	LogLevel       string
}

const (
	defaultServerAddress  = "localhost:8080"
	defaultDBAddress      = ""
	defaultAccrualAddress = ""
	defaultSecretKey      = "secretKey"
	defaultLogLevel       = "INFO"
)

func ParseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func InitConfig() (*Config, error) {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", defaultServerAddress, "address and port to run server")
	flag.StringVar(&config.DBAddress, "d", defaultDBAddress, "database connection address")
	flag.StringVar(&config.AccrualAddress, "r", defaultAccrualAddress, "address of the accrual calculation system")
	flag.StringVar(&config.SecretKey, "k", defaultSecretKey, "key for hashing the data")
	flag.StringVar(&config.LogLevel, "l", defaultLogLevel, "the level of output logs")

	flag.Parse()

	if envAddr := os.Getenv("RUN_ADDRESS"); envAddr != "" {
		config.ServerAddress = envAddr
	}
	if envAddrDB := os.Getenv("DATABASE_URI"); envAddrDB != "" {
		config.DBAddress = envAddrDB
	}
	if envAddrAccrual := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAddrAccrual != "" {
		config.AccrualAddress = envAddrAccrual
	}
	if envSecretKey := os.Getenv("SECRET_KEY"); envSecretKey != "" {
		config.SecretKey = envSecretKey
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		config.LogLevel = envLogLevel
	}

	if config.DBAddress == "" {
		return nil, fmt.Errorf("the database address is not specified")
	} else if config.AccrualAddress == "" {
		return nil, fmt.Errorf("the address accrual calculation system is not specified")
	}

	return &config, nil
}
