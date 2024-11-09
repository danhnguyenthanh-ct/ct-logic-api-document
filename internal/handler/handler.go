package handler

import (
	"context"
	"net/http"
	"regexp"

	"github.com/ct-logic-api-document/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	inputHandler *InputHandler
}

func NewHandler(
	inputHandler *InputHandler,
) *Handler {
	return &Handler{
		inputHandler: inputHandler,
	}
}

func RegisterCustomHTTPHandler(
	_ context.Context,
	_ *config.Config,
	mux *runtime.ServeMux,
	handler *Handler,
) {
	e := echo.New()

	internalGroup := e.Group("/internal")

	handler.inputHandler.RegisterInternalGroup(internalGroup)

	echo.WrapHandler(mux)

	for _, route := range e.Routes() {
		// to replace any path param to *
		m1 := regexp.MustCompile(`:[a-z0-9_:]+`)
		path := m1.ReplaceAllString(route.Path, "*")
		_ = mux.HandlePath(
			route.Method,
			path,
			func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				e.ServeHTTP(w, r)
			})
	}
}
