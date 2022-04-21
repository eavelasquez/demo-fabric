/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// MyPrivateAssetContract contract for managing CRUD for MyPrivateAsset
type MyPrivateAssetContract struct {
	contractapi.Contract
}

func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {
	mspid, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", err
	}

	collectionName := "_implicit_org_" + mspid

	return collectionName, nil
}

// MyPrivateAssetExists returns true when asset with given ID exists in private data collection
func (c *MyPrivateAssetContract) MyPrivateAssetExists(ctx contractapi.TransactionContextInterface, myPrivateAssetID string) (bool, error) {
	collectionName, collectionNameErr := getCollectionName(ctx)
	if collectionNameErr != nil {
		return false, collectionNameErr
	}

	data, err := ctx.GetStub().GetPrivateDataHash(collectionName, myPrivateAssetID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateMyPrivateAsset creates a new instance of MyPrivateAsset
func (c *MyPrivateAssetContract) CreateMyPrivateAsset(ctx contractapi.TransactionContextInterface, myPrivateAssetID string) error {
	exists, err := c.MyPrivateAssetExists(ctx, myPrivateAssetID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("The asset %s already exists", myPrivateAssetID)
	}

	myPrivateAsset := new(MyPrivateAsset)

	transientData, _ := ctx.GetStub().GetTransient()

	privateValue, exists := transientData["privateValue"]

	if len(transientData) == 0 || !exists {
		return fmt.Errorf("The privateValue key was not specified in transient data. Please try again")
	}

	myPrivateAsset.PrivateValue = string(privateValue)

	bytes, _ := json.Marshal(myPrivateAsset)

	collectionName, collectionNameErr := getCollectionName(ctx)
	if collectionNameErr != nil {
		return collectionNameErr
	}

	return ctx.GetStub().PutPrivateData(collectionName, myPrivateAssetID, bytes)
}

// ReadMyPrivateAsset retrieves an instance of MyPrivateAsset from the private data collection
func (c *MyPrivateAssetContract) ReadMyPrivateAsset(ctx contractapi.TransactionContextInterface, myPrivateAssetID string) (*MyPrivateAsset, error) {
	exists, err := c.MyPrivateAssetExists(ctx, myPrivateAssetID)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("The asset %s does not exist", myPrivateAssetID)
	}

	collectionName, collectionNameErr := getCollectionName(ctx)
	if collectionNameErr != nil {
		return nil, collectionNameErr
	}

	bytes, _ := ctx.GetStub().GetPrivateData(collectionName, myPrivateAssetID)

	myPrivateAsset := new(MyPrivateAsset)

	err = json.Unmarshal(bytes, myPrivateAsset)

	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal private data collection data to type MyPrivateAsset")
	}

	return myPrivateAsset, nil
}

// UpdateMyPrivateAsset retrieves an instance of MyPrivateAsset from the private data collection and updates its value
func (c *MyPrivateAssetContract) UpdateMyPrivateAsset(ctx contractapi.TransactionContextInterface, myPrivateAssetID string) error {
	exists, err := c.MyPrivateAssetExists(ctx, myPrivateAssetID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", myPrivateAssetID)
	}

	transientData, _ := ctx.GetStub().GetTransient()
	newValue, exists := transientData["privateValue"]

	if len(transientData) == 0 || !exists {
		return fmt.Errorf("The privateValue key was not specified in transient data. Please try again")
	}

	myPrivateAsset := new(MyPrivateAsset)
	myPrivateAsset.PrivateValue = string(newValue)

	bytes, _ := json.Marshal(myPrivateAsset)

	collectionName, collectionNameErr := getCollectionName(ctx)
	if collectionNameErr != nil {
		return collectionNameErr
	}

	return ctx.GetStub().PutPrivateData(collectionName, myPrivateAssetID, bytes)
}

// DeleteMyPrivateAsset deletes an instance of MyPrivateAsset from the private data collection
func (c *MyPrivateAssetContract) DeleteMyPrivateAsset(ctx contractapi.TransactionContextInterface, myPrivateAssetID string) error {
	exists, err := c.MyPrivateAssetExists(ctx, myPrivateAssetID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", myPrivateAssetID)
	}

	collectionName, collectionNameErr := getCollectionName(ctx)
	if collectionNameErr != nil {
		return collectionNameErr
	}

	return ctx.GetStub().DelPrivateData(collectionName, myPrivateAssetID)
}

// VerifyMyPrivateAsset verifies the hash for an instance of MyPrivateAsset from the private data collection matches the hash stored in the public ledger //FIXME check this
func (c *MyPrivateAssetContract) VerifyMyPrivateAsset(ctx contractapi.TransactionContextInterface, mspid string, myPrivateAssetID string, objectToVerify *MyPrivateAsset) (bool, error) {
	bytes, _ := json.Marshal(objectToVerify)
	hashToVerify := sha256.New()
	hashToVerify.Write(bytes)

	pdHashBytes, err := ctx.GetStub().GetPrivateDataHash("_implicit_org_" + mspid, myPrivateAssetID)
	if err != nil {
		return false, err
	} else if len(pdHashBytes) == 0 {
		return false, fmt.Errorf("No private data hash with the Key: %s", myPrivateAssetID)
	}

	return hex.EncodeToString(hashToVerify.Sum(nil)) == hex.EncodeToString(pdHashBytes), nil
}
