package main

import (
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"

    "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
    "github.com/hyperledger/fabric-sdk-go/pkg/gateway"

    "bytes"
    "encoding/json"
    "io"
    "strings"

    "github.com/ghodss/yaml"
    "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
    "github.com/spf13/viper"

    "github.com/go-gota/gota/dataframe"
    "time"
)

// Global variable
var (
    sdk         *fabsdk.FabricSDK
    configFile  = "/home/jiale001/bcs/data_synthesis/demo/config/bcs-test-channel-sdk-config.json"
    org         = "4f08db41ded98093a7266580a4a2ae3ce62ce74a"
    sdkfile     *simplejson.Json
    chaincodeID string
    channelID   string
    privateKey  string
)

func main() {
    os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
    wallet, err := gateway.NewFileSystemWallet("wallet")
    if err != nil {
        fmt.Printf("Failed to create wallet: %s\n", err)
        os.Exit(1)
    }

    if !wallet.Exists("admin") {
        err = populateWallet(wallet)
        if err != nil {
            fmt.Printf("Failed to populate wallet contents: %s\n", err)
            os.Exit(1)
        }
    }
    // load config file to config
    loadConfig()
    // initialize sdk
    initializeSdk()

    gw, err := gateway.Connect(
        gateway.WithSDK(sdk),
        gateway.WithIdentity(wallet, "admin"),
    )
    if err != nil {
        fmt.Printf("Failed to connect to gateway: %s\n", err)
        os.Exit(1)
    }
    defer gw.Close()

    network, err := gw.GetNetwork("channel")
    if err != nil {
        fmt.Printf("Failed to get network: %s\n", err)
        os.Exit(1)
    }

    contract := network.GetContract("fisherysc")

    submitData(contract)

}

func submitData(contract *client.Contract) {
    fileDir := "data/single_path_changing_gtin"
    files, err := ioutil.ReadDir(fileDir)
    if err != nil {
        log.Fatal(err)
    }
    for _, f := range files {
        filePath := filepath.Join(fileDir, f.Name())
        f, err := os.Open(filePath)
        if err != nil {
            log.Fatal("Unable to read input file "+filePath, err)
        }
        defer f.Close()
        df := dataframe.ReadCSV(f)
        dmap := df.Maps()
        for i := 0; i <= df.Nrow(); i++ {
            event_type := fmt.Sprint(dmap[i]["event_type"])
            event_time := fmt.Sprint(dmap[i]["event_time"])
            generator_gln := fmt.Sprint(dmap[i]["generator_gln"])
            serial_number := fmt.Sprint(dmap[i]["serial_number"])
            var amount string
            if _, ok := dmap[i]["weight"]; ok {
                amount = fmt.Sprint(dmap[i]["weight"])
            } else {
                amount = fmt.Sprint(dmap[i]["amount"])
            }
            var input_gtin string
            var output_gtin string
            if _, ok := dmap[i]["gtin"]; ok {
                input_gtin = fmt.Sprint(dmap[i]["gtin"])
                output_gtin = fmt.Sprint(dmap[i]["gtin"])
            } else {
                input_gtin = fmt.Sprint(dmap[i]["input_gtin"])
                output_gtin = fmt.Sprint(dmap[i]["output_gtin"])
            }
            args := []string{
                generator_gln,
                event_type,
                input_gtin,
                output_gtin,
                serial_number,
                generator_gln,
                event_time,
                amount,
            }
            createEvent(contract, args)
            time.Sleep(2 * time.Second)
        }
    }
}

// Submit a transaction synchronously, blocking until it has been committed to the ledger.
func addCTEwithAsset(contract *gateway.Contract, args []string) {
    fmt.Printf("Submit Transaction: CreateAsset, creates new asset with ID, Color, Size, Owner and AppraisedValue arguments \n")

    _, err := contract.SubmitTransaction("addCTEwithAsset", args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7])
    if err != nil {
        panic(fmt.Errorf("failed to submit transaction: %w", err))
    }

    fmt.Printf("*** Transaction committed successfully\n")
}

func populateWallet(wallet *gateway.Wallet) error {
    credPath := "/home/jiale001/bcs/data_synthesis/demo/config/4f08db41ded98093a7266580a4a2ae3ce62ce74a.peer/msp"

    certPath := filepath.Join(credPath, "signcerts", "cert.pem")
    // read the certificate pem
    cert, err := ioutil.ReadFile(filepath.Clean(certPath))
    if err != nil {
        return err
    }

    keyDir := filepath.Join(credPath, "keystore")
    // there's a single file in this dir containing the private key
    files, err := ioutil.ReadDir(keyDir)
    if err != nil {
        return err
    }
    if len(files) != 1 {
        return errors.New("keystore folder should have contain one file")
    }
    keyPath := filepath.Join(keyDir, files[0].Name())
    key, err := ioutil.ReadFile(filepath.Clean(keyPath))
    if err != nil {
        return err
    }

    identity := gateway.NewX509Identity("4f08db41ded98093a7266580a4a2ae3ce62ce74aMSP", string(cert), string(key))

    err = wallet.Put("admin", identity)
    if err != nil {
        return err
    }
    // fmt.Println("wallet done")
    return nil
}

func loadConfig() {
    data, _ := ReadFile(configFile)
    data, _ = yaml.YAMLToJSON(data)
    sdkfile, _ = simplejson.NewJson(data)
    //fmt.Println(sdkfile)
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

func formatJSON(data []byte) string {
    var prettyJSON bytes.Buffer
    if err := json.Indent(&prettyJSON, data, " ", ""); err != nil {
        panic(fmt.Errorf("failed to parse JSON: %w", err))
    }
    return prettyJSON.String()
}
