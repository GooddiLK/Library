package config

import (
	"fmt"
	"net"
	"os"
)

type (
	Config struct {
		GRPC
		PG
	}

	GRPC struct {
		Port        string `env:"GRPC_PORT"`
		GatewayPort string `env:"GRPC_GATEWAY_PORT"`
	}

	PG struct {
		URL      string
		Host     string `env:"POSTGRES_HOST"`
		Port     string `env:"POSTGRES_PORT"`
		DB       string `env:"POSTGRES_DB"`
		User     string `env:"POSTGRES_USER"`
		Password string `env:"POSTGRES_PASSWORD"`
		MaxConn  string `env:"POSTGRES_MAX_CONNECTION"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	cfg.GRPC.Port = os.Getenv("GRPC_PORT")
	if cfg.GRPC.Port == "" {
		cfg.GRPC.Port = "9090"
	}

	cfg.GRPC.GatewayPort = os.Getenv("GRPC_GATEWAY_PORT")
	if cfg.GRPC.GatewayPort == "" {
		cfg.GRPC.GatewayPort = "8080"
	}

	cfg.PG.Host = os.Getenv("POSTGRES_HOST")
	if cfg.PG.Host == "" {
		cfg.PG.Host = "localhost"
	}
	cfg.PG.Port = os.Getenv("POSTGRES_PORT")
	if cfg.PG.Port == "" {
		cfg.PG.Port = "5432"
	}

	cfg.PG.DB = os.Getenv("POSTGRES_DB")
	if cfg.PG.DB == "" {
		cfg.PG.DB = "library"
	}

	cfg.PG.User = os.Getenv("POSTGRES_USER")
	if cfg.PG.User == "" {
		cfg.PG.User = "user"
	}

	cfg.PG.Password = os.Getenv("POSTGRES_PASSWORD")
	if cfg.PG.Password == "" {
		cfg.PG.Password = "1234567"
	}

	cfg.PG.MaxConn = os.Getenv("POSTGRES_MAX_CONN")
	if cfg.PG.MaxConn == "" {
		cfg.PG.MaxConn = "10"
	}

	hostPort := net.JoinHostPort(cfg.PG.Host, cfg.PG.Port)
	cfg.PG.URL = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&pool_max_conns=%s",
		cfg.PG.User,
		cfg.PG.Password,
		hostPort,
		cfg.PG.DB,
		cfg.PG.MaxConn,
	)

	return cfg, nil
}
