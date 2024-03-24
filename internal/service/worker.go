package service

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"cyb-test/entity"
)

func (s *Service) startWatcher(ctx context.Context) error {
	var (
		fqdns []string
		err   error
	)

	defer close(s.fqdnChan)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case fqdns = <-s.fqdnChan:
			if err = s.dataUpdate(ctx, fqdns); err != nil {
				s.log.Error("failed to update data by call", zap.Error(err))
				continue
			}
		case <-ticker.C:
			fqdns, err = s.rep.ExpiredFQDNsGet(ctx)
			if err != nil {
				s.log.Error("failed to get expired fqdns", zap.Error(err))
				continue
			}

			if len(fqdns) == 0 {
				continue
			}

			if err = s.dataUpdate(ctx, fqdns); err != nil {
				s.log.Error("failed to update data by ticker", zap.Error(err))
				continue
			}
		}
	}
}

func (s *Service) dataUpdate(ctx context.Context, fqdns []string) error {
	ipsMap, err := s.updateIPs(ctx, fqdns)
	if err != nil {
		return err
	}

	succeedFqdns := make([]string, 0, len(ipsMap))
	data := make([]*entity.FqdnIpExpiresAt, 0)

	for fqdn, ips := range ipsMap {
		succeedFqdns = append(succeedFqdns, fqdn)
		for _, ip := range ips {
			data = append(data, &entity.FqdnIpExpiresAt{
				FQDN:      fqdn,
				IP:        ip.IP,
				ExpiresAt: ip.ExpiresAt,
			})
		}
	}

	if err = s.rep.WithTx(ctx, func(ctx context.Context) error {
		if err = s.rep.FQDNsSetNotActive(ctx, succeedFqdns); err != nil {
			s.log.Error("failed to set not active", zap.Error(err))
			return err
		}

		if err = s.rep.DataSave(ctx, data); err != nil {
			s.log.Error("failed to save data", zap.Error(err))
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	s.log.Info("data saved", zap.Strings("fqdns", succeedFqdns))

	return nil
}

const maxWorkers = 500

func (s *Service) updateIPs(ctx context.Context, fqdns []string) (map[string][]*entity.IPExpiresAt, error) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	fqdnChan := make(chan string)

	ipsMap := make(map[string][]*entity.IPExpiresAt, len(fqdns))

	for i := 0; i < len(fqdns) && i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(ctx, fqdnChan, mu, ipsMap)
		}()
	}

	for _, fqdn := range fqdns {
		fqdnChan <- fqdn
	}

	close(fqdnChan)
	wg.Wait()

	return ipsMap, nil
}

func (s *Service) worker(
	ctx context.Context, fqdns <-chan string, mu *sync.Mutex,
	ipsMap map[string][]*entity.IPExpiresAt,
) {
	for fqdn := range fqdns {
		ips, err := s.dns.IPsGet(ctx, fqdn)
		if err != nil {
			s.log.Error("get ips for "+fqdn, zap.Error(err))
			continue
		}

		mu.Lock()
		ipsMap[fqdn] = ips
		mu.Unlock()
	}
}
