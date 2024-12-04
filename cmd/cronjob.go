package cmd

import (
	"github.com/ct-logic-api-document/internal/controller"
	"github.com/spf13/cobra"
)

var cronjobCmd = &cobra.Command{
	Use:   "cronjob",
	Short: "cronjob",
	Long:  "cronjob",
	Run: func(_ *cobra.Command, _ []string) {
		// agent.Start()
		// defer agent.Stop()
		Invoke(controller.NewCronJob)
	},
}
