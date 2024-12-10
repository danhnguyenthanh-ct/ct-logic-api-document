package buildstructure

import (
	"context"
	"encoding/json"
	"time"

	"github.com/carousell/ct-go/pkg/workerpool"
	"github.com/ct-logic-api-document/config"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	"github.com/ct-logic-api-document/internal/repository/mongodb"

	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
)

const (
	maxApisProccessing = 100
	defaultLimit       = 20
)

type IBuildStructureUC interface {
	BuildStructure(ctx context.Context) error
	DoBuildStructure(ctx context.Context, api *entity.Api) error
}

type buildStructureIC struct {
	conf    *config.Config
	storage mongodb.MongoStorage
}

func NewBuildStructureUC(
	conf *config.Config,
	storage mongodb.MongoStorage,
) IBuildStructureUC {
	return &buildStructureIC{
		conf:    conf,
		storage: storage,
	}
}

func (f *buildStructureIC) BuildStructure(ctx context.Context) error {
	pool := workerpool.NewE(maxApisProccessing)
	defer func() {
		pool.Close()
	}()
	offset := int64(0)
	for {
		apis, err := f.storage.GetApis(ctx, defaultLimit, offset)
		if err != nil {
			return err
		}
		if len(apis) == 0 {
			break
		}
		for _, api := range apis {
			api := api
			pool.Run(func() error {
				return f.DoBuildStructure(ctx, api)
			})
		}
		offset += defaultLimit
	}
	if err := pool.Wait(); err != nil {
		return err
	}
	return nil
}

func (f *buildStructureIC) DoBuildStructure(ctx context.Context, api *entity.Api) error {
	if api == nil || api.Id.IsZero() {
		return nil
	}
	logctx.Infof(ctx, "build structure for api: %s", api.Path)
	// handle build request structure
	if err := f.buildRequestStructure(ctx, api); err != nil {
		return err
	}
	// handle build response structure
	if err := f.buildResponseStructure(ctx, api); err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := f.storage.UpdateLatestBuildStructure(ctx, api.Id, &now); err != nil {
		return err
	}
	return nil
}

func (f *buildStructureIC) buildRequestStructure(ctx context.Context, api *entity.Api) error {
	logctx.Info(ctx, "build request structure")
	// validate if sample requests are updated
	getSampleRequestByApiIdRequest := &entity.GetSampleRequestByApiIdRequest{
		ApiId:  api.Id,
		Limit:  1,
		Offset: 0,
	}
	sampleRequests, err := f.storage.GetSampleRequestByApiId(ctx, getSampleRequestByApiIdRequest)
	if err != nil {
		return err
	}
	if len(sampleRequests) == 0 {
		logctx.Infof(ctx, "no sample requests found for api: %s", api.Path)
		return nil
	}

	if api.LatestBuildStructure != nil && api.LatestBuildStructure.Before(*sampleRequests[0].CreatedAt) {
		logctx.Infof(ctx, "sample requests are not updated for api: %s", api.Path)
		return nil
	}
	// handle build request structure
	return f.buildRequestStructureBySampleRequest(ctx, api)
}

func (f *buildStructureIC) buildResponseStructure(ctx context.Context, api *entity.Api) error {
	logctx.Info(ctx, "build response structure")
	// validate if sample responses are updated
	getSampleResponseByApiIdRequest := &entity.GetSampleResponseByApiIdRequest{
		ApiId:  api.Id,
		Limit:  1,
		Offset: 0,
	}
	sampleResponses, err := f.storage.GetSampleResponseByApiId(ctx, getSampleResponseByApiIdRequest)
	if err != nil {
		return err
	}
	if len(sampleResponses) == 0 {
		logctx.Infof(ctx, "no sample responses found for api: %s", api.Path)
		return nil
	}
	if api.LatestBuildStructure != nil && api.LatestBuildStructure.Before(*sampleResponses[0].CreatedAt) {
		logctx.Infof(ctx, "sample responses are not updated for api: %s", api.Path)
		return nil
	}
	// handle build response structure
	return f.buildResponseStructureBySampleResponse(ctx, api)
}

