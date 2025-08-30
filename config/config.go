package config

import "os"

type (
	Config struct {
		GRPC
	}

	GRPC struct {
		Port        string `env:"GRPC_PORT"`
		GatewayPort string `env:"GRPC_GATEWAY_PORT"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	cfg.GRPC.Port = os.Getenv("GRPC_PORT")
	if cfg.GRPC.Port == "" {
		cfg.GRPC.Port = "9090"
	}

	cfg.GRPC.GatewayPort = os.Getenv("GRPC_GATEWAY_PORT")
	if cfg.GRPC.GatewayPort == "" {
		cfg.GRPC.GatewayPort = "8080"
	}

	return cfg, nil
}
