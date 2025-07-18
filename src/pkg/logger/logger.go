package logger

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

// LogLevel representa os níveis de log disponíveis
type LogLevel string

const (
	DEBUG  LogLevel = "DEBUG"
	INFO   LogLevel = "INFO"
	WARN   LogLevel = "WARN"
	ERROR  LogLevel = "ERROR"
	SILENT LogLevel = "SILENT"
)

// GetGormLogLevel converte string para nível de log do GORM
func GetGormLogLevel(level string) logger.LogLevel {
	switch LogLevel(level) {
	case SILENT:
		return logger.Silent
	case ERROR:
		return logger.Error
	case WARN:
		return logger.Warn
	case INFO:
		return logger.Info
	default:
		// Em produção, padrão é WARN para reduzir logs
		return logger.Warn
	}
}

// NewGormLogger cria um logger customizado para GORM
func NewGormLogger(level string) logger.Interface {
	logLevel := GetGormLogLevel(level)

	// Se for SILENT, retorna o logger silencioso
	if logLevel == logger.Silent {
		return logger.Default.LogMode(logger.Silent)
	}

	// Cria um logger customizado
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logLevel,    // Log level baseado na variável de ambiente
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  false,       // Disable color for production
		},
	)

	return newLogger
}

// IsDebugMode verifica se está em modo debug
func IsDebugMode(logLevel string) bool {
	return LogLevel(logLevel) == DEBUG
}

// ShouldLogLevel verifica se deve logar para o nível especificado
func ShouldLogLevel(currentLevel, targetLevel string) bool {
	levels := map[LogLevel]int{
		SILENT: 0,
		ERROR:  1,
		WARN:   2,
		INFO:   3,
		DEBUG:  4,
	}

	current, exists1 := levels[LogLevel(currentLevel)]
	target, exists2 := levels[LogLevel(targetLevel)]

	if !exists1 || !exists2 {
		return false
	}

	return current >= target
}
