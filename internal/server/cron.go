package server

import (
	"time"

	"github.com/robfig/cron/v3"
)

func NewCronWithLocation(location *time.Location) *cron.Cron {
	return cron.New(cron.WithLocation(location))
}
