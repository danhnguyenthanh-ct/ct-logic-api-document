package entity

import "github.com/carousell/ct-go/pkg/container"

type LoadStructureByApiIdResponse struct {
	OpenApi             string        `json:"openapi"`
	Info                *Info         `json:"info"`
	Servers             []*Server     `json:"servers"`
	Tags                []*Tag        `json:"tags,omitempty"`
	Paths               container.Map `json:"paths"`
	SecurityDefinitions container.Map `json:"securityDefinitions,omitempty"`
	Security            []any         `json:"security,omitempty"`
}

func (l *LoadStructureByApiIdResponse) LoadDefault() {
	l.OpenApi = "3.0.0"
	l.Info = &Info{
		Title:       "Chotot API Document",
		Description: "Chotot API Document",
		Version:     "1.0.0",
	}
	l.Servers = []*Server{}
	l.Tags = []*Tag{}
	l.Paths = make(container.Map)
	l.SecurityDefinitions = container.Map{
		"Bearer": container.Map{
			"type":        "apiKey",
			"name":        "Authorization",
			"in":          "header",
			"description": "Enter the token with the `Bearer: ` prefix, e.g. \"Bearer abcde12345\".",
		},
	}
	l.Security = []any{
		container.Map{
			"Bearer": []string{},
		},
	}
}

type Info struct {
	Title       string `json:"title" default:"Chotot API"`
	Description string `json:"description" default:"Chotot API"`
	Version     string `json:"version" default:"1.0.0"`
}

type Server struct {
	Url string `json:"url"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
