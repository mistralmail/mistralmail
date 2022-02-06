package gopistolet

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/handlers"
	"github.com/gopistolet/gopistolet/handlers/received"
	"github.com/gopistolet/gopistolet/handlers/spf"
	imapbackend "github.com/gopistolet/imap-backend"
	"github.com/gopistolet/smtp/mta"
	log "github.com/sirupsen/logrus"
)

// Serve runs GoPistolet
func Serve(config *Config) {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	log.Println("GoPistolet at your service!")

	// Run IMAP
	go func() {
		imapbackend.Serve(config.GenerateIMAPBackendConfig())
	}()

	// Run SMTP
	mtaConfig := config.GenerateMTAConfig()

	handlerChain := &handlers.HandlerMachanism{
		Handlers: []handlers.Handler{
			received.New(mtaConfig),
			spf.New(mtaConfig),
			NewIMAPHandler(mtaConfig),
		},
	}

	mta := mta.NewDefault(*mtaConfig, handlerChain)
	go func() {
		<-sigc
		mta.Stop()
	}()
	err := mta.ListenAndServe()
	if err != nil {
		log.Errorln(err)
	}
}
