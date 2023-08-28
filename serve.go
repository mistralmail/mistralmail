package mistralmail

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mistralmail/imap"
	"github.com/mistralmail/mistralmail/api"
	"github.com/mistralmail/mistralmail/backend"
	"github.com/mistralmail/mistralmail/backend/services/certificates"
	"github.com/mistralmail/mistralmail/handlers"
	imaphandler "github.com/mistralmail/mistralmail/handlers/imap"
	"github.com/mistralmail/mistralmail/handlers/received"
	"github.com/mistralmail/mistralmail/handlers/relay"
	"github.com/mistralmail/mistralmail/handlers/spf"
	"github.com/mistralmail/smtp/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Serve runs MistralMail
func Serve(config *Config) {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	log.Println("MistralMail at your service!")

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

	// Run admin api
	api, err := api.New(api.Config{HTTPAddress: config.HTTPAddress, Secret: []byte(config.Secret)}, backend)
	if err != nil {
		log.Fatalf("Couldn't create API: %v", err)
	}
	go func() {
		err := api.Serve()
		if err != nil {
			log.Fatalf("Couldn't serve API: %v", err)
		}
	}()

	// Serve metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Serving metrics on %s/metrics", config.MetricsAddress)
		err := http.ListenAndServe(config.MetricsAddress, nil)
		if err != nil {
			log.Fatalf("Couldn't serve metrics: %v", err)
		}
	}()

	// Create certificates store
	certificates, err := certificates.NewCertificateService(config.TLSCertificatesDirectory, config.AcmeEndpoint, config.AcmeEmail)
	if err != nil {
		log.Fatalf("Couldn't create certificate service: %v", err)
	}
	// Create all the certificates
	msaCert, err := certificates.GetOrCreateCertificate(config.SubDomainOutgoing)
	if err != nil {
		log.Fatalf("Couldn't create MSA certificate: %v", err)
	}
	mtaCert, err := certificates.GetOrCreateCertificate(config.SubDomainIncoming)
	if err != nil {
		log.Fatalf("Couldn't create MTA certificate: %v", err)
	}
	imapCert, err := certificates.GetOrCreateCertificate(config.SubDomainIMAP)
	if err != nil {
		log.Fatalf("Couldn't create IMAP certificate: %v", err)
	}

	// Run SMTP MSA
	go func() {
		msaConfig := config.GenerateMSAConfig()
		msaConfig.TlsCert = msaCert.CertificateFile
		msaConfig.TlsKey = msaCert.PrivateKeyFile

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
		err = msa.ListenAndServe()
		if err != nil {
			log.Errorln(err)
		}
	}()

	// Run IMAP
	go func() {
		imapConfig := config.GenerateIMAPBackendConfig()
		imapConfig.TlsCert = imapCert.CertificateFile
		imapConfig.TlsKey = imapCert.PrivateKeyFile

		imap.Serve(imapConfig, backend.IMAPBackend)
	}()

	// Run SMTP MTA
	mtaConfig := config.GenerateMTAConfig()
	mtaConfig.TlsCert = mtaCert.CertificateFile
	mtaConfig.TlsKey = mtaCert.PrivateKeyFile

	mtaHandlerChain := &handlers.HandlerMachanism{
		Handlers: []handlers.Handler{
			received.New(mtaConfig),
			spf.New(mtaConfig),
			imaphandler.New(mtaConfig, backend.IMAPBackend),
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
