package gopistolet

import (
	"os"
	"os/signal"
	"syscall"

	imapbackend "github.com/gopistolet/gopistolet/backend/imap"
	smtpbackend "github.com/gopistolet/gopistolet/backend/smtp"
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

	// Create backends
	db, err := imap.InitDB(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Couldn't connect to database: %v", err)
	}

	imapBackend, err := imapbackend.NewIMAPBackend(db)
	if err != nil {
		log.Fatalf("Couldn't create IMAP backend: %v", err)
	}

	smtpBackend, err := smtpbackend.NewSMTPBackend(db)
	if err != nil {
		log.Fatalf("Couldn't create SMTP backend: %v", err)
	}

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
		msa.Server.AuthBackend = smtpBackend

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
		imap.Serve(config.GenerateIMAPBackendConfig(), imapBackend)
	}()

	// Run SMTP MTA
	mtaConfig := config.GenerateMTAConfig()

	mtaHandlerChain := &handlers.HandlerMachanism{
		Handlers: []handlers.Handler{
			received.New(mtaConfig),
			spf.New(mtaConfig),
			NewIMAPHandler(mtaConfig, imapBackend),
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

	// TODO close database
}
