/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

func main() {
	myPrivateAssetContract := new(MyPrivateAssetContract)
	myPrivateAssetContract.Info.Version = "0.0.1"
	myPrivateAssetContract.Info.Description = "My Private Data Smart Contract"
	myPrivateAssetContract.Info.License = new(metadata.LicenseMetadata)
	myPrivateAssetContract.Info.License.Name = "Apache-2.0"
	myPrivateAssetContract.Info.Contact = new(metadata.ContactMetadata)
	myPrivateAssetContract.Info.Contact.Name = "John Doe"

	chaincode, err := contractapi.NewChaincode(myPrivateAssetContract)
	chaincode.Info.Title = "demo-private-contract chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		panic("Could not create chaincode from MyPrivateAssetContract." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
