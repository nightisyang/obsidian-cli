package output

import (
	"encoding/json"
	"io"
)

type Envelope struct {
	OK    bool      `json:"ok"`
	Data  any       `json:"data,omitempty"`
	Error *ErrField `json:"error,omitempty"`
}

type ErrField struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func Success(data any) Envelope {
	return Envelope{OK: true, Data: data}
}

func Failure(code int, message string) Envelope {
	return Envelope{OK: false, Error: &ErrField{Code: code, Message: message}}
}
