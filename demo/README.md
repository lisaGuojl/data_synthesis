# demo


## Requirements 
> * fabric v1.1 v1.4 v2.2
> * fabrci-go-sdk v1.0.0
> * go >=1.12，<1.16

## 1. Downloading SDK configurations and certificates
> * (1). Log in to the BCS console.
> * (2). On the Service Management page, click "Download Client Configureation" in a service card.
> * (3). Select "SDK Configuration File", "Chaincode Name" is fisherysc, the "Certificate Path" should be the full path where those files will be stored for this demo, e.g. /root/gosdkdemo/config
> * (4). Select Orderer Certificate. Select Peer Certificates, select organization for Peer Organization, and select Administrator certificate.
> * (5). Click Download to download the SDK configuration file and the administrator certificates for the demo-orderer and organization organizations.
> * (6). Decompress demo-config.zip and copy the orderer and peer folders and the sdk-config.json and sdk-config.yaml files to the config directory where the demo is stored, e.g. /root/gosdkdemo/config. 
> * (7). Modify the the value of the variable "configFile" in main.go accoringly. 

## 2、
	fabric-go-demo
        --chaincode/chaincode.go  the current version of chaincode installed on BCS
        --src/main.go  demo code to submit transactions
        --src/fabric-sdk-go  fabric-sdk-go v1.0.0

## 3、Run
```
cd src
go mod tidy
go run .
```
