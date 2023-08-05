package gopistolet

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/backend"
	"github.com/gopistolet/gopistolet/handlers"
	"github.com/gopistolet/gopistolet/handlers/received"
	"github.com/gopistolet/gopistolet/handlers/relay"
	"github.com/gopistolet/gopistolet/handlers/spf"
	"github.com/gopistolet/imap"
	"github.com/gopistolet/smtp/server"
	log "github.com/sirupsen/logrus"
)

// Serve runs GoPistolet
func Serve(config *Config) {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	log.Println("GoPistolet at your service!")

	// Create backend
	backend, err := backend.New(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Couldn't create backend: %v", err)
	}
	defer func() {
		err := backend.Close()
		if err != nil {
			log.Errorf("Couldn't close backend: %v", err)
		}
	}()

	// Run SMTP MSA
	go func() {
		msaConfig := config.GenerateMSAConfig()

		msaHandlerChain := &handlers.HandlerMachanism{
			Handlers: []handlers.Handler{
				received.New(msaConfig),
				relay.New(config.ExternalRelayHostname, config.ExternalRelayPort, config.ExternalRelayUsername, config.ExternalRelayPassword, config.ExternalRelayInsecureSkipVerify),
			},
		}

		msa := server.NewDefault(*msaConfig, msaHandlerChain)
		msa.Server.AuthBackend = backend.SMTPBackend

		go func() {
			<-sigc
			msa.Stop()
		}()
		err := msa.ListenAndServe()
		if err != nil {
			log.Errorln(err)
		}
	}()

	// Run IMAP
	go func() {
		imap.Serve(config.GenerateIMAPBackendConfig(), backend.IMAPBackend)
	}()

	// Run SMTP MTA
	mtaConfig := config.GenerateMTAConfig()

	mtaHandlerChain := &handlers.HandlerMachanism{
		Handlers: []handlers.Handler{
			received.New(mtaConfig),
			spf.New(mtaConfig),
			NewIMAPHandler(mtaConfig, backend.IMAPBackend),
		},
	}

	mta := server.NewDefault(*mtaConfig, mtaHandlerChain)
	go func() {
		<-sigc
		mta.Stop()
	}()
	err = mta.ListenAndServe()
	if err != nil {
		log.Errorln(err)
	}

}
