package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Address represents the expected JSON structure.
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country"`
}

// NullableAddress safely handles NULL JSONB values.
type NullableAddress struct {
	Addr  *Address
	Valid bool
}

// Scan implements sql.Scanner.
func (a *NullableAddress) Scan(src any) error {
	if src == nil {
		a.Addr = nil
		a.Valid = false
		return nil
	}

	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("failed to scan address: invalid type")
	}

	var addr Address
	if err := json.Unmarshal(bytes, &addr); err != nil {
		return err
	}

	a.Addr = &addr
	a.Valid = true
	return nil
}

// Value implements driver.Valuer.
func (a NullableAddress) Value() (driver.Value, error) {
	if !a.Valid || a.Addr == nil {
		return nil, nil
	}

	return json.Marshal(a.Addr)
}
