package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress  string
	DBAddress      string
	AccrualAddress string
	SecretKey      string
}

func InitConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.DBAddress, "d", "", "database connection address")
	flag.StringVar(&config.AccrualAddress, "r", "", "address of the accrual calculation system")
	flag.StringVar(&config.SecretKey, "k", "secretKey", "key for hashing the data")

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

	return &config
}
