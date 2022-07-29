package hx509

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/plugins/huawei/hbccsp/internal/hverify"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/plugins/huawei/hbccsp/internal/sx509"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/pkg/errors"
	"reflect"
	"strings"
	"unsafe"
)

var logger = flogging.MustGetLogger("hx509")

func ParseCertificate(asn1Data []byte) (*x509.Certificate, error) {
	return sx509.ParseCertificate(asn1Data)
}

func Verify(cert *x509.Certificate, opts x509.VerifyOptions) (chains [][]*x509.Certificate, err error) {
	if val, err := cert.Verify(opts); err == nil {
		return val, nil
	}
	validationChains, err := (*hverify.SMCertificate)(unsafe.Pointer(cert)).SMVerify(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "verify with sm cert fail")
	}
	return validationChains, nil
}

func WrapHashResult(bsp bccsp.BCCSP, msg []byte, digest []byte) []byte {
	var val = isBccspOfSM(bsp)
	if val {
		return msg
	}
	return digest
}

func isBccspOfSM(bsp bccsp.BCCSP) bool {
	if val := reflect.TypeOf(bsp).String(); strings.Contains(val, "sm") {
		return true
	}
	return false
}

// PEMtoPrivateKey unmarshals a pem to private key
func PEMtoPrivateKey(raw []byte, pwd []byte) (interface{}, error) {
	if len(raw) == 0 {
		return nil, errors.New("Invalid PEM. It must be different from nil.")
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("Failed decoding PEM. Block must be different from nil.")
	}

	// TODO: derive from header the type of the key

	if x509.IsEncryptedPEMBlock(block) {
		if len(pwd) == 0 {
			return nil, errors.New("Encrypted Key. Need a password")
		}

		decrypted, err := x509.DecryptPEMBlock(block, pwd)
		if err != nil {
			return nil, fmt.Errorf("Failed PEM decryption [%s]", err)
		}

		key, err := sx509.DERToPrivateKey(decrypted)
		if err != nil {
			return nil, err
		}
		return key, nil
	}

	key, err := sx509.DERToPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func PrivateKeyToDER(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("Invalid sm2 private key. It must be different from nil.")
	}

	return sx509.MarshalECPrivateKey(privateKey)
}
