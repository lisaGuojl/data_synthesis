package hsm

import (
	"hash"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
)

const (
	// GM hash identifier
	SM3 = "SM3"
)

type SM3Opts struct {
}

// Algorithm returns the hash algorithm identifier (to be used).
func (opts *SM3Opts) Algorithm() string {
	return "SM3"
}

type hasher struct {
	hash func() hash.Hash
}

func (c *hasher) Hash(msg []byte, opts bccsp.HashOpts) (hash []byte, err error) {
	h := c.hash()
	h.Write(msg)
	return h.Sum(nil), nil
}

func (c *hasher) GetHash(opts bccsp.HashOpts) (h hash.Hash, err error) {
	return c.hash(), nil
}
