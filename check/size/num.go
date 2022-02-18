package size

import (
	"fmt"
	"math/bits"
	"strconv"
)

type Num struct {
	Sig uint64
	Exp uint64
}

const maxUint64 = (1 << 64) - 1

func N(n uint64) Num {
	return Num{n, 0}
}

func (n Num) Mul(m Num) Num {
	if n.Sig == 0 || m.Sig == 0 {
		return N(0)
	}
	shiftRight(&n)
	shiftRight(&m)

	for n.Sig > maxUint64/m.Sig {
		if n.Sig > m.Sig {
			n.Sig >>= 1
			n.Exp += 1
		} else {
			m.Sig >>= 1
			m.Exp += 1
		}
	}
	return Num{n.Sig * m.Sig, n.Exp + m.Exp}
}

func (n Num) Div(m Num) Num {
	if m.Sig == 0 {
		panic("division by zero")
	}
	if n.Sig == 0 {
		return N(0)
	}
	shiftLeft(&n)
	shiftLeft(&m)

	for n.Exp < m.Exp && m.Sig < maxUint64>>1 {
		m.Sig <<= 1
		m.Exp -= 1
	}
	for n.Exp < m.Exp && n.Sig > 0 {
		n.Sig >>= 1
		n.Exp += 1
	}
	if n.Sig == 0 {
		return N(0)
	}
	return Num{n.Sig / m.Sig, n.Exp - m.Exp}
}

func (n Num) CompareTo(m Num) int {
	shiftLeft(&n)
	shiftLeft(&m)

	if n.Exp < m.Exp {
		return -1
	}
	if n.Exp > m.Exp {
		return 1
	}
	if n.Sig < m.Sig {
		return -1
	}
	if n.Sig > m.Sig {
		return 1
	}
	return 0
}

func (n Num) String() string {
	shiftLeft(&n)

	if n.Exp == 0 {
		return strconv.FormatUint(n.Sig, 10)
	}
	return fmt.Sprintf("%d₂%d", n.Sig, n.Exp)
}

func shiftLeft(n *Num) {
	s := uint64(bits.LeadingZeros64(n.Sig))

	if s >= n.Exp {
		n.Sig <<= n.Exp
		n.Exp = 0
	} else {
		n.Sig <<= s
		n.Exp -= s
	}
}

func shiftRight(n *Num) {
	s := uint64(bits.TrailingZeros64(n.Sig))
	n.Sig >>= s
	n.Exp += s
}
