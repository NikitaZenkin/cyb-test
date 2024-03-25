package service

import (
	"context"

	"go.uber.org/zap"

	"cyb-test/entity"
	"cyb-test/internal/repository/txmanager"
)

type repository interface {
	WithTx(ctx context.Context, exec txmanager.TxBody) error
	FQDNsGet(ctx context.Context, ips []string) (entity.IpFQDNs, error)
	DataSave(ctx context.Context, models []*entity.FqdnIpExpiresAt) error
	ExpiredFQDNsGet(ctx context.Context) ([]string, error)
	FQDNsSetNotActive(ctx context.Context, fqdns []string) error
}

type dnsServer interface {
	IPsGet(ctx context.Context, fqdn string) ([]*entity.IPExpiresAt, error)
}

type Service struct {
	log *zap.Logger
	rep repository
	dns dnsServer

	fqdnChan chan []string
}

func New(ctx context.Context, log *zap.Logger, rep repository, dns dnsServer) *Service {
	srv := &Service{
		log:      log,
		rep:      rep,
		dns:      dns,
		fqdnChan: make(chan []string),
	}

	go func() {
		srv.startWatcher(ctx)
	}()

	return srv
}

func (s *Service) FQDNsLoad(_ context.Context, fqdns []string) error {
	s.fqdnChan <- fqdns

	return nil
}

func (s *Service) FQDNsGet(ctx context.Context, ips []string) (entity.IpFQDNs, error) {
	var (
		err    error
		result entity.IpFQDNs
	)

	if err = s.rep.WithTx(ctx, func(ctx context.Context) error {
		result, err = s.rep.FQDNsGet(ctx, ips)
		if err != nil {
			s.log.Error("failed to get fqdns", zap.Error(err))
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}
