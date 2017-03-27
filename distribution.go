package main

import (
	"crypto/rand"
	"math/big"
)

type UniformIntDistribution struct {
	min int
	max int
}

// maybe expensive for min > 0 :)
func (u *UniformIntDistribution) Gen() int {
	var r int64
	if u.min == u.max { // occurs if a random number result is used which is already zero
		return u.max
	}
	big := big.NewInt(int64(u.max))
	for {
		nBig, err := rand.Int(rand.Reader, big)
		if err != nil {
			panic(err)
		}
		r = nBig.Int64()
		if r >= int64(u.min) {
			break
		}
	}
	return int(r)
}
