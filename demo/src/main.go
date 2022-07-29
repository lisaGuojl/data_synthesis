package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/ghodss/yaml"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	pmsp "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Global variable
var (
	sdk         *fabsdk.FabricSDK
	configFile  = "/Users/lisagjl/code/bcs/gosdkdemo/config/gosdkdemo-channel-sdk-config.yaml"
	org         = "4f08db41ded98093a7266580a4a2ae3ce62ce74a"
	sdkfile     *simplejson.Json
	chaincodeID string
	channelID   string
	privateKey  string
)

const emptyString = ""

func main() {
	// load config file to config
	loadConfig()
	// initialize sdk
	initializeSdk()
	// insert data 
	// ReadAsset("ReadAsset")
	addEventwithAsset("AddCTEwithAsset")
	// ReadAsset("ReadAsset")

}

// loadConfig load the config file to initialize some Global variable
func loadConfig() {
	data, _ := ReadFile(configFile)
	data, _ = yaml.YAMLToJSON(data)
	sdkfile, _ = simplejson.NewJson(data)
	fmt.Println(sdkfile)
	channelID = GetDefaultChannel()
	chaincodeID = GetDefaultChaincodeId()

}

// InitializeSdk initialize the sdk
func initializeSdk() {
	cnfg := config.FromFile(configFile)
	configProvider := cnfg
	var opts []fabsdk.Option
	opts, err := getOptsToInitializeSDK(configFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}

	/*
		We could use function GetTlsCryptoKey to read tls key file(encrypted or not) from path specified by config file.
		For example:
		tlsKey,err := GetTlsCryptoKey(org)
		if err !=nil{
			panic(fmt.Sprintf("Failed to get TlsCryptoKey: %s", err))

		}
		If such tls key is encrypted, you must use SetClientTlsKey to update the tlskey in fabric-sdk after decrypting it.
		Or it will cause decode error.
		SetClientTlsKey(decryptedTlsKey)
		After that if we need to reset tls key specified by config file, use function ClearClientTlsKey please.
	*/

	sdk, err = fabsdk.New(configProvider, opts...)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}
}

// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	var buf bytes.Buffer
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if errors, ok := e.(error); ok && errors == bytes.ErrTooLarge {
			err = errors
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

// ReadFile reads the file named by filename and returns the contents.
func ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// It's a good but not certain bet that FileInfo will tell us exactly how much to
	// read, so let's try it but be prepared for the answer to be wrong.
	var n int64 = bytes.MinRead

	if fi, err := f.Stat(); err == nil {
		if size := fi.Size() + bytes.MinRead; size > n {
			n = size
		}
	}
	return readAll(f, n)
}

// GetDefaultChaincodeId is a funtion to get the default chaincodeId
func GetDefaultChaincodeId() string {
	chaincodes := sdkfile.Get("channels").Get(channelID).Get("chaincodes").MustArray()
	if str, ok := chaincodes[0].(string); ok {
		return strings.Split(str, ":")[0]
	}
	return ""
}

// GetDefaultChannel is a funtion to get the default Channel
func GetDefaultChannel() string {
	channels := sdkfile.Get("channels").MustMap()
	for k, _ := range channels {
		return k
	}
	return ""
}

//GetCryptoPath get msp directory from sdk config file's path
func GetCryptoPath(orgId string) string {
	cryptoPath := sdkfile.Get("organizations").Get(orgId).Get("cryptoPath").MustString()
	return cryptoPath
}

//GetTlsCryptoKeyPath get tlsCryptoKeyPath from sdk config file's path with orgId
func GetTlsCryptoKeyPath(orgId string) string {
	tlsCryptoKeyPath := sdkfile.Get("organizations").Get(orgId).Get("tlsCryptoKeyPath").MustString()
	return tlsCryptoKeyPath
}

