package config

import (
	"os"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Server       ServerConfig
	DB           DBConfig
	Kafka        KafkaConfig
	AI           AIConfig
	Notification NotificationConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port string
	Env  string
}

// DBConfig holds database connection settings.
type DBConfig struct {
	DSN string
}

// KafkaConfig holds Kafka broker and topic settings.
type KafkaConfig struct {
	Brokers       []string
	EventsTopic   string
	AnalysisTopic string
}

// AIConfig holds AI provider credentials and routing.
type AIConfig struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

// NotificationConfig holds all notification channel settings.
type NotificationConfig struct {
	GoogleChatWebhookURL string
	SMTPHost             string
	SMTPPort             string
	SMTPFrom             string
	WebhookURL           string
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() Config {
	brokerStr := getEnv("KAFKA_BROKERS", "localhost:9092")
	brokers := strings.Split(brokerStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	return Config{
		Server: ServerConfig{
			Port: getEnv("CUSTOS_PORT", "8080"),
			Env:  getEnv("CUSTOS_ENV", "development"),
		},
		DB: DBConfig{
			DSN: getEnv("DATABASE_URL", ""),
		},
		Kafka: KafkaConfig{
			Brokers:       brokers,
			EventsTopic:   getEnv("KAFKA_EVENTS_TOPIC", "custos.events"),
			AnalysisTopic: getEnv("KAFKA_ANALYSIS_TOPIC", "custos.analysis"),
		},
		AI: AIConfig{
			Provider: getEnv("CUSTOS_AI_PROVIDER", ""),
			APIKey:   getEnv("CUSTOS_AI_API_KEY", ""),
			Model:    getEnv("CUSTOS_AI_MODEL", ""),
			BaseURL:  getEnv("CUSTOS_AI_BASE_URL", ""),
		},
		Notification: NotificationConfig{
			GoogleChatWebhookURL: getEnv("GOOGLE_CHAT_WEBHOOK_URL", ""),
			SMTPHost:             getEnv("SMTP_HOST", ""),
			SMTPPort:             getEnv("SMTP_PORT", "587"),
			SMTPFrom:             getEnv("SMTP_FROM", ""),
			WebhookURL:           getEnv("WEBHOOK_URL", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
