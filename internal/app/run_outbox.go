package app

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/project/library/config"
	"github.com/project/library/internal/usecase/outbox"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

const (
	MaxIdleConnections    = 100
	MaxConnectionsPerHost = 100
	IdleConnectionTimeout = 90 * time.Second
	TLSHandshakeTimeout   = 15 * time.Second
	ExpectContinueTimeout = 2 * time.Second
	Timeout               = 30 * time.Second
	KeepAlive             = 180 * time.Second
)

func runOutbox(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) {
	dialer := &net.Dialer{
		Timeout:   Timeout,
		KeepAlive: KeepAlive,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          MaxIdleConnections,
		MaxConnsPerHost:       MaxConnectionsPerHost,
		IdleConnTimeout:       IdleConnectionTimeout,
		TLSHandshakeTimeout:   TLSHandshakeTimeout,
		ExpectContinueTimeout: ExpectContinueTimeout,
	}

	client := &http.Client{Transport: transport}

	globalHandler := globalOutboxHandler(
		client, cfg.Outbox.BookSendURL, cfg.Outbox.AuthorSendURL)
	outboxService := outbox.New(
		logger, outboxRepository, globalHandler, cfg, transactor)

	outboxService.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.WaitTimeMS,
		cfg.Outbox.InProgressTTLMS,
	)
}