func (f *buildStructureIC) buildRequestStructureBySampleRequest(ctx context.Context, api *entity.Api) error {
	// handle build request structure
	logctx.Info(ctx, "build request structure by sample request")
	requestStructure, err := f.storage.GetRequestStructureByApiId(ctx, api.Id)
	if err != nil {
		return err
	}
	// build parameter structure
	currentParameters := make(map[string]*entity.Parameter)
	if requestStructure != nil && len(requestStructure.Parameters) > 0 {
		for _, parameter := range requestStructure.Parameters {
			currentParameters[parameter.Name] = &entity.Parameter{
				Name:     parameter.Name,
				Type:     parameter.Type,
				In:       parameter.In,
				Required: parameter.Required,
			}
		}
	}
	// build body structure
	currentBodySchema := map[string]any{}
	if requestStructure != nil && requestStructure.BodySchema != nil {
		currentBodySchema = requestStructure.BodySchema
	}
	// load sample requests
	offset := int64(0)
	for {
		getSampleRequestByApiIdRequest := &entity.GetSampleRequestByApiIdRequest{
			ApiId:  api.Id,
			Limit:  20,
			Offset: offset,
		}
		if api.LatestBuildStructure != nil {
			getSampleRequestByApiIdRequest.From = api.LatestBuildStructure
		}
		sampleRequests, err := f.storage.GetSampleRequestByApiId(ctx, getSampleRequestByApiIdRequest)
		if err != nil {
			return err
		}
		if len(sampleRequests) == 0 {
			break
		}
		for _, sampleRequest := range sampleRequests {
			// update parameter structure
			if len(sampleRequest.Parameters) > 0 {
				for _, parameter := range sampleRequest.Parameters {
					if currentParameter, ok := currentParameters[parameter.Name]; !ok {
						currentParameters[parameter.Name] = &entity.Parameter{
							Name:     parameter.Name,
							Type:     parameter.Type,
							In:       parameter.In,
							Required: parameter.Required,
						}
					} else {
						if currentParameter.Type != parameter.Type {
							currentParameter.Type = constants.ParameterTypeAny
						}
					}
				}
			}
			// update body structure
			if sampleRequest.Body != "" {
				body := map[string]any{}
				if err := json.Unmarshal([]byte(sampleRequest.Body), &body); err != nil {
					body := []any{}
					if err = json.Unmarshal([]byte(sampleRequest.Body), &body); err != nil {
						return err
					}
				}
				sampleBodySchema := generateSchema(body)
				currentBodySchema = mergeMaps(currentBodySchema, sampleBodySchema)
			}
		}
		offset += 20
	}
	// update request structure
	currentParametersSlice := make([]*entity.Parameter, 0, len(currentParameters))
	for _, parameter := range currentParameters {
		currentParametersSlice = append(currentParametersSlice, parameter)
	}
	if requestStructure == nil {
		requestStructure := &entity.RequestStructure{
			ApiId:      api.Id,
			Parameters: currentParametersSlice,
			BodySchema: currentBodySchema,
		}
		return f.storage.CreateRequestStructure(ctx, requestStructure)
	}
	return f.storage.UpdateRequestStructure(ctx, requestStructure.Id, currentParametersSlice, currentBodySchema)
}

func (f *buildStructureIC) buildResponseStructureBySampleResponse(ctx context.Context, api *entity.Api) error {
	// handle build response structure
	logctx.Info(ctx, "build response structure by response request")
	responseStructure, err := f.storage.GetResponseStructureByApiId(ctx, api.Id)
	if err != nil {
		return err
	}
	// build body structure
	currentBodySchema := map[string]any{}
	if responseStructure != nil && responseStructure.BodySchema != nil {
		currentBodySchema = responseStructure.BodySchema
	}
	// load sample responses
	offset := int64(0)
	for {
		getSampleResponseByApiIdRequest := &entity.GetSampleResponseByApiIdRequest{
			ApiId:  api.Id,
			Limit:  20,
			Offset: offset,
		}
		if api.LatestBuildStructure != nil {
			getSampleResponseByApiIdRequest.From = api.LatestBuildStructure
		}
		sampleResponses, err := f.storage.GetSampleResponseByApiId(ctx, getSampleResponseByApiIdRequest)
		if err != nil {
			return err
		}
		if len(sampleResponses) == 0 {
			break
		}
		for _, sampleResponse := range sampleResponses {
			// update body structure
			if sampleResponse.Body != "" {
				body := map[string]any{}
				if err := json.Unmarshal([]byte(sampleResponse.Body), &body); err != nil {
					body := []any{}
					if err = json.Unmarshal([]byte(sampleResponse.Body), &body); err != nil {
						return err
					}
				}
				sampleBodySchema := generateSchema(body)
				currentBodySchema = mergeMaps(currentBodySchema, sampleBodySchema)
			}
		}
		offset += 20
	}
	// update response structure
	if responseStructure == nil {
		responseStructure := &entity.ResponseStructure{
			ApiId:      api.Id,
			BodySchema: currentBodySchema,
		}
		return f.storage.CreateResponseStructure(ctx, responseStructure)
	}
	return f.storage.UpdateResponseStructure(ctx, responseStructure.Id, currentBodySchema)
}
