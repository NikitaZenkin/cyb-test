package dns

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/miekg/dns"

	"cyb-test/entity"
)

type controller struct {
	client        *dns.Client
	serverAddress string
}

func New(serverAddress string) *controller {
	return &controller{
		client:        &dns.Client{},
		serverAddress: serverAddress,
	}
}

const maxRetries = 4

func (c *controller) IPsGet(ctx context.Context, fqdn string) ([]*entity.IPExpiresAt, error) {
	client := dns.Client{}

	msg := dns.Msg{}
	msg.SetQuestion(dns.Fqdn(fqdn), dns.TypeA)

	backOff := backoff.NewConstantBackOff(time.Second)
	backOffWithConditions := backoff.WithContext(backoff.WithMaxRetries(backOff, maxRetries), ctx)

	var (
		resp *dns.Msg
		err  error
	)

	operation := func() error {
		resp, _, err = client.Exchange(&msg, c.serverAddress)
		return err
	}

	err = backoff.Retry(operation, backOffWithConditions)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup: %w", err)
	}

	result := make([]*entity.IPExpiresAt, 0, len(resp.Answer))
	now := time.Now()

	for _, answer := range resp.Answer {
		if a, ok := answer.(*dns.A); ok {
			result = append(result, &entity.IPExpiresAt{
				IP:        a.A.String(),
				ExpiresAt: now.Add(time.Second * time.Duration(a.Header().Ttl)),
			})
		}
	}

	return result, nil
}
