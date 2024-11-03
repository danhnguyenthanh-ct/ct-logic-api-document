package cmd

import (
	"github.com/ct-logic-standard/internal/controller"
	"github.com/spf13/cobra"
)

var workerKafka = &cobra.Command{
	Use:   "worker_kafka",
	Short: "worker_kafka",
	Long:  "worker_kafka",
	Run: func(cmd *cobra.Command, args []string) {
		Invoke(controller.NewKafkaWorker).Run()
	},
}
