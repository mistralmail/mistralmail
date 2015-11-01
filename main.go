package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gopistolet/mta"
)

type Chain struct {
	handlers []mta.Handler
}

func (c *Chain) HandleMail(state *mta.State) {
	for _, handler := range c.handlers {
		handler.HandleMail(state)
	}
}

func main() {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	log.Println("GoPistolet at your service!")

	// Default config
	c := mta.Config{
		Hostname: "localhost",
		Port:     25,
		TlsCert:  "ssl/server.crt",
		TlsKey:   "ssl/server.key",
	}

	// Load config from JSON file
	err := helpers.DecodeFile("config.json", &c)
	if err != nil {
		log.Warnln(err, "- Using default configuration instead.")
	}

	mta := mta.NewDefault(c,
		&Chain{handlers: []mta.Handler{
			mta.HandlerFunc(mail),
			mta.HandlerFunc(handleSPF),
			mta.HandlerFunc(handleMailDir),
		}})
	go func() {
		<-sigc
		mta.Stop()
	}()
	err = mta.ListenAndServe()
	if err != nil {
		log.Errorln(err)
	}
}
