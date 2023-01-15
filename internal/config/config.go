package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	ClientId     string `env:"CLIENT_ID,required"`
	ClientSecret string `env:"CLIENT_SECRET,required"`
	HostURL      string `env:"HOST_URL,required"`
	Addr         string `env:"ADDR,required"`
	SessionKey   string `env:"SESSION_KEY,required"`
	LogFolder    string `env:"LOG_FOLDER,required"`
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	err = env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
