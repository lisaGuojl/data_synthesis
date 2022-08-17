package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically

// Asset struct is used to store credits of stakeholders
// Currently we use generator_gln as its ID
// Value is the credit value
type Asset struct {
	ID    string `json:"ID"`
	Owner string `json:"Owner"`
	Value int    `json:"Value"`
}

// CTE data
type Event struct {
	EventId      string `json:"event_id"`
	EventType    int    `json:"event_type"`
	InputGtin    string `json:"input_gtin"`
	OutputGtin   string `json:"output_gtin"`
	SerialNumber string `json:"serial_number"`
	EventTime    string `json:"event_time"`
	EventLoc     string `json:"event_loc"`
	LocationName string `json:"location_name"`
	CompanyName  string `json:"company_name"`
}

type TxInfo struct {
	Txid      string `json:"Txid"`
	Timestamp string `json:"Timestamp"`
}

func str2slice(str string) []string {
	s1 := strings.Replace(str, "[", "", -1)
	s2 := strings.Replace(s1, "]", "", -1)
	s := strings.Split(s2, ",")
	return s
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	// just for test
	assets := []Asset{
		{ID: "asset1", Owner: "FishingCompany", Value: 0},
		{ID: "asset2", Owner: "AuctionCenter", Value: 0},
		{ID: "asset3", Owner: "LogisticServiceProvider", Value: 0},
		{ID: "asset4", Owner: "ProcessingCompany", Value: 0},
		{ID: "asset5", Owner: "Wholesaler", Value: 0},
		{ID: "asset6", Owner: "Retailer", Value: 0},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new Asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, owner string, Value int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists != nil {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:    id,
		Owner: owner,
		Value: 0,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	// PutState: put (key, value) into tx's writeset, here key is id, value is assetJSON.
	return ctx.GetStub().PutState(id, assetJSON)
}

// AddCTEwithAsset record event data and increase the generator's credit
// prekey: the key to retrieve the previous CTE data stored in world state
// newkey: the key used to store this new CTE data. newkey is usually same as the prekey expect on a merge or split point of a path.
// id: the key to get generator's Asset info (in order to add credit)
// the rest params will be stored as Event
func (s *SmartContract) AddCTEwithAsset(ctx contractapi.TransactionContextInterface, prekey string, newkey string, id string, eventid string, eventtype int, input_gtin string, output_gtin string, serialnumber string, time string, loc string, locationname string, companyname string) (string, error) {
	event := Event{
		EventId:      eventid,
		EventType:    eventtype,
		InputGtin:    input_gtin,
		OutputGtin:   output_gtin,
		SerialNumber: serialnumber,
		EventTime:    time,
		EventLoc:     loc,
		LocationName: locationname,
		CompanyName:  companyname,
	}
	neweventJSON, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	prekeyList := str2slice(prekey)
	for _, k := range prekeyList {
		// Get previous CTE data
		_, err = ctx.GetStub().GetState(k)
		if err != nil {
			return "", fmt.Errorf("failed to get the previous transaction: %v", err)
		}
	}
	// Put neweventJSON into tx's writeset using the key=newkey
	ctx.GetStub().PutState(newkey, neweventJSON)

	txid := ctx.GetStub().GetTxID()
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", err
	}
	txinfo := TxInfo{txid, strconv.FormatInt(timestamp.GetSeconds(), 10)}
	resJson, err := json.Marshal(txinfo)
	if err != nil {
		return "", err
	}

	err = s.AddCoin(ctx, id)
	if err != nil {
		return "", err
	}
	//return ctx.GetStub().PutState(OutputGTIN, eventJSON)
	//return ctx.GetStub().setEvent("addCTE", eventJSON)
	return string(resJson), nil
}

// ReadAsset returns the Asset data stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	asset := new(Asset)
	_ = json.Unmarshal(assetJSON, asset)

	return asset, nil
}

// AddCoin add a coin to an existing asset in the world state with provided parameters.
func (s *SmartContract) AddCoin(ctx contractapi.TransactionContextInterface, id string) error {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	asset := new(Asset)
	if assetJSON == nil {
		asset.ID = id
		asset.Owner = ""
		asset.Value = 0
	} else {
		_ = json.Unmarshal(assetJSON, asset)
	}

	asset.Value = asset.Value + 1
	assetJSON, err = json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldOwner := asset.Owner
	asset.Owner = newOwner

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return oldOwner, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fishery chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fishery chaincode: %s", err.Error())
	}
}
