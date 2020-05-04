package crypto

import (
	"math/big"
	"math/rand"
	"time"
)

// Split Diffie-Hellman PQ
func SplitPQ(pq *big.Int) (p1, p2 *big.Int) {
	value0 := big.NewInt(0)
	value1 := big.NewInt(1)
	value15 := big.NewInt(15)
	value17 := big.NewInt(17)
	rndMax := big.NewInt(0).SetBit(big.NewInt(0), 64, 1)

	what := big.NewInt(0).Set(pq)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := big.NewInt(0)
	i := 0
	for !(g.Cmp(value1) == 1 && g.Cmp(what) == -1) {
		q := big.NewInt(0).Rand(rnd, rndMax)
		q = q.And(q, value15)
		q = q.Add(q, value17)
		q = q.Mod(q, what)

		x := big.NewInt(0).Rand(rnd, rndMax)
		whatnext := big.NewInt(0).Sub(what, value1)
		x = x.Mod(x, whatnext)
		x = x.Add(x, value1)

		y := big.NewInt(0).Set(x)
		lim := 1 << (uint(i) + 18)
		j := 1
		flag := true

		for j < lim && flag {
			a := big.NewInt(0).Set(x)
			b := big.NewInt(0).Set(x)
			c := big.NewInt(0).Set(q)

			for b.Cmp(value0) == 1 {
				b2 := big.NewInt(0)
				if b2.And(b, value1).Cmp(value0) == 1 {
					c.Add(c, a)
					if c.Cmp(what) >= 0 {
						c.Sub(c, what)
					}
				}
				a.Add(a, a)
				if a.Cmp(what) >= 0 {
					a.Sub(a, what)
				}
				b.Rsh(b, 1)
			}
			x.Set(c)

			z := big.NewInt(0)
			if x.Cmp(y) == -1 {
				z.Add(what, x)
				z.Sub(z, y)
			} else {
				z.Sub(x, y)
			}
			g.GCD(nil, nil, z, what)

			if (j & (j - 1)) == 0 {
				y.Set(x)
			}
			j = j + 1

			if g.Cmp(value1) != 0 {
				flag = false
			}
		}
		i = i + 1
	}

	p1 = big.NewInt(0).Set(g)
	p2 = big.NewInt(0).Div(what, g)

	if p1.Cmp(p2) == 1 {
		p1, p2 = p2, p1
	}

	return
}
