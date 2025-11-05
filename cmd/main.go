package main

import (
	"img_cache_control/app"
	"log"
)

func main() {

	a, err := app.NewApp()
	if err != nil {
		log.Fatalf("failed to init App: %v", err)
	}

	updater := &app.CacheControlUpdater{App: a}

	err = updater.Run()
	if err != nil {
		log.Fatalf("Failed to update images: %v", err)
	}

}
