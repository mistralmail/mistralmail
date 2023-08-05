package main

import (
	"github.com/gopistolet/gopistolet"
	log "github.com/sirupsen/logrus"
)

func main() {

	config := gopistolet.BuildConfigFromEnv()

	err := config.Validate()
	if err != nil {
		log.Fatalf("config invalid: %v", err)
	}

	log.Printf("starting GoPistolet with config: %+v", config)

	gopistolet.Serve(config)

}
