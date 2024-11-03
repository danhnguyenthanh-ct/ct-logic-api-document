package controller

import (
	"github.com/carousell/ct-go/pkg/logger"
	"github.com/ct-logic-standard/internal/usecase"
)

type Controller struct {
	log *logger.Logger
	uc  usecase.AdListingUC
}

func NewController(
	uc usecase.AdListingUC,
) *Controller {
	h := &Controller{
		log: logger.MustNamed("controller"),
		uc:  uc,
	}
	return h
}
