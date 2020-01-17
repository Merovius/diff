// +build !go1.14

package diff

import (
	"hash/fnv"
	"math/rand"
	"os"
	"time"
)

func defaultHash() HashFunc {
	seed := make([]byte, 8)
	s := time.Now().UnixNano() + int64(os.Getpid())
	rnd := rand.New(rand.NewSource(s))
	rnd.Read(seed)
	return func(b []byte) uint64 {
		h := fnv.New64a()
		h.Write(seed)
		h.Write(b)
		return h.Sum64()
	}
}
