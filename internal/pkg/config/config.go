package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"database"`
	Redis      Redis      `yaml:"redis"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-default:"0.0.0.0:8080"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

type Database struct {
	Host     string `yaml:"host" env:"DB_HOST" env-required:"true"`
	Port     int    `yaml:"port" env:"DB_PORT" env-required:"true"`
	User     string `yaml:"user" env:"DB_USER" env-required:"true"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
	DBName   string `yaml:"db_name" env:"DB_NAME" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSLMODE" env-default:"disable"`
}

type Redis struct {
	Host         string        `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port         int           `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	Password     string        `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	DB           int           `yaml:"db" env:"REDIS_DB" env-default:"0"`
	TTL          time.Duration `yaml:"ttl" env:"REDIS_TTL" env-default:"5m"`
	DialTimeout  time.Duration `yaml:"dial_timeout" env:"REDIS_DIAL_TIMEOUT" env-default:"1s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"REDIS_READ_TIMEOUT" env-default:"200ms"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"REDIS_WRITE_TIMEOUT" env-default:"200ms"`
}

func (d Database) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

func (r Redis) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, relying on system env")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("can not read the config: %s", err.Error())
	}
	return &cfg
}
