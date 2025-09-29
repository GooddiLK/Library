package app

import (
	"net"
	"os"

	"github.com/project/library/config"
	generated "github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func runGrpc(cfg *config.Config, logger *zap.Logger, libraryService generated.LibraryServer) {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)

	if err != nil {
		logger.Error("Can not open tcp socket.", zap.Error(err))
		os.Exit(-1)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	generated.RegisterLibraryServer(s, libraryService)

	logger.Info("Grpc server listening.", zap.String("port", port))

	if err = s.Serve(lis); err != nil {
		logger.Error("Grpc server listen error.", zap.Error(err))
	}

	// Можно было бы добавить остановку
}
