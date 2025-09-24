package app

import (
	"context"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/project/library/config"
	generated "github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func runRest(ctx context.Context, cfg *config.Config, logger *zap.Logger) {
	// Создание мультиплексора, преобразующего REST HTTP запросы в gRPC вызовы
	mux := runtime.NewServeMux()
	// Параметры подключения к gRPC серверу. Подключение без TLS.
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	// Связь между REST и gRPC
	err := generated.RegisterLibraryHandlerFromEndpoint(ctx, mux, address, opts)

	if err != nil {
		logger.Error("can not register grpc gateway: ", zap.Error(err))
		os.Exit(-1)
	}

	gatewayPort := ":" + cfg.GatewayPort
	logger.Info("gateway listening at port: ", zap.String("port", gatewayPort))

	// Запуск http сервера
	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error: ", zap.Error(err))
	}
}
