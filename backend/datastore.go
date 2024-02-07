package backend

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mistralmail/mistralmail/backend/models"
	"github.com/xo/dburl"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// initDB creates a new datastore with the given database connection string/url
// e.g. postgres://user:pass@localhost/dbname
// e.g. sqlite:/path/to/file.db
func initDB(dbURL string) (*gorm.DB, error) {

	var db *gorm.DB

	u, err := dburl.Parse(dbURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse database connection url: %w", err)
	}

	c := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Warn, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      true,        // Don't include params in the SQL log
				Colorful:                  true,        // Enable color
			},
		),
	}

	switch u.Driver {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(u.DSN), c)

	case "postgres":
		db, err = gorm.Open(postgres.Open(u.DSN), c)

	case "mysql":
		db, err = gorm.Open(mysql.Open(u.DSN), c)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", u.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to establish a database connection: %w", err)
	}

	err = migrate(db)
	if err != nil {
		return nil, fmt.Errorf("couldn't migrate: %w", err)
	}

	return db, nil

}

// migrate the database models.
func migrate(db *gorm.DB) error {

	// Migrate
	// TODO: how to do this properly?
	err := db.AutoMigrate(
		&models.User{},
		&models.Mailbox{},
		&models.Message{},
	)
	if err != nil {
		return err
	}

	err = db.Migrator().DropView(models.MessageWithSequenceNumberViewName)
	if err != nil {
		return err
	}

	err = db.Migrator().CreateView(
		models.MessageWithSequenceNumberViewName,
		gorm.ViewOption{Query: db.Raw(models.MessageWithSequenceNumberViewQuery)},
	)
	if err != nil {
		return err
	}

	return nil

}

// closeDB closes the database connection
func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("couldn't get sql db: %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("couldn't close db connection: %w", err)
	}
	return nil
}

/*
func seedDB() {

	user := &User{Username_: "username@example.com", Password: "password", Email: "username@localhost"}

	result := db.Create(&user)
	if result.Error != nil {
		log.Warnf("couldn't seed user: %v", result.Error)
		return
	}

	mailbox := &Mailbox{Name_: "INBOX", User: user}
	result = db.Create(&mailbox)
	if result.Error != nil {
		log.Warnf("couldn't seed mailbox: %v", result.Error)
		return
	}

	body := "From: contact@example.org\r\n" +
		"To: contact@example.org\r\n" +
		"Subject: A little message, just for you\r\n" +
		"Date: Wed, 11 May 2016 14:31:59 +0000\r\n" +
		"Message-ID: <0000000@localhost/>\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hi there :)"

	message := &Message{
		UID:   1,
		Date:  time.Now(),
		Flags: []string{"\\Seen"},
		Size:  uint32(len(body)),
		Body:  []byte(body),

		MailboxID: mailbox.ID,
	}
	result = db.Create(&message)
	if result.Error != nil {
		log.Warnf("couldn't seed message: %v", result.Error)
		return
	}

}
*/
