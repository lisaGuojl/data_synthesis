package hsm

import (
	"crypto/ecdsa"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
)

type sm2KeyGenerator struct {
}

func (kg *sm2KeyGenerator) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	key, err := sm2.GenerateKey()
	if err != nil {
		return nil, errors.WithMessage(err, "generate key fail")
	}
	var privateKey = &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: key.PublicKey.Curve,
			X:     key.PublicKey.X,
			Y:     key.PublicKey.Y,
		},
		D: key.D,
	}
	return &sm2PrivateKey{privateKey}, nil
}
