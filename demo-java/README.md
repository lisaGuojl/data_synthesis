# java demo


## Requirements 
> * java
> * maven

## 1. Downloading SDK configurations and certificates
> * (1). Log in to the BCS console.
> * (2). On the Service Management page, click "Download Client Configureation" in a service card.
> * (3). Select "SDK Configuration File", "Chaincode Name" is fisherysc, the "Certificate Path" should be the full path where those files will be stored for this demo, e.g. /root/gosdkdemo/config
> * (4). Select Orderer Certificate. Select Peer Certificates, select organization for Peer Organization, and select Administrator certificate.
> * (5). Click Download to download the SDK configuration file and the administrator certificates for the demo-orderer and organization organizations.
> * (6). Decompress demo-config.zip and copy the orderer and peer folders and the sdk-config.json and sdk-config.yaml files to the config directory where the demo is stored, e.g. /root/gosdkdemo/config. 


## 2、Run
```
mvn package
java -cd ./target/gatewayjavademo-1.0-SNAPSHOT-jar-with-dependencies.jar handler.Main
```
