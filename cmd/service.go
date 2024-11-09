package cmd

import (
	"context"
	"fmt"

	"github.com/carousell/ct-go/pkg/gateway"
	"github.com/carousell/ct-go/pkg/logger"
	"github.com/ct-logic-api-document/config"
	"github.com/ct-logic-api-document/internal/handler"
	"github.com/ct-logic-api-document/internal/usecase"
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
	hdl *handler.Handler,
) *gateway.Server {
	log := logger.MustNamed("server")
	ctx := context.Background()

	serverConfig := gateway.NewServerConfig().
		SetLogger(logger.MustNamed("gateway")).
		SetHTTPAddr(conf.App.HTTPAddr).
		RegisterGRPC(func(s *grpc.Server) {}).
		RegisterHTTP(func(mux *runtime.ServeMux, conn *grpc.ClientConn) {
			handler.RegisterCustomHTTPHandler(ctx, conf, mux, hdl)
		})
	server, err := gateway.NewServer(serverConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
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
			usecase.NewInputUC,
			handler.NewInputHandler,
			handler.NewHandler,
		),
		fx.Supply(conf),
		fx.Invoke(invokers...),
	)

	return app
}
