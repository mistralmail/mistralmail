package smtp

import "net/mail"
import "io"

type MailMessage mail.Message

func ReadMessage(r io.Reader) (*MailMessage, error) {
    m, err := mail.ReadMessage(r)
    if err != nil {
        return nil, err
    }
    msg := MailMessage(*m)
    return &msg, err
}