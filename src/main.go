package main

import (
	"app/api"
	"app/config"
	"app/cron"
	"app/infrastructure/postgres"
	"app/infrastructure/repository"
	"app/kafka"
	usecase_user "app/usecase/user"
	"log"

	_ "time/tzdata" // Required for tzdata to work
)

func main() {

	config.ReadEnvironmentVars()

	cron.StartCronJobs()

	conn := postgres.Connect()
	postgres.Migrations()

	usecase := usecase_user.NewService(
		repository.NewUserPostgres(conn),
	)

	err := usecase.CreateAdminUser()
	if err != nil {
		log.Println("---------->     Error creating admin user     <----------")
		log.Println(err)
	}

	go kafka.StartKafka()

	api.StartWebServer()
}
