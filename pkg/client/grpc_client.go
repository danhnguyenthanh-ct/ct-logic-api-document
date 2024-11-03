package client

import (
	"context"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/labstack/gommon/log"
	circuit "github.com/rubyist/circuitbreaker"
	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type GRPCClient struct {
	clientName     string
	domainDialHost string
	conn           *grpc.ClientConn
	breaker        *circuit.Breaker
}

func (g *GRPCClient) GetGRPCConnection() *grpc.ClientConn {
	return g.conn
}

func (g *GRPCClient) SetBreaker(breaker *circuit.Breaker) {
	g.breaker = breaker
}

func (g *GRPCClient) SetName(name string) {
	g.clientName = name
}

func (g *GRPCClient) GetName() string {
	return g.clientName
}

func (g *GRPCClient) SetDomainDialHost(domainDialHost string) {
	g.domainDialHost = domainDialHost
}

func (g *GRPCClient) NewGrpcClient() error {
	chainUnaryClient := []grpc.UnaryClientInterceptor{
		apmgrpc.NewUnaryClientInterceptor(),
	}
	opts := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffLinear(100 * time.Millisecond)),
	}
	chainUnaryClient = append(chainUnaryClient, grpcRetry.UnaryClientInterceptor(opts...))

	if g.breaker != nil {
		chainUnaryClient = append(chainUnaryClient, g.UnaryClientInterceptorWithBreaker())
	}
	conn, err := grpc.Dial(
		g.domainDialHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcMiddleware.ChainUnaryClient(chainUnaryClient...)),
	)
	if err != nil {
		log.Errorf("Cannot initial grpc client for: %v, error : %v", g.clientName, err.Error())
		return err
	}
	g.conn = conn
	return nil
}

func (g *GRPCClient) UnaryClientInterceptorWithBreaker() grpc.UnaryClientInterceptor {
	return func(parentCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := g.breaker.CallContext(parentCtx, func() error {
			return invoker(parentCtx, method, req, reply, cc, opts...)
		}, 0)
		return err
	}
}
