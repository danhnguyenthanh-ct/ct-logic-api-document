package controller

import (
	"context"
	"os"

	"github.com/carousell/ct-go/pkg/cronjob"
	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
	"github.com/ct-logic-api-document/internal/constants"
	buildstructure "github.com/ct-logic-api-document/internal/usecase/build_structure"
	fetchdata "github.com/ct-logic-api-document/internal/usecase/fetch_data"
)

type CronJobOptions struct {
	Name    string
	Handler func(ctx context.Context) error
}

func NewCronJob(
	fetchDataUC fetchdata.IFetchDataUC,
	buildStructureUC buildstructure.IBuildStructureUC,
) (map[string]CronJobOptions, error) {
	ctx := context.Background()
	logctx.AppendName(ctx, "cron_job")
	cronJobHandlerMap := map[string]CronJobOptions{
		constants.CommandFetchDataFromGcs: {
			Name:    constants.CommandFetchDataFromGcs,
			Handler: fetchDataUC.FetchDataFromGcs,
		},
		constants.CommandFetchDataFromLocal: {
			Name:    constants.CommandFetchDataFromLocal,
			Handler: fetchDataUC.FetchDataFromLocal,
		},
		constants.CommandBuildStructure: {
			Name:    constants.CommandBuildStructure,
			Handler: buildStructureUC.BuildStructure,
		},
	}
	argsWithProg := os.Args
	cronJobType := argsWithProg[2]
	cronJobOpt, ok := cronJobHandlerMap[cronJobType]
	if !ok {
		panic("cronjob type not found")
	}
	cronjob.StartCronjobWithMetric(cronJobOpt.Handler, cronJobType)
	return cronJobHandlerMap, nil
}
