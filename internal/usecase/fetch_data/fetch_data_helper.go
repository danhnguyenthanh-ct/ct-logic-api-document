package fetchdata

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/carousell/ct-go/pkg/container"
	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	"github.com/google/uuid"
)

func parseRawUrl(ctx context.Context, rawUrl string) (string, string) {
	parsedURL, err := url.Parse(rawUrl)
	if err != nil {
		logctx.Errorw(ctx, "failed to parse url", "err", err)
		return "", ""
	}
	host := strings.Split(parsedURL.Host, ":")[0]
	return host, parsedURL.Path
}

func findParameterInPath(_ context.Context, path string) string {
	pathParts := strings.Split(path, "/")
	updatedPath := []string{}
	for _, pathPart := range pathParts {
		if isValidUUID(pathPart) {
			updatedPath = append(updatedPath, "{uuid}")
		} else if isNumber(pathPart) {
			updatedPath = append(updatedPath, "{id}")
		} else {
			updatedPath = append(updatedPath, pathPart)
		}
	}
	return strings.Join(updatedPath, "/")
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func buildParametersFromQueryString(queryStringInMap container.Map) []*entity.Parameter {
	headers := []*entity.Parameter{}
	for key, value := range queryStringInMap {
		headers = append(headers, &entity.Parameter{
			Name:  key,
			Type:  getPossibleType(key),
			Value: value.(string),
			In:    constants.ParameterInQuery,
		})
	}
	return headers
}

func getPossibleType(s string) string {
	if constants.ParametersTypeInteger.Contains(s) {
		return constants.ParameterTypeInteger
	}
	if _, err := strconv.ParseBool(s); err == nil {
		return constants.ParameterTypeBoolean
	}
	if _, err := strconv.Atoi(s); err == nil {
		return constants.ParameterTypeInteger
	}
	return constants.ParameterTypeString
}
