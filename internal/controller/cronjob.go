package controller

import (
	"context"
	"os"

	"github.com/carousell/ct-go/pkg/cronjob"
	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/usecase"
)

type CronJobOptions struct {
	Name    string
	Handler func(ctx context.Context) error
}

func NewCronJob(
	fetchDataUC usecase.IFetchDataUC,
) (map[string]CronJobOptions, error) {
	ctx := context.Background()
	logctx.AppendName(ctx, "cron_job")
	cronJobHandlerMap := map[string]CronJobOptions{
		constants.CommandFetchDataFromGcs: {
			Name:    constants.CommandFetchDataFromGcs,
			Handler: fetchDataUC.FetchDataFromGcs,
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
