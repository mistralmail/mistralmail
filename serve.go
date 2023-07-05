package gopistolet

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/handlers"
	"github.com/gopistolet/gopistolet/handlers/received"
	"github.com/gopistolet/gopistolet/handlers/relay"
	"github.com/gopistolet/gopistolet/handlers/spf"
	imapbackend "github.com/gopistolet/imap-backend"
	"github.com/gopistolet/smtp/server"
	log "github.com/sirupsen/logrus"
)

// Serve runs GoPistolet
func Serve(config *Config) {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	log.Println("GoPistolet at your service!")

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
		// TODO: use real auth backend
		msa.Server.AuthBackend = server.NewAuthBackendMemory(map[string]string{"username@example.com": "password"})

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
		imapbackend.Serve(config.GenerateIMAPBackendConfig())
	}()

	// Run SMTP MTA
	mtaConfig := config.GenerateMTAConfig()

	mtaHandlerChain := &handlers.HandlerMachanism{
		Handlers: []handlers.Handler{
			received.New(mtaConfig),
			spf.New(mtaConfig),
			NewIMAPHandler(mtaConfig),
		},
	}

	mta := server.NewDefault(*mtaConfig, mtaHandlerChain)
	go func() {
		<-sigc
		mta.Stop()
	}()
	err := mta.ListenAndServe()
	if err != nil {
		log.Errorln(err)
	}
}
