package loadstructure

import (
	"context"
	"strings"

	"github.com/carousell/ct-go/pkg/container"
	"github.com/ct-logic-api-document/config"
	"github.com/ct-logic-api-document/internal/entity"
	"github.com/ct-logic-api-document/internal/errors"
	"github.com/ct-logic-api-document/internal/repository/mongodb"
)

type ILoadstructure interface {
	LoadStructureByApiId(ctx context.Context, apiId string) (*entity.LoadStructureByApiIdResponse, error)
}

type loadStructureUC struct {
	conf    *config.Config
	storage mongodb.MongoStorage
}

func NewLoadStructureUC(
	conf *config.Config,
	storage mongodb.MongoStorage,
) ILoadstructure {
	return &loadStructureUC{
		conf:    conf,
		storage: storage,
	}
}

func (uc *loadStructureUC) LoadStructureByApiId(ctx context.Context, apiId string) (*entity.LoadStructureByApiIdResponse, error) {
	apiObject, err := uc.storage.GetApiByIdInStr(ctx, apiId)
	if err != nil {
		return nil, err
	}
	if apiObject == nil {
		return nil, errors.ErrApiNotFound
	}
	requestStructureByApiId, err := uc.storage.GetRequestStructureByApiId(ctx, apiObject.Id)
	if err != nil {
		return nil, err
	}
	responseStructureByApiId, err := uc.storage.GetResponseStructureByApiId(ctx, apiObject.Id)
	if err != nil {
		return nil, err
	}

	resp := uc.buildLoadStructureByApiIdResponse(ctx, apiObject, requestStructureByApiId, responseStructureByApiId)
	return resp, nil
}

func (uc *loadStructureUC) buildLoadStructureByApiIdResponse(_ context.Context,
	apiObject *entity.Api,
	requestStructure *entity.RequestStructure,
	responseStructure *entity.ResponseStructure,
) *entity.LoadStructureByApiIdResponse {
	resp := &entity.LoadStructureByApiIdResponse{}
	resp.LoadDefault()
	resp.Servers = []*entity.Server{
		{
			Url: apiObject.Host,
		},
	}
	apiInfo := container.Map{}
	if strings.Contains(apiObject.Path, "private") {
		apiInfo["security"] = []container.Map{
			{
				"Bearer": []string{},
			},
		}
	}
	if parameters := requestStructure.BuildParameters(); parameters != nil {
		apiInfo["parameters"] = parameters
	}
	if requestBody := requestStructure.BuildRequestBody(); requestBody != nil {
		apiInfo["requestBody"] = requestBody
	}
	if responseBody := responseStructure.BuildResponseBody(); responseBody != nil {
		apiInfo["responses"] = responseBody
	}
	resp.Paths[apiObject.Path] = container.Map{
		strings.ToLower(apiObject.Method): apiInfo,
	}

	return resp
}
