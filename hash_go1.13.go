// +build !go1.14

package diff

import (
	"hash/fnv"
	"math/rand"
	"os"
	"time"
)

func defaultHash(b []byte) uint64 {
	h := fnv.New64a()
	h.Write(seed)
	h.Write(b)
	return h.Sum64()
}

var seed = make([]byte, 8)

func init() {
	s := time.Now().UnixNano() + int64(os.Getpid())
	rnd := rand.New(rand.NewSource(s))
	rnd.Read(seed)
}
