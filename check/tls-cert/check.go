package tlscert

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	Host   string
	Addr   string
	MinTTL time.Duration
}

type Config struct {
	Host   string
	Port   int
	MinTTL string
}

func init() {
	check.Register("tls-cert", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	minTTL, err := time.ParseDuration(conf.MinTTL)

	if err != nil {
		return nil, err
	}
	return &Check{conf.Host, fmt.Sprintf("%s:%d", conf.Host, conf.Port), minTTL}, nil
}

func (c Check) Exec() result.Result {
	conn, err := tls.Dial("tcp", c.Addr, nil)

	if err != nil {
		return result.Fail("failed to connect to %s: %s", c.Addr, err)
	}
	defer conn.Close()

	if err := conn.VerifyHostname(c.Host); err != nil {
		return result.Fail("failed to verify hostname %s: %s", c.Host, err)
	}
	exp := conn.ConnectionState().PeerCertificates[0].NotAfter
	ttl := time.Until(exp).Truncate(time.Second)

	if ttl < c.MinTTL {
		return result.Fail("certificate of %s expires in %s", c.Addr, ttl)
	}
	return result.OK("certificate of %s expires in %s", c.Addr, ttl)
}
