// +build go1.14

package diff

import (
	"hash/maphash"
)

func defaultHash(b []byte) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	h.Write(b)
	return h.Sum64()
}

var seed = maphash.MakeSeed()