//GetSigncertsBytes can get Signcerts from sdk config file's path
//cryptoPath is get from function GetCryptoPath
func GetSigncertsBytes(orgId string) ([]byte, error) {
	cryptoPath := GetCryptoPath(orgId)
	signcertsPathdir := filepath.Join(cryptoPath, "signcerts")
	files, _ := ioutil.ReadDir(signcertsPathdir)
	if len(files) != 1 {
		return nil, errors.Errorf("file count invalid in the directory [%s]", signcertsPathdir)
	}

	f, err := ioutil.ReadFile(filepath.Join(signcertsPathdir, files[0].Name()))
	if err != nil {
		return nil, errors.Errorf("read signcerts from [%s] fail", files[0].Name())
	} else if f == nil {
		return nil, errors.Errorf("result of read signcerts file [%s] is null", files[0].Name())
	}

	return f, nil
}

//GetPrivateKeyBytes can get privateKey from sdk config file's path
//cryptoPath is get from function GetCryptoPath
func GetPrivateKeyBytes(orgId string) ([]byte, error) {
	cryptoPath := GetCryptoPath(orgId)
	keystorePathdir := filepath.Join(cryptoPath, "keystore")
	files, _ := ioutil.ReadDir(keystorePathdir)
	if len(files) != 1 {
		return nil, errors.Errorf("file count invalid in the directory [%s]", keystorePathdir)
	}

	f, err := ioutil.ReadFile(filepath.Join(keystorePathdir, files[0].Name()))
	if err != nil {
		return nil, errors.Errorf("read signcerts from [%s] fail", files[0].Name())
	} else if f == nil {
		return nil, errors.Errorf("result of read keystore file [%s] is null", files[0].Name())
	}

	return f, nil
}

//GetTlsCryptoKey can get tlsCryptoKey content from tlsCryptoKeyPath
//tlsCryptoKeyPath is get from function GetTlsCryptoKeyPath
func GetTlsCryptoKey(orgId string) (string, error) {
	tlsCryptoKeyPath := GetTlsCryptoKeyPath(orgId)
	f, err := ioutil.ReadFile(tlsCryptoKeyPath)
	if err != nil {
		return "", errors.Errorf("read tlsCryptoKey from [%s] fail", tlsCryptoKeyPath)
	} else if f == nil {
		return "", errors.Errorf("result of read tlsCryptoKey file [%s] is null", tlsCryptoKeyPath)
	}

	return string(f), nil
}

// getOptsToInitializeSDK is a function to initialize SDK
func getOptsToInitializeSDK(configFile string) ([]fabsdk.Option, error) {
	var opts []fabsdk.Option

	vc := viper.New()
	vc.SetConfigFile(configFile)
	err := vc.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}

	org := vc.GetString("client.originalOrganization")
	if org == "" {
		org = vc.GetString("client.organization")
	}

	opts = append(opts, fabsdk.WithOrgid(org))

	opts = append(opts, fabsdk.WithUserName("Admin"))
	return opts, nil
}

// ChannelClient creates a new channel client
func ChannelClient(channelID string, user mspapi.SigningIdentity) (*channel.Client, error) {
	session := sdk.Context(fabsdk.WithIdentity(user))
	contextImpl.NewChannel(session, channelID)
	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(session, channelID)
	}
	return channel.New(channelProvider)
}

// UserIdentityWithOrgAndName Identify users through org and Name,
// which can pass cert and private key with external variables
// if cert or pvtKey is empty, it will read from file path in sdk config file
func UserIdentityWithOrgAndName(orgID string, userName string, cert []byte, pvtKey string) (mspapi.SigningIdentity, error) {
	if userName == "" {
		return nil, errors.Errorf("No username specified")
	}

	mspClient, err := msp.New(sdk.Context(), msp.WithOrg(orgID))
	if err != nil {
		return nil, errors.Errorf("Error creating MSP client: %s", err)
	}

	if pvtKey == emptyString {
		user, err := mspClient.GetSigningIdentity(userName)
		if err != nil {
			return nil, errors.Errorf("GetSigningIdentity returned error: %v", err)
		}
		return user, nil

	} else {
		if cert == nil {
			if cert, err = GetSigncertsBytes(orgID); err != nil {
				return nil, err
			}
		}

		// pvtKey must be decrypted when it is passed in function CreateSigningIdentity.
		user, err := mspClient.CreateSigningIdentity(pmsp.WithCert(cert), pmsp.WithPrivateKey([]byte(pvtKey)))
		if err != nil {
			return nil, errors.Errorf("CreateSigningIdentity returned error: %v", err)
		}
		return user, nil
	}
	return nil, nil
}


