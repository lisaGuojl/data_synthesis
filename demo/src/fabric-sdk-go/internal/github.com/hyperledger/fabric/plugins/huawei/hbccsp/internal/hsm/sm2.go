package hsm

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"errors"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/tjfoc/gmsm/sm2"
	"math/big"
)

type ECDSASignature struct {
	R, S *big.Int
}

func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ECDSASignature{r, s})
}

const userID = "1234567812345678"

type sm2Signer struct{}

func signSM2(k *ecdsa.PrivateKey, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	sm2.GenerateKey()
	var privateKey = &sm2.PrivateKey{
		PublicKey: sm2.PublicKey{
			Curve: k.PublicKey.Curve,
			X:     k.PublicKey.X,
			Y:     k.PublicKey.Y,
		},
		D: k.D,
	}
	r, s, err := sm2.Sm2Sign(privateKey, digest, []byte(userID))
	if err != nil {
		return nil, errors.New("Sm2Sign fail")
	}
	return MarshalECDSASignature(r, s)
}

func (s *sm2Signer) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	return signSM2(k.(*sm2PrivateKey).privKey, digest, opts)
}

type sm2PublicKeyKeyVerifier struct{}

func (v *sm2PublicKeyKeyVerifier) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	return VerifySM2(k.(*sm2PublicKey).pubKey, signature, digest, nil)
}

func VerifySM2(k *ecdsa.PublicKey, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	ecdsaSignature := &ECDSASignature{}
	asn1.Unmarshal(signature, ecdsaSignature)
	var publicKey = &sm2.PublicKey{
		Curve: k.Curve,
		X:     k.X,
		Y:     k.Y,
	}
	valid = sm2.Sm2Verify(publicKey, digest, []byte(userID), ecdsaSignature.R, ecdsaSignature.S)
	return valid, nil
}
