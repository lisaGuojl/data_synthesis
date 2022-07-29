package apiKey

import (
	"errors"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/plugins/huawei/hbccsp/hutil"
)

func SetUserPassword(password string) error {
	hutil.UserPassword = password
	return nil
}

func NewsecureKeyClient() (*hutil.SecureUserKey, error) {
	secureKeyClient := new(hutil.SecureUserKey)
	if secureKeyClient == nil {
		return nil, errors.New("fail to get secure key client")
	}
	return secureKeyClient, nil
}
