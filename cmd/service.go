package cmd

import (
	"context"
	"fmt"
	"github.com/carousell/ct-go/pkg/gateway"
	"github.com/carousell/ct-go/pkg/logger"
	pb "github.com/carousell/ct-grpc-go/pkg/ct-logic-standard"
	"github.com/ct-logic-standard/config"
	"github.com/ct-logic-standard/internal/controller"
	"github.com/ct-logic-standard/internal/repository/ad_listing_client"
	"github.com/ct-logic-standard/internal/repository/kafka"
	"github.com/ct-logic-standard/internal/repository/rabbitmq"
	"github.com/ct-logic-standard/internal/usecase"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"google.golang.org/grpc"
)

var service = &cobra.Command{
	Use:   "service",
	Short: "API Command of service",
	Long:  "API Command of service",
	Run: func(cmd *cobra.Command, args []string) {
		Invoke(Start).Run()
	},
}

func Start(
	lc fx.Lifecycle,
	conf *config.Config,
	controller *controller.Controller,
) *gateway.Server {
	log := logger.MustNamed("server")
	ctx := context.Background()

	serverConfig := gateway.NewServerConfig().
		SetLogger(logger.MustNamed("gateway")).
		SetGRPCAddr(conf.App.GRPCAddr).
		SetHTTPAddr(conf.App.HTTPAddr).
		RegisterGRPC(func(s *grpc.Server) {
			pb.RegisterLogicStandardServiceServer(s, controller)
		}).
		RegisterHTTP(func(mux *runtime.ServeMux, conn *grpc.ClientConn) {
			pb.RegisterLogicStandardServiceHandler(ctx, mux, conn)
		})
	server, err := gateway.NewServer(serverConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Infof("grpc server starting at: %s", conf.App.GRPCAddr)
				log.Infof("http server starting at: %s", conf.App.HTTPAddr)
				err := server.Serve(ctx)
				if err != nil {
					log.Fatalf("failed to start server: %v", err)
				}
			}()
			return nil
		},
		OnStop: server.Stop,
	})

	return server
}

func Invoke(invokers ...interface{}) *fx.App {
	conf := config.MustLoad()
	log := logger.MustNamed("app")
	log.Debugf("[config] %+v", conf)
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{
				Logger: log.Unwrap().Desugar(),
			}
		}),
		fx.StartTimeout(conf.App.StartTimeout),
		fx.StopTimeout(conf.App.StopTimeout),
		fx.Provide(
			controller.NewController,
			usecase.NewAdListingUC,
			ad_listing_client.NewAdListingClient,
			kafka.NewKafkaProducer,
			rabbitmq.NewRabbitMQProducer,
		),
		fx.Supply(conf),
		fx.Invoke(invokers...),
	)

	return app
}
