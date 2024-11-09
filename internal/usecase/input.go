package usecase

import "github.com/ct-logic-api-document/config"

type InputUC struct {
	conf *config.Config
}

func NewInputUC(conf *config.Config) *InputUC {
	return &InputUC{
		conf: conf,
	}
}
