package gopistolet

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/gopistolet/gopistolet/helpers"
	imapbackend "github.com/gopistolet/imap-backend"
	"github.com/gopistolet/smtp/mta"
	log "github.com/sirupsen/logrus"
)

var (
	defaultSMTPPort    = 25
	defaultSMTPAddress = fmt.Sprintf(":%d", defaultSMTPPort)
	defaultIMAPPort    = 143
	defaultIMAPAddress = fmt.Sprintf(":%d", defaultIMAPPort)
	defaultDatabaseURL = "sqlite:test.db"
	defaultSeedDB      = false
)

// BuildConfigFromEnv populates a GoPistolet config from env variables
func BuildConfigFromEnv() *Config {
	config := &Config{}

	config.Hostname = getEnv("HOSTNAME", "")
	config.SMTPAddress = getEnv("SMTP_ADDRESS", defaultSMTPAddress)
	config.IMAPAddress = getEnv("IMAP_ADDRESS", defaultIMAPAddress)
	config.DatabaseURL = getEnv("DATABASE_URL", defaultDatabaseURL)

	return config
}

// Config contains all the config for serving GoPistolet
type Config struct {
	Hostname    string
	SMTPAddress string
	IMAPAddress string
	DatabaseURL string

	SeedDB bool
}

// Validate validates whether all config is set and valid
func (config *Config) Validate() error {

	if config.SMTPAddress == "" {
		return fmt.Errorf("SMTPAddress cannot be empty")
	}

	if config.IMAPAddress == "" {
		return fmt.Errorf("IMAPAddress cannot be empty")
	}

	if config.DatabaseURL == "" {
		return fmt.Errorf("DatabaseURL cannot be empty")
	}

	// TODO enforce hostname to be set

	return nil
}

// GenerateMTAConfig generates the SMTP config for the MTA
func (config *Config) GenerateMTAConfig() *mta.Config {

	host, port, err := net.SplitHostPort(config.SMTPAddress)
	if err != nil {
		// TODO: handle
		log.Fatalf("couldn't determine SMTP address/port: %v", err)
	}

	portInt, _ := strconv.Atoi(port)

	nixspamBlacklist, err := helpers.NewNixspam()
	if err != nil {
		log.Warnln("couldn't create Nixspam Blacklist instance: ", err)
	}

	return &mta.Config{
		Hostname:    config.Hostname,
		Ip:          host,
		Port:        uint32(portInt),
		Blacklist:   nixspamBlacklist,
		DisableAuth: true,
	}
}

// GenerateIMAPBackendConfig generates the config object for the IMAP backend
func (config *Config) GenerateIMAPBackendConfig() *imapbackend.Config {
	return &imapbackend.Config{
		IMAPAddress: config.IMAPAddress,
		DatabaseURL: config.DatabaseURL,
		SeedDB:      config.SeedDB,
	}
}

// getEnv gets the env variable with the given key if the key exists
// else it falls back to the fallback value
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
