package postgres

import (
	"app/config"
	"app/entity"
	custom_logger "app/pkg/logger"
	"fmt"

	// "gorm.io/driver/postgres"
	postgres "go.elastic.co/apm/module/apmgormv2/driver/postgres"
	"gorm.io/gorm"
)

var gormDB *gorm.DB

func Connect() *gorm.DB {

	if gormDB == nil {
		return conn()
	}

	return gormDB
}

func Migrations() {
	db := Connect()

	db.AutoMigrate(&entity.EntityUser{})
}

func conn() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.EnvironmentVariables.POSTGRES_HOST,
		config.EnvironmentVariables.POSTGRES_USER,
		config.EnvironmentVariables.POSTGRES_PASSWORD,
		config.EnvironmentVariables.POSTGRES_DB,
		config.EnvironmentVariables.POSTGRES_PORT,
	)

	// Configurar logger do GORM baseado na variável de ambiente
	gormLogger := custom_logger.NewGormLogger(config.EnvironmentVariables.GormLogLevel)
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		panic(err)
	}

	gormDB = conn

	return gormDB
}
