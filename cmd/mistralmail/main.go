package main

import (
	"github.com/mistralmail/mistralmail"
	log "github.com/sirupsen/logrus"
)

func main() {

	config := mistralmail.BuildConfigFromEnv()

	err := config.Validate()
	if err != nil {
		log.Fatalf("config invalid: %v", err)
	}

	log.Printf("starting MistralMail with config: %+v", config)

	mistralmail.Serve(config)

}
