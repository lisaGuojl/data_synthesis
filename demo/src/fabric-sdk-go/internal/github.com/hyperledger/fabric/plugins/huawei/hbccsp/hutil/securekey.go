package hutil

import (
	"errors"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/plugins/huawei/hbccsp/internal/sx509"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"io/ioutil"
	"log"
	"os"
)

var logger = flogging.MustGetLogger("securekey")

type SecureUserKey struct{}

var UserPassword = ""

func (suk *SecureUserKey) EncryptKey(PlainMSPKeyFilePath, PlainTLSKeyFilePath, EncryptedMSPKeyFilePath, EncryptedTLSKeyFilePath, password, keyType string) error {
	// check parameters
	if PlainMSPKeyFilePath == "" || PlainTLSKeyFilePath == "" || EncryptedMSPKeyFilePath == "" || EncryptedTLSKeyFilePath == "" || password == "" || keyType == "" {
		logger.Fatal("input parameters is empty")
		return errors.New("input parameters is empty")
	}

	err := CheckPass(password)
	if err != nil {
		logger.Errorf("invalid password: %s", err.Error())
		return errors.New("invalid password: " + err.Error())
	}

	if keyType != "sm" && keyType != "sw" {
		logger.Errorf("unsupported key type, should be sm or sw")
		return errors.New("uunsupported key type, should be sm or sw")
	}

	// handle msp private key
	PlainMSPpKeyRaw, err := readLocalInfo(PlainMSPKeyFilePath)
	if err != nil {
		logger.Error("fail to read msp key file: ", PlainMSPKeyFilePath)
		return err
	}
	if keyType == "sm" {
		PlainMSPKeyCont, err := sx509.PEMtoPrivateKey([]byte(PlainMSPpKeyRaw), nil)
		if err != nil {
			return err
		}
		EncryptedMSPKeyRaw, err := sx509.PrivateKeyToPEM(PlainMSPKeyCont, []byte(password))
		if err != nil {
			return err
		}

		err = writeLocalInfo(EncryptedMSPKeyFilePath, string(EncryptedMSPKeyRaw))
		if err != nil {
			logger.Error("fail to write encrypted key file: ", EncryptedMSPKeyFilePath)
			return err
		}
	} else if keyType == "sw" {
		PlainMSPKeyCont, err := sx509.PEMtoPrivateKey([]byte(PlainMSPpKeyRaw), nil)
		if err != nil {
			return err
		}
		EncryptedMSPKeyRaw, err := sx509.PrivateKeyToPEM(PlainMSPKeyCont, []byte(password))
		if err != nil {
			return err
		}

		err = writeLocalInfo(EncryptedMSPKeyFilePath, string(EncryptedMSPKeyRaw))
		if err != nil {
			logger.Error("fail to write encrypted key file: ", EncryptedMSPKeyFilePath)
			return err
		}
	}

	// handle tls private key, note for now tls private key only support ECDSA
	PlainTLSKeyRaw, err := readLocalInfo(PlainTLSKeyFilePath)
	if err != nil {
		logger.Error("fail to read tls key file: ", PlainTLSKeyFilePath)
		return err
	}

	PlainTLSKeyCont, err := sx509.PEMtoPrivateKey([]byte(PlainTLSKeyRaw), nil)
	if err != nil {
		return err
	}
	EncryptedTLSKeyRaw, err := sx509.PrivateKeyToPEM(PlainTLSKeyCont, []byte(password))
	if err != nil {
		return err
	}

	err = writeLocalInfo(EncryptedTLSKeyFilePath, string(EncryptedTLSKeyRaw))
	if err != nil {
		logger.Error("fail to write encrypted key file: ", EncryptedTLSKeyFilePath)
		return err
	}

	return nil
}

func readLocalInfo(file string) (string, error) {

	_, err := os.Stat(file)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
		return "", err
	}
	if os.IsNotExist(err) {

		return "", err
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}

	return string(data), err
}

func writeLocalInfo(file string, data string) error {

	fileInfo, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {

		log.Fatalln(err)
		return err
	}
	_, err = fileInfo.WriteString(data)
	if err != nil {

		log.Fatalln(err)
		return err
	}
	return err
}
