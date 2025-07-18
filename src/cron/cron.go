package cron

import (
	"time"

	"github.com/go-co-op/gocron"
)

func StartCronJobs() {
	s := gocron.NewScheduler(time.UTC)

	s.StartAsync()
}
