package config

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Path to config file")
}

type Config struct {
	EnvLevel    uint8      `yaml:"env_level" env-default:"1"`
	StoragePath string     `yaml:"storage_path" env-required:"true"`
	GRPC        GRPCConfig `yaml:"grpc"`
	HTTP        HttpConfig `yaml:"http"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"44044"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type HttpConfig struct {
	UrlShortener UrlShortenerConfig `yaml:"url_shortener"`
}

type UrlShortenerConfig struct {
	MaxRetires int           `yaml:"max_retries" env-default:"3"`
	BaseUrl    string        `yaml:"base_url" env-required:"true"`
	Timeout    time.Duration `yaml:"timeout" env-default:"5s"`
}

func MustLoadByString(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file not found")
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}

func MustLoad() *Config {
	if err := fetchConfigPath(); err != nil {
		panic(err)
	}

	return MustLoadByString(configPath)
}

func fetchConfigPath() error {
	// --config="path/to/config.yaml"

	flag.Parse()
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath == "" {
			return errors.New("no config path provided")
		}
	}

	return nil
}
