package constants

import "github.com/carousell/ct-go/pkg/container"

const (
	TypeRequest  string = "request"
	TypeResponse string = "response"
)

var ValidTypes = container.List[string]{TypeRequest, TypeResponse}
