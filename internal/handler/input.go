package handler

import (
	"net/http"

	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
	"github.com/ct-logic-api-document/internal/usecase"
	"github.com/labstack/echo/v4"
)

type InputHandler struct {
	inputUC *usecase.InputUC
}

func NewInputHandler(inputUC *usecase.InputUC) *InputHandler {
	return &InputHandler{
		inputUC: inputUC,
	}
}

func (h *InputHandler) RegisterInternalGroup(internalGroup *echo.Group) {
	internalGroup.POST("/input", h.createInput)
}

func (h *InputHandler) createInput(e echo.Context) error {
	ctx := e.Request().Context()
	hello := 1
	logctx.Infow(ctx, "hello", hello)
	return e.JSON(http.StatusOK, hello)
}
