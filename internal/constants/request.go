package constants

import "github.com/carousell/ct-go/pkg/container"

const (
	ParameterInQuery = "query"
)

const (
	ParameterTypeString  = "string"
	ParameterTypeBoolean = "boolean"
	ParameterTypeInteger = "integer"
)

var (
	ParametersTypeInteger = container.List[string]{
		"limit",
		"offset",
		"skip",
		"page",
		"size",
	}
)
