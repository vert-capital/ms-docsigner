package config

type EnvironmentVars struct {
	LogLevel     string
	GinMode      string
	GormLogLevel string

	POSTGRES_DB       string
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_HOST     string
	POSTGRES_PORT     int

	KAFKA_BOOTSTRAP_SERVER string
	KAFKA_CLIENT_ID        string
	KAFKA_GROUP_ID         string

	EMAIL_HOST          string
	EMAIL_HOST_USER     string
	EMAIL_HOST_PASSWORD string
	EMAIL_PORT          int

	EMAIL_FROM string

	DEFAULT_ADMIN_MAIL     string
	DEFAULT_ADMIN_PASSWORD string

	ISRELEASE bool
}
