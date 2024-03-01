package filemtime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	Address string
	Type    string
	Name    string
	Value   string
}

func init() {
	check.Register("dns-record", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	check := Check{}

	if err := configure(&check); err != nil {
		return nil, err
	}
	if check.Type == "" {
		return nil, errors.New("missing Type")
	}
	if check.Name == "" {
		return nil, errors.New("missing Name")
	}
	return check, nil
}

func (c Check) Exec() result.Result {
	records, err := lookup(resolver(c.Address), c.Type, c.Name)

	if err != nil {
		return result.Fail("failed to lookup %s: %s", c.Name, err)
	}
	for _, r := range records {
		if r == c.Value {
			return result.OK("%s %s resolves to %#v", c.Type, c.Name, c.Value)
		}
	}
	if len(records) == 0 {
		return result.Fail("could not find any records for %s %s", c.Type, c.Name)
	}
	return result.Fail("%s %s resolves to %#v", c.Type, c.Name, records[0])
}

func resolver(addr string) *net.Resolver {
	if addr == "" {
		return net.DefaultResolver
	}
	d := &net.Dialer{Timeout: time.Second}

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network string, _ string) (net.Conn, error) {
			return d.DialContext(ctx, network, addr)
		},
	}
}

func lookup(r *net.Resolver, typ, name string) (result []string, err error) {
	switch typ {
	case "A":
		records, err := r.LookupIP(context.Background(), "ip4", name)

		if err != nil {
			return nil, err
		}
		return mapToStrings(records), nil
	case "AAAA":
		records, err := r.LookupIP(context.Background(), "ip6", name)

		if err != nil {
			return nil, err
		}
		return mapToStrings(records), nil
	case "CNAME":
		record, err := r.LookupCNAME(context.Background(), name)

		if err != nil {
			return nil, err
		}
		return []string{record}, nil
	case "TXT":
		return r.LookupTXT(context.Background(), name)
	default:
		return nil, fmt.Errorf("unsupported record type: %s", typ)
	}
}

func mapToStrings[T fmt.Stringer](values []T) []string {
	s := make([]string, len(values))

	for i, val := range values {
		s[i] = val.String()
	}
	return s
}
