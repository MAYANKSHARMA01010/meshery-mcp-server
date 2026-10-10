package config

import "encoding/json"

const redacted = "[REDACTED]"

type Secret string

func (s Secret) String() string {
	if s == "" {
		return ""
	}
	return redacted
}

func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
