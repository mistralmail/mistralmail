package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/handlers"
	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/smtp/mta"
)

var c mta.Config

func main() {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	nixspamBlacklist, err := helpers.NewNixspam()
	if err != nil {
		log.Warnln("Couldn't create Nixspam Blacklist instance: ", err)
	}

	log.Println("GoPistolet at your service!")

	// Default config
	c = mta.Config{
		Hostname:  "localhost",
		Port:      25,
		Blacklist: nixspamBlacklist,
	}

	// Load config from JSON file
	err = helpers.DecodeFile("config.json", &c)
	if err != nil {
		log.Warnln(err, "- Using default configuration instead.")
	}

	mta := mta.NewDefault(c, handlers.LoadHandlers(&c))
	go func() {
		<-sigc
		mta.Stop()
	}()
	err = mta.ListenAndServe()
	if err != nil {
		log.Errorln(err)
	}
}
