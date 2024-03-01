package dnsrecord

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Config struct {
	Address string
	Type    string
	Name    string
	Value   string
}

type Check struct {
	conf     Config
	resolver *net.Resolver
	lookup   func(*net.Resolver, string) ([]string, error)
}

func init() {
	check.Register("dns-record", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	if conf.Name == "" {
		return nil, errors.New("missing Name")
	}
	check := Check{conf: conf}

	if conf.Address == "" {
		check.resolver = net.DefaultResolver
	} else {
		check.resolver = resolver(conf.Address)
	}

	switch conf.Type {
	case "A":
		check.lookup = lookupA
	case "AAAA":
		check.lookup = lookupAAAA
	case "CNAME":
		check.lookup = lookupCNAME
	case "TXT":
		check.lookup = lookupTXT
	default:
		return nil, fmt.Errorf("unsupported record type: %#v", conf.Type)
	}
	return check, nil
}

func (c Check) Exec() result.Result {
	records, err := c.lookup(c.resolver, c.conf.Name)

	if err != nil {
		return result.Fail("failed to lookup %s: %s", c.conf.Name, err)
	}
	for _, r := range records {
		if r == c.conf.Value {
			return result.OK("%s %s resolves to %#v", c.conf.Type, c.conf.Name, c.conf.Value)
		}
	}
	if len(records) == 0 {
		return result.Fail("could not find any records for %s %s", c.conf.Type, c.conf.Name)
	}
	return result.Fail("%s %s resolves to %#v", c.conf.Type, c.conf.Name, records[0])
}

func resolver(addr string) *net.Resolver {
	d := &net.Dialer{Timeout: time.Second}

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network string, _ string) (net.Conn, error) {
			return d.DialContext(ctx, network, addr)
		},
	}
}

func lookupA(r *net.Resolver, name string) ([]string, error) {
	records, err := r.LookupIP(context.Background(), "ip4", name)

	if err != nil {
		return nil, err
	}
	return mapToStrings(records), nil
}

func lookupAAAA(r *net.Resolver, name string) ([]string, error) {
	records, err := r.LookupIP(context.Background(), "ip6", name)

	if err != nil {
		return nil, err
	}
	return mapToStrings(records), nil
}

func lookupCNAME(r *net.Resolver, name string) ([]string, error) {
	record, err := r.LookupCNAME(context.Background(), name)

	if err != nil {
		return nil, err
	}
	return []string{record}, nil
}

func lookupTXT(r *net.Resolver, name string) ([]string, error) {
	return r.LookupTXT(context.Background(), name)
}

func mapToStrings[T fmt.Stringer](values []T) []string {
	s := make([]string, len(values))

	for i, val := range values {
		s[i] = val.String()
	}
	return s
}
