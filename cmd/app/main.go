package main

import (
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/app/di"
)

func main() {
	conf := config.Load()
	app, _ := di.InitializeApiWebServer(conf)
	app.Serve()
}
