package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPServer `yaml:"http_server"`
	Dsn        `yaml:"database"`
	Settings   `yaml:"settings"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env-default:"localhost:8081"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"10s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Settings struct {
	LengthShortUrl int    `yaml:"length_hort_url" env-default:"10"`
	StorageType    string `yaml:"storage_type" env-default:"memory" env:"STORAGE_TYPE"`
}

type Dsn struct {
	DSN string `yaml:"dsn" env:"DB_DSN" env-default:"postgresql://postgres:password@localhost:5432/OzonShortener?sslmode=disable"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		configPath = "configs/local.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	return &cfg
}
