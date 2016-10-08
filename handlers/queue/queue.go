package queue

import (
	"os"

	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/smtp/mta"
	"github.com/gopistolet/smtp/smtp"
)

// TODO: I just copypasted this from the old handlers file. It isn't used anymore at the moment
// would be good if we could build this into the HandlerMechanism

var mailQueue = make(chan smtp.State)

func handleQueue(state *smtp.State) {
	// Save mail to disk
	save(state)

	// Put mail in mail queue
	mailQueue <- (*state)
}

func MailQueueWorker(q chan smtp.State, handler mta.Handler) {

	for {
		state := <-q

		// Handle mail
		handler.HandleMail(&state)

		// Remove mail from disk
		delete(&state)

	}

}

func fileNameForState(state *smtp.State) (s string) {
	s += state.SessionId.String()
	s += "." + state.From.String()
	s += ".json"
	return
}

// Save mails to disk, since we are responsible for the message do be delivered
func save(state *smtp.State) {

	filename := "mailstore/" + fileNameForState(state)

	err := helpers.EncodeFile(filename, state)
	if err != nil {
		log.Fatal("Couldn't save mail to disk: ", err.Error())
	}

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
	}).Debug("Serialized mail to disk: ", filename)

}

func delete(state *smtp.State) {
	filename := "mailstore/" + fileNameForState(state)
	err := os.Remove(filename)
	if err != nil {
		log.Warnln("Couldn't save mail to disk: ", err.Error())
		return
	}

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
	}).Debug("Removed temp mail from disk: ", filename)
}
