package handler

import (
	"net/http"

	loadstructure "github.com/ct-logic-api-document/internal/usecase/load_structure"
	"github.com/labstack/echo/v4"
)

type LoadStructureHandler struct {
	LoadstructureUC loadstructure.ILoadstructure
}

func NewLoadStructureHandler(loadstructureUC loadstructure.ILoadstructure) *LoadStructureHandler {
	return &LoadStructureHandler{
		LoadstructureUC: loadstructureUC,
	}
}

func (h *LoadStructureHandler) RegisterHandler(internalGroup *echo.Group) {
	internalGroup.GET("/load-structure/:api_id", h.LoadStructureByApiId)
}

func (h *LoadStructureHandler) LoadStructureByApiId(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	resp, err := h.LoadstructureUC.LoadStructureByApiId(ctx, echoCtx.Param("api_id"))
	if err != nil {
		return echoCtx.JSON(http.StatusInternalServerError, err.Error())
	}
	echoCtx.Response().Header().Set("Access-Control-Allow-Origin", "*")
	return echoCtx.JSON(http.StatusOK, resp)
}
