package cmd

import (
	"github.com/ct-logic-standard/internal/controller"
	"github.com/spf13/cobra"
)

var workerRabbitmq = &cobra.Command{
	Use:   "worker_rabbitmq",
	Short: "worker_rabbitmq",
	Long:  "worker_rabbitmq",
	Run: func(cmd *cobra.Command, args []string) {
		Invoke(controller.NewRabbitMQWorker).Run()
	},
}
