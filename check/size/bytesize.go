package size

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	bExp   = 0
	kibExp = 10
	mibExp = 20
	gibExp = 30
	tibExp = 40
	pibExp = 50
	eibExp = 60
	zibExp = 70
	yibExp = 80
)

type Size struct {
	Num
}

func B(n uint64) Size {
	return Size{Num{n, bExp}}
}

func KiB(n uint64) Size {
	return Size{Num{n, kibExp}}
}

func MiB(n uint64) Size {
	return Size{Num{n, mibExp}}
}

func GiB(n uint64) Size {
	return Size{Num{n, gibExp}}
}

func TiB(n uint64) Size {
	return Size{Num{n, tibExp}}
}

func PiB(n uint64) Size {
	return Size{Num{n, pibExp}}
}

func EiB(n uint64) Size {
	return Size{Num{n, eibExp}}
}

func ZiB(n uint64) Size {
	return Size{Num{n, zibExp}}
}

func YiB(n uint64) Size {
	return Size{Num{n, yibExp}}
}

func (s Size) Mul(n Num) Size {
	return Size{s.Num.Mul(n)}
}

func (s Size) Div(n Num) Size {
	return Size{s.Num.Div(n)}
}

func (s Size) DivSize(d Size) Num {
	return s.Num.Div(d.Num)
}

func (s Size) CompareTo(d Size) int {
	return s.Num.CompareTo(d.Num)
}

const unitPrefixes = "KMGTPEZY"
const maxExp = uint64(len(unitPrefixes) * 10)

func (s Size) String() string {
	shiftRight(&s.Num)

	diffExp := s.Exp % 10

	if diffExp != 0 {
		if s.Sig < maxUint64>>diffExp {
			s.Sig <<= diffExp
			s.Exp -= diffExp
		} else {
			s.Sig >>= (10 - diffExp)
			s.Exp += (10 - diffExp)
		}
	}
	if s.Exp > maxExp {
		s.Sig <<= s.Exp - maxExp
		s.Exp = maxExp
	}
	var frc uint64

	for s.Sig >= 1<<10 && s.Exp < maxExp {
		frc = (s.Sig % (1 << 10) * 10) >> 10
		s.Sig >>= 10
		s.Exp += 10
	}
	if s.Exp == 0 {
		return fmt.Sprintf("%d B", s.Sig)
	}
	return fmt.Sprintf("%d.%d %ciB", s.Sig, frc, unitPrefixes[s.Exp/10-1])
}

func Parse(str string) (s Size, err error) {
	valueStr, unitStr := splitValueUnit(str)
	exp, err := parseUnit(unitStr)

	if err != nil {
		return s, err
	}
	sig, err := strconv.ParseFloat(valueStr, 64)

	if err != nil {
		return s, err
	}
	if sig < 0 {
		return s, errors.New("bytesize overflow")
	}
	if sig < 1<<10 && exp > 0 {
		sig *= (1 << 10)
		exp -= 10
	}
	return Size{Num{uint64(sig), exp}}, nil
}

func splitValueUnit(str string) (string, string) {
	i := strings.LastIndexAny(str, "0123456789")

	if i < 0 {
		return "", str
	}
	return str[:i+1], strings.TrimLeftFunc(str[i+1:], unicode.IsSpace)
}

func parseUnit(str string) (uint64, error) {
	switch strings.ToLower(str) {
	case "", "b":
		return bExp, nil
	case "k", "kb", "kib":
		return kibExp, nil
	case "m", "mb", "mib":
		return mibExp, nil
	case "g", "gb", "gib":
		return gibExp, nil
	case "t", "tb", "tib":
		return tibExp, nil
	case "p", "pb", "pib":
		return pibExp, nil
	case "e", "eb", "eib":
		return eibExp, nil
	case "z", "zb", "zib":
		return zibExp, nil
	case "y", "yb", "yib":
		return yibExp, nil
	default:
		return 0, errors.New("unknown unit: " + str)
	}
}