func addEventwithAsset(fName string) (channel.Response, error) {
	//We can get private key(encrypted or not) using function GetPrivateKeyBytes
	//We recommend that you set the private key to be decrypted here using function SetPrivateKey。
	//for example:
	//encryptedbytekey,_:= GetPrivateKeyBytes(org)
	//decrypt the encryptedkey
	//SetPrivateKey(decryptedKey)
	//if cert or privateKey is empty, anyone of them will read from file path in sdk config file
	user, err := UserIdentityWithOrgAndName(org, "Admin", nil, privateKey)
	if err != nil {
		panic(err)
	}

	//Generating the Channel Context
	chClient, err := ChannelClient(channelID, user)
	if err != nil {
		panic(err)
	}

	//The client need to send a requset to the channel Chaincode Name,Function Name, Parameter
	args := [][]byte{
		[]byte("asset1"),
		[]byte("1"),
		[]byte("62567598498627"),
		[]byte("62567598498627"),
		[]byte("3604604142873"),
		[]byte("107HAXI"),
		[]byte("2022-Jul-24T08:43:08 +0000"),
		[]byte("10"),
	}
	response, err := chClient.Execute(
		channel.Request{
			ChaincodeID: chaincodeID,
			Fcn:         fName,
			Args:        args,      
		})
	if err != nil {
		fmt.Println("insert faild", err)
		panic(err)
	}
	fmt.Printf("store new data success\n")
	fmt.Printf("response: ", response, "\n")
	return response, nil
}

// Query Function
func ReadAsset(fName string) (channel.Response, error) {
	//We can get private key(encrypted or not) using function GetPrivateKeyBytes
	//We recommend that you set the private key to be decrypted here using function SetPrivateKey。
	//for example:
	//encryptedkey,_:= GetPrivateKeyBytes(org)
	//decrypt the encryptedkey
	//SetPrivateKey(decryptedKey)
	//if cert or privateKey is empty, anyone of them will read from file path in sdk config file
	user, err := UserIdentityWithOrgAndName(org, "Admin", nil, privateKey)
	if err != nil {
		panic(err)
	}

	chClient, err := ChannelClient(channelID, user)
	if err != nil {
		panic(err)
	}

	// The client need to send a requset to the channel Chaincode Name,Function Name, Parameter
	args := [][]byte{
			[]byte("asset1")}
	queryRes, err := chClient.Query(channel.Request{
		ChaincodeID: chaincodeID,
		Fcn:         fName,
		Args:        args,
	})
	fmt.Printf("query key <%s> value is %s\n", string(args[0]), string(queryRes.Payload))
	return queryRes, err
}



//SetPrivateKey update the privateKey used in ChannelClient
func SetPrivateKey(key string) {
	privateKey = key
}

//ClearPrivateKey set privateKey empty
func ClearPrivateKey() {
	privateKey = ""
}

//SetClientTlsKey update the tls key in fabric-sdk
func SetClientTlsKey(tlsKey string) {
	vc := viper.New()
	vc.SetConfigFile(configFile)
	err := vc.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to read configFile: %s", configFile))
	}

	orgID := vc.GetString("client.originalOrganization")
	if orgID == "" {
		orgID = vc.GetString("client.organization")
	}

	//SetTlsClientKey can be used to update the tlskey in fabric-sdk
	fab.SetTlsClientKey(orgID, tlsKey)
}

//ClearClientTlsKey reset the tls key in fabric-sdk with tls key file specified in config
func ClearClientTlsKey() {
	vc := viper.New()
	vc.SetConfigFile(configFile)
	err := vc.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to read configFile: %s", configFile))
	}

	orgID := vc.GetString("client.originalOrganization")
	if orgID == "" {
		orgID = vc.GetString("client.organization")
	}
	fab.ResetTlsClientKeyWithOrgID(orgID)
}
