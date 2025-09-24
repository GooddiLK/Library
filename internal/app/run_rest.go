package app

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/project/library/config"
	generated "github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
)

func runRest(ctx context.Context, cfg *config.Config, logger *zap.Logger) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	err := generated.RegisterLibraryHandlerFromEndpoint(ctx, mux, address, opts)

	if err != nil {
		logger.Error("can not register grpc gateway: ", zap.Error(err))
		os.Exit(-1)
	}

	gatewayPort := ":" + cfg.GatewayPort
	logger.Info("gateway listening at port: ", zap.String("port", gatewayPort))

	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error: ", zap.Error(err))
	}
}
