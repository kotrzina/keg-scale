package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Debug bool

	RedisAddr string
	RedisDB   int

	AuthToken string // used for communication with the scale
	Password  string // shared admin password

	FrontendPath string

	PrometheusURL      string
	PrometheusUser     string
	PrometheusPassword string
	PrometheusOrg      string

	DBString string

	WhatsAppOpenJid string
	AnthropicAPIKey string
	OpenAiAPIKey    string

	CalendarPubURL      string
	CalendarConcertsURL string
}

func NewConfig() *Config {
	return &Config{
		Debug: getBoolEnvDefault("DEBUG", false),

		RedisAddr: getStringEnvDefault("REDIS_ADDR", "localhost:6379"),
		RedisDB:   getIntEnvDefault("REDIS_DB", 0),

		AuthToken: getStringEnvDefault("AUTH_TOKEN", "test"),
		Password:  getStringEnvDefault("PASSWORD", "test"),

		FrontendPath: getStringEnvDefault("FRONTEND_PATH", "./../frontend/build/"),

		PrometheusURL:      getStringEnvDefault("PROMETHEUS_URL", "http://localhost:9090"),
		PrometheusUser:     getStringEnvDefault("PROMETHEUS_USER", "test"),
		PrometheusPassword: getStringEnvDefault("PROMETHEUS_PASSWORD", "test"),
		PrometheusOrg:      getStringEnvDefault("PROMETHEUS_ORG", "test"),

		DBString: getStringEnvDefault("DB_STRING", "host=localhost port=5432 user=postgres password=admin dbname=pub sslmode=disable"),

		WhatsAppOpenJid: getStringEnvDefault("WHATSAPP_OPEN_JID", ""),
		AnthropicAPIKey: getStringEnvDefault("ANTHROPIC_API_KEY", ""),
		OpenAiAPIKey:    getStringEnvDefault("OPENAI_API_KEY", ""),

		CalendarPubURL:      getStringEnvDefault("CALENDAR_PUB_URL", ""),
		CalendarConcertsURL: getStringEnvDefault("CALENDAR_CONCERTS_URL", ""),
	}
}

func getBoolEnvDefault(key string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}

	fmt.Printf("Using default value for %s\n", key)
	return defaultValue
}

func getStringEnvDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Printf("Using default value for %s\n", key)
	return defaultValue
}

func getIntEnvDefault(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	fmt.Printf("Using default value for %s\n", key)
	return defaultValue
}
