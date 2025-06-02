package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CustomMessage represents a option to send custom opening messages directly to the WhatsApp users
type CustomMessage struct {
	Name  string
	Phone string
}

// BotkaCommands represents secret commands that can be used in the WhatsApp chat
// they starts with ! and the body of message needs to be exactly the same as the one defined here
type BotkaCommands struct {
	Help       string
	Volleyball string
	NoMessage  string
}

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

	WhatsAppOpenJid        string
	WhatsAppCustomMessages []CustomMessage
	AnthropicAPIKey        string
	OpenAiAPIKey           string

	FioToken string

	Commands BotkaCommands

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

		WhatsAppOpenJid:        getStringEnvDefault("WHATSAPP_OPEN_JID", ""),
		WhatsAppCustomMessages: parseCustomMessages(getStringEnvDefault("WHATSAPP_CUSTOM_MESSAGES", "")),
		AnthropicAPIKey:        getStringEnvDefault("ANTHROPIC_API_KEY", ""),
		OpenAiAPIKey:           getStringEnvDefault("OPENAI_API_KEY", ""),

		FioToken: getStringEnvDefault("FIO_TOKEN", ""),

		Commands: parseBotkaCommands(os.Getenv("BOTKA_COMMANDS")),

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

func parseCustomMessages(envString string) []CustomMessage {
	messages := strings.Split(envString, ",")

	customMessages := make([]CustomMessage, 0, len(messages))
	for _, message := range messages {
		parts := strings.Split(message, ":")
		if len(parts) != 2 {
			fmt.Printf("Invalid custom message: %s\n", message)
			continue
		}
		customMessages = append(customMessages, CustomMessage{
			Name:  parts[0],
			Phone: parts[1],
		})
	}
	return customMessages
}

func parseBotkaCommands(input string) BotkaCommands {
	rawCommands := strings.Split(input, ",")
	commands := make(map[string]string, len(rawCommands))

	for _, rawCommand := range rawCommands {
		command := strings.Split(rawCommand, ":")
		if len(command) != 2 {
			fmt.Printf("Invalid botka command: %s\n", rawCommand)
			continue
		}

		commands[command[0]] = command[1]
	}

	return BotkaCommands{
		Help:       commands["help"],
		Volleyball: commands["volleyball"],
		NoMessage:  commands["no_message"],
	}
}
