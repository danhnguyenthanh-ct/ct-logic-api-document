package entity

import "github.com/ct-logic-api-document/internal/constants"

type CreateInputRequest struct {
	Endpoint string `json:"endpoint"`
	Type     string `json:"type"`
	Headers  any    `json:"headers"`
	Body     any    `json:"body"`
}

func (r *CreateInputRequest) IsValidRequest() bool {
	if r.Endpoint == "" {
		return false
	}
	if !constants.ValidTypes.Contains(r.Type) {
		return false
	}
	return true
}
