package models

import "encoding/json"

type EmailCode struct {
	Email string
	Code  string
}

func (u EmailCode) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *EmailCode) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}
