package config

import (
	"os"
	"strconv"
	"strings"
)

var EnvironmentVariables EnvironmentVars

func ReadEnvironmentVars() {

	EnvironmentVariables.ISRELEASE = os.Getenv("IS_RELEASE") == "true"

	// Logging configuration
	LogLevel := getEnvOrDefault("LOG_LEVEL", "INFO")
	GinMode := getEnvOrDefault("GIN_MODE", "release")
	GormLogLevel := getEnvOrDefault("GORM_LOG_LEVEL", "WARN")

	EnvironmentVariables.LogLevel = strings.ToUpper(LogLevel)
	EnvironmentVariables.GinMode = strings.ToLower(GinMode)
	EnvironmentVariables.GormLogLevel = strings.ToUpper(GormLogLevel)

	// Read environment variables
	EnvironmentVariables.POSTGRES_DB = os.Getenv("POSTGRES_DB")
	EnvironmentVariables.POSTGRES_USER = os.Getenv("POSTGRES_USER")
	EnvironmentVariables.POSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD")
	EnvironmentVariables.POSTGRES_HOST = os.Getenv("POSTGRES_HOST")
	EnvironmentVariables.POSTGRES_PORT, _ = strconv.Atoi(os.Getenv("POSTGRES_PORT"))

	EnvironmentVariables.KAFKA_BOOTSTRAP_SERVER = os.Getenv("KAFKA_BOOTSTRAP_SERVER")
	EnvironmentVariables.KAFKA_CLIENT_ID = os.Getenv("KAFKA_CLIENT_ID")
	EnvironmentVariables.KAFKA_GROUP_ID = os.Getenv("KAFKA_GROUP_ID")

	EnvironmentVariables.EMAIL_HOST = os.Getenv("EMAIL_HOST")
	EnvironmentVariables.EMAIL_HOST_USER = os.Getenv("EMAIL_HOST_USER")
	EnvironmentVariables.EMAIL_HOST_PASSWORD = os.Getenv("EMAIL_HOST_PASSWORD")
	EnvironmentVariables.EMAIL_PORT, _ = strconv.Atoi(os.Getenv("EMAIL_PORT"))

	EnvironmentVariables.EMAIL_FROM = os.Getenv("EMAIL_FROM")

	EnvironmentVariables.DEFAULT_ADMIN_MAIL = os.Getenv("DEFAULT_ADMIN_MAIL")
	EnvironmentVariables.DEFAULT_ADMIN_PASSWORD = os.Getenv("DEFAULT_ADMIN_PASSWORD")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
