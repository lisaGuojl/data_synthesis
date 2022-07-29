package hsm

import (
	"fmt"

	"errors"

	"crypto/ecdsa"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"math/big"
)

type sm2PublicKeyKeyDeriver struct{}

func (kd *sm2PublicKeyKeyDeriver) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	// Validate opts
	if opts == nil {
		return nil, errors.New("Invalid opts parameter. It must not be nil.")
	}

	sm2K := k.(*sm2PublicKey)

	switch opts.(type) {
	// Re-randomized an SM2 private key
	case *SM2ReRandKeyOpts:
		reRandOpts := opts.(*SM2ReRandKeyOpts)
		tempSK := &ecdsa.PublicKey{
			Curve: sm2K.pubKey.Curve,
			X:     new(big.Int),
			Y:     new(big.Int),
		}

		var k = new(big.Int).SetBytes(reRandOpts.ExpansionValue())
		var one = new(big.Int).SetInt64(1)
		n := new(big.Int).Sub(sm2K.pubKey.Params().N, one)
		k.Mod(k, n)
		k.Add(k, one)

		// Compute temporary public key
		tempX, tempY := sm2K.pubKey.ScalarBaseMult(k.Bytes())
		tempSK.X, tempSK.Y = tempSK.Add(
			sm2K.pubKey.X, sm2K.pubKey.Y,
			tempX, tempY,
		)

		// Verify temporary public key is a valid point on the reference curve
		isOn := tempSK.Curve.IsOnCurve(tempSK.X, tempSK.Y)
		if !isOn {
			return nil, errors.New("Failed temporary public key IsOnCurve check.")
		}

		return &sm2PublicKey{tempSK}, nil
	default:
		return nil, fmt.Errorf("Unsupported 'KeyDerivOpts' provided [%v]", opts)
	}
}

type sm2PrivateKeyKeyDeriver struct{}

func (kd *sm2PrivateKeyKeyDeriver) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	// Validate opts
	if opts == nil {
		return nil, errors.New("Invalid opts parameter. It must not be nil.")
	}

	sm2K := k.(*sm2PrivateKey)

	switch opts.(type) {
	// Re-randomized an SM2 private key
	case *SM2ReRandKeyOpts:
		reRandOpts := opts.(*SM2ReRandKeyOpts)
		tempSK := &ecdsa.PrivateKey{
			PublicKey: ecdsa.PublicKey{
				Curve: sm2K.privKey.Curve,
				X:     new(big.Int),
				Y:     new(big.Int),
			},
			D: new(big.Int),
		}

		var k = new(big.Int).SetBytes(reRandOpts.ExpansionValue())
		var one = new(big.Int).SetInt64(1)
		n := new(big.Int).Sub(sm2K.privKey.Params().N, one)
		k.Mod(k, n)
		k.Add(k, one)

		tempSK.D.Add(sm2K.privKey.D, k)
		tempSK.D.Mod(tempSK.D, sm2K.privKey.PublicKey.Params().N)

		// Compute temporary public key
		tempX, tempY := sm2K.privKey.PublicKey.ScalarBaseMult(k.Bytes())
		tempSK.PublicKey.X, tempSK.PublicKey.Y =
			tempSK.PublicKey.Add(
				sm2K.privKey.PublicKey.X, sm2K.privKey.PublicKey.Y,
				tempX, tempY,
			)

		// Verify temporary public key is a valid point on the reference curve
		isOn := tempSK.Curve.IsOnCurve(tempSK.PublicKey.X, tempSK.PublicKey.Y)
		if !isOn {
			return nil, errors.New("Failed temporary public key IsOnCurve check.")
		}

		return &sm2PrivateKey{tempSK}, nil
	default:
		return nil, fmt.Errorf("Unsupported 'KeyDerivOpts' provided [%v]", opts)
	}
}
