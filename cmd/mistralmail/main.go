package main

import (
	"github.com/mistralmail/mistralmail"
	log "github.com/sirupsen/logrus"
)

func main() {

	config, err := mistralmail.BuildConfigFromEnv()
	if err != nil {
		log.Fatalf("couldn't build config: %v", err)
	}

	err = config.Validate()
	if err != nil {
		log.Fatalf("config invalid: %v", err)
	}

	log.Printf("starting MistralMail with config: %+v", config)

	mistralmail.Serve(config)

}
