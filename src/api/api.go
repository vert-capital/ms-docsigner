package api

import (
	"log"

	"app/api/handlers"
	"app/config"
	"app/infrastructure/postgres"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin"
	"gorm.io/gorm"

	_ "app/docs"

	custom_logger "app/pkg/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupDatabase() *gorm.DB {
	conn := postgres.Connect()
	return conn
}

func setupRouter(conn *gorm.DB) *gin.Engine {
	gin.SetMode(config.EnvironmentVariables.GinMode)

	r := gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("authorization")

	r.Use(apmgin.Middleware(r))
	r.Use(cors.New(corsConfig))

	// Configurar middleware de logging baseado no nível de log
	if config.EnvironmentVariables.GinMode == "debug" || custom_logger.ShouldLogLevel(config.EnvironmentVariables.LogLevel, "INFO") {
		r.Use(gin.Logger())
	}

	r.Use(gin.Recovery())

	handlers.MountSamplesHandlers(r)
	handlers.MountUsersHandlers(r, conn)

	return r
}

func SetupRouters() *gin.Engine {
	conn := setupDatabase()
	return setupRouter(conn)
}

func StartWebServer() {
	config.ReadEnvironmentVars()

	r := SetupRouters()

	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json") // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// se for release, reduz o log
	if config.EnvironmentVariables.ISRELEASE {
		gin.SetMode(gin.ReleaseMode)
	}

	// Bind to a port and pass our router in
	log.Fatal(r.Run())
}
