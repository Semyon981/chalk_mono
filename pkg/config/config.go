package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RedisURI    string     `yaml:"redis_uri"`
	PostgresURI string     `yaml:"postgres_uri"`
	Auth        AuthConfig `yaml:"auth"`
	HTTPPort    string     `yaml:"http_port"`
}

type AuthConfig struct {
	EmailFromAddr   string        `yaml:"email_from_addr"`
	EmailFromName   string        `yaml:"email_from_name"`
	CodeTTL         time.Duration `yaml:"code_ttl"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
	Mailer          MailerConfig  `yaml:"mailer"`
}

type MailerConfig struct {
	SmtpHost         string `yaml:"smtp_host"`
	SmtpPort         string `yaml:"smtp_port"`
	SmtpSASLUsername string `yaml:"smtp_sasl_username"`
	SmtpSASLPassword string `yaml:"smtp_sasl_password"`
}

func LoadConfig(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return cfg
}
