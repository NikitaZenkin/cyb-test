package entity

import "time"

type IPExpiresAt struct {
	IP        string
	ExpiresAt time.Time
}

type FqdnIpExpiresAt struct {
	FQDN      string
	IP        string
	ExpiresAt time.Time
}

type IpFQDNs map[string][]string
