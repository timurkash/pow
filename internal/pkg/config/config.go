package config

import (
	"encoding/json"
	"github.com/kelseyhightower/envconfig"
	"os"
)

type Config struct {
	ServerHost            string `envconfig:"SERVER_HOST"`
	ServerPort            int    `envconfig:"SERVER_PORT"`
	CacheHost             string `envconfig:"CACHE_HOST"`
	CachePort             int    `envconfig:"CACHE_PORT"`
	HashCashZerosCount    int
	HashCashDuration      int64
	HashCashMaxIterations int
}

func Load(path string) (*Config, error) {
	config := Config{}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return &config, err
	}
	err = envconfig.Process("", &config)
	return &config, err
}
