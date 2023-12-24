package spamcheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	postmarkLong   = "long"
	postmarkShort  = "short"
	postmarkAPIURL = "https://spamcheck.postmarkapp.com/filter"
)

// SpamScoreAPI is an interface for getting the Spam Score values for email messages.
type SpamScoreAPI interface {
	// getSpamScore returns the spam score value.
	getSpamScore(message string) (*response, error)
}

// PostmarkAPI implements the SpamScoreAPI for Postmark.
type PostmarkAPI struct{}

// ErrEmptySpamScore denotes an error when the response from the spam check api is empty.
var ErrEmptySpamScore = fmt.Errorf("received empty spam score from api")

// getSpamScore gets the spam score from the Postmark api:
// https://spamcheck.postmarkapp.com
func (api *PostmarkAPI) getSpamScore(message string) (*response, error) {

	data := request{
		Email:   message,
		Options: postmarkLong,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", postmarkAPIURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	spamResponse := &response{}

	err = json.NewDecoder(resp.Body).Decode(spamResponse)
	if err != nil {
		return nil, err
	}

	if spamResponse.Score == "" {
		return nil, ErrEmptySpamScore
	}

	return spamResponse, nil
}

type request struct {
	Email   string `json:"email"`
	Options string `json:"options"`
}

type rule struct {
	Score       string `json:"score"`
	Description string `json:"description"`
}

type response struct {
	Success bool   `json:"success"`
	Score   string `json:"score"`
	Rules   []rule `json:"rules"`
	Report  string `json:"report"`
}
