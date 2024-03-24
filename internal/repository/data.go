package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"cyb-test/entity"
)

const (
	chunkSizeForIN     = 3000
	chunkSizeForInsert = 1500
)

func (r *Repository) FQDNsGet(ctx context.Context, ips []string) (entity.IpFQDNs, error) {
	if len(ips) == 0 {
		return entity.IpFQDNs{}, nil
	}

	type row struct {
		FQDN string `db:"fqdn"`
		IP   string `db:"ip"`
	}

	var data []*row
	result := make(entity.IpFQDNs, len(ips))

	chunks := sliceSplit[string](ips, chunkSizeForIN)

	for _, chunk := range chunks {
		query, args, err := sqlx.In(`SELECT fqdn, ip FROM data WHERE is_active AND ip IN (?)`, chunk)
		query = sqlx.Rebind(sqlx.DOLLAR, query)
		if err != nil {
			return nil, fmt.Errorf("build IN query: %w", err)
		}

		if err = r.Executor(ctx).SelectContext(ctx, &data, query, args...); err != nil {
			return nil, fmt.Errorf("select data: %w", err)
		}

		for _, next := range data {
			result[next.IP] = append(result[next.IP], next.FQDN)
		}
	}

	return result, nil
}

func (r *Repository) DataSave(ctx context.Context, models []*entity.FqdnIpExpiresAt) error {
	if len(models) == 0 {
		return nil
	}

	chunks := sliceSplit[*entity.FqdnIpExpiresAt](models, chunkSizeForInsert)

	for _, chunk := range chunks {
		builder := squirrel.Insert("data").Columns("fqdn", "ip", "expires_at")

		for _, data := range chunk {
			builder = builder.Values(data.FQDN, data.IP, data.ExpiresAt)
		}

		query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
		if err != nil {
			return fmt.Errorf("build INSERT query: %w", err)
		}

		if _, err = r.Executor(ctx).ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("insert data: %w", err)
		}
	}

	return nil
}

func (r *Repository) ExpiredFQDNsGet(ctx context.Context) ([]string, error) {
	fqdns := make([]string, 0)
	if err := r.Executor(ctx).SelectContext(ctx, &fqdns,
		`SELECT DISTINCT fqdn FROM data 
                     WHERE is_active AND expires_at <= now()
                     LIMIT 1000`,
	); err != nil {
		return nil, fmt.Errorf("select expired fqdns: %w", err)
	}

	return fqdns, nil
}

func (r *Repository) FQDNsSetNotActive(ctx context.Context, fqdns []string) error {
	if len(fqdns) == 0 {
		return nil
	}

	chunks := sliceSplit[string](fqdns, chunkSizeForIN)

	for _, chunk := range chunks {
		query, args, err := sqlx.In(`UPDATE data SET is_active = false WHERE fqdn IN (?)`, chunk)
		query = sqlx.Rebind(sqlx.DOLLAR, query)
		if err != nil {
			return fmt.Errorf("build IN query: %w", err)
		}

		if _, err = r.Executor(ctx).ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("update data is_active: %w", err)
		}
	}

	return nil
}
