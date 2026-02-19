package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddress  string
	DBAddress      string
	AccrualAddress string
	SecretKey      string
}

const (
	defaultServerAddress  = "localhost:8080"
	defaultDBAddress      = ""
	defaultAccrualAddress = ""
	defaultSecretKey      = "secretKey"
)

func InitConfig() (*Config, error) {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", defaultServerAddress, "address and port to run server")
	flag.StringVar(&config.DBAddress, "d", defaultDBAddress, "database connection address")
	flag.StringVar(&config.AccrualAddress, "r", defaultAccrualAddress, "address of the accrual calculation system")
	flag.StringVar(&config.SecretKey, "k", defaultSecretKey, "key for hashing the data")

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

	if config.DBAddress == "" {
		return nil, fmt.Errorf("the database address is not specified")
	} else if config.AccrualAddress == "" {
		return nil, fmt.Errorf("the address accrual calculation system is not specified")
	}

	return &config, nil
}
