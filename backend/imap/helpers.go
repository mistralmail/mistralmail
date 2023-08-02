package imapbackend

import (
	"database/sql/driver"
	"encoding/json"
)

type StringSlice []string

func (ss *StringSlice) Scan(src interface{}) error {
	switch data := src.(type) {
	case []byte:
		return json.Unmarshal(data, ss)
	case string:
		return json.Unmarshal([]byte(data), ss)
	}
	return nil // or an error
}

func (ss StringSlice) Value() (driver.Value, error) {
	return json.Marshal(ss)
}
