/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const getStateError = "private data get error"

var transient map[string][]byte

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (ms *MockStub) GetPrivateData(collection string, key string) ([]byte, error) {
	args := ms.Called(collection, key)

	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) GetPrivateDataHash(collection string, key string) ([]byte, error) {
	args := ms.Called(collection, key)

	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) GetTransient() (map[string][]byte, error) {

	return transient, nil
}

func (ms *MockStub) PutPrivateData(collection string, key string, value []byte) error {
	args := ms.Called(collection, key, value)

	return args.Error(0)
}

func (ms *MockStub) DelPrivateData(collection string, key string) error {
	args := ms.Called(collection, key)

	return args.Error(0)
}

type MockClientIdentity struct {
	cid.ClientIdentity
	mock.Mock
}

func (mci *MockClientIdentity) GetMSPID() (string, error) {
	args := mci.Called()
	return args.Get(0).(string), args.Error(1)
}

type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()

	return args.Get(0).(*MockStub)
}

func (mc *MockContext) GetClientIdentity() cid.ClientIdentity {
	args := mc.Called()

	return args.Get(0).(*MockClientIdentity)
}

func configureStub() (*MockContext, *MockStub) {
	var nilBytes []byte
	transient = make(map[string][]byte)

	testMyPrivateAsset := new(MyPrivateAsset)
	testMyPrivateAsset.PrivateValue = "set value"
	myPrivateAssetBytes, _ := json.Marshal(testMyPrivateAsset)
	hashToVerify := sha256.New()
	hashToVerify.Write(myPrivateAssetBytes)

	ms := new(MockStub)
	ms.On("GetPrivateData", mock.AnythingOfType("string"), "statebad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetPrivateData", mock.AnythingOfType("string"), "missingkey").Return(nilBytes, nil)
	ms.On("GetPrivateData", mock.AnythingOfType("string"), "existingkey").Return([]byte("some value"), nil)
	ms.On("GetPrivateData", mock.AnythingOfType("string"), "myPrivateAssetkey").Return(myPrivateAssetBytes, nil)
	ms.On("PutPrivateData", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	ms.On("DelPrivateData", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	ms.On("GetPrivateDataHash", mock.AnythingOfType("string"), "statebad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetPrivateDataHash", mock.AnythingOfType("string"), "missingkey").Return(nilBytes, nil)
	ms.On("GetPrivateDataHash", mock.AnythingOfType("string"), "existingkey").Return([]byte("some hash value"), nil)
	ms.On("GetPrivateDataHash", mock.AnythingOfType("string"), "myPrivateAssetkey").Return(hashToVerify.Sum(nil), nil)

	mci := new(MockClientIdentity)
	mci.On("GetMSPID").Return("Org1MSP", nil)

	mc := new(MockContext)
	mc.On("GetStub").Return(ms)
	mc.On("GetClientIdentity").Return(mci)

	return mc, ms
}

func TestMyPrivateAssetExists(t *testing.T) {
	var exists bool
	var err error

	ctx, _ := configureStub()
	c := new(MyPrivateAssetContract)

	exists, err = c.MyPrivateAssetExists(ctx, "statebad")
	assert.EqualError(t, err, getStateError)
	assert.False(t, exists, "should return false on error")

	exists, err = c.MyPrivateAssetExists(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when can read from world state but no value for key")
	assert.False(t, exists, "should return false when no value for key in world state")

	exists, err = c.MyPrivateAssetExists(ctx, "existingkey")
	assert.Nil(t, err, "should not return error when can read from world state and value exists for key")
	assert.True(t, exists, "should return true when value for key in world state")
}

func TestCreateMyPrivateAsset(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(MyPrivateAssetContract)

	err = c.CreateMyPrivateAsset(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	err = c.CreateMyPrivateAsset(ctx, "existingkey")
	assert.EqualError(t, err, "The asset existingkey already exists", "should error when exists returns true")

	err = c.CreateMyPrivateAsset(ctx, "missingkey")
	assert.EqualError(t, err, "The privateValue key was not specified in transient data. Please try again")

	transient["privateValue"] = []byte("some value")
	err = c.CreateMyPrivateAsset(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when transaction data provided")
	stub.AssertCalled(t, "PutPrivateData", "_implicit_org_Org1MSP", "missingkey", []byte("{\"privateValue\":\"some value\"}"))
}

func TestReadMyPrivateAsset(t *testing.T) {
	var myPrivateAsset *MyPrivateAsset
	var err error

	ctx, _ := configureStub()
	c := new(MyPrivateAssetContract)

	myPrivateAsset, err = c.ReadMyPrivateAsset(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when reading")
	assert.Nil(t, myPrivateAsset, "should not return MyPrivateAsset when exists errors when reading")

	myPrivateAsset, err = c.ReadMyPrivateAsset(ctx, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when reading")
	assert.Nil(t, myPrivateAsset, "should not return MyPrivateAsset when key does not exist in private data collection when reading")

	myPrivateAsset, err = c.ReadMyPrivateAsset(ctx, "existingkey")
	assert.EqualError(t, err, "Could not unmarshal private data collection data to type MyPrivateAsset", "should error when data in key is not MyPrivateAsset")
	assert.Nil(t, myPrivateAsset, "should not return MyPrivateAsset when data in key is not of type MyPrivateAsset")

	myPrivateAsset, err = c.ReadMyPrivateAsset(ctx, "myPrivateAssetkey")
	expectedMyPrivateAsset := new(MyPrivateAsset)
	expectedMyPrivateAsset.PrivateValue = "set value"
	assert.Nil(t, err, "should not return error when MyPrivateAsset exists in private data collection when reading")
	assert.Equal(t, expectedMyPrivateAsset, myPrivateAsset, "should return deserialized MyPrivateAsset from private data collection")
}

func TestUpdateMyPrivateAsset(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(MyPrivateAssetContract)

	err = c.UpdateMyPrivateAsset(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when updating")

	err = c.UpdateMyPrivateAsset(ctx, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists is false when updating")

	transient["privateValue"] = []byte("new value")
	err = c.UpdateMyPrivateAsset(ctx, "myPrivateAssetkey")
	expectedMyPrivateAsset := new(MyPrivateAsset)
	expectedMyPrivateAsset.PrivateValue = "new value"
	expectedMyPrivateAssetBytes, _ := json.Marshal(expectedMyPrivateAsset)
	assert.Nil(t, err, "should not return error when MyPrivateAsset exists in private data collection when updating")
	stub.AssertCalled(t, "PutPrivateData", "_implicit_org_Org1MSP", "myPrivateAssetkey", expectedMyPrivateAssetBytes)
}

func TestDeleteMyPrivateAsset(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(MyPrivateAssetContract)

	err = c.DeleteMyPrivateAsset(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	err = c.DeleteMyPrivateAsset(ctx, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns false when deleting")

	err = c.DeleteMyPrivateAsset(ctx, "myPrivateAssetkey")
	assert.Nil(t, err, "should not return error when MyPrivateAsset exists in private data collection when deleting")
	stub.AssertCalled(t, "DelPrivateData", "_implicit_org_Org1MSP", "myPrivateAssetkey")
}

func TestVerifyMyPrivateAsset(t *testing.T) {
	var myPrivateAsset *MyPrivateAsset
	var exists bool
	var err error

	ctx, stub := configureStub()
	c := new(MyPrivateAssetContract)

	myPrivateAsset = new(MyPrivateAsset)
	myPrivateAsset.PrivateValue = "set value"

	exists, err = c.VerifyMyPrivateAsset(ctx, "Org1MSP", "statebad", myPrivateAsset)
	assert.False(t, exists, "should return false when unable to read the hash")
	assert.EqualError(t, err, getStateError)

	exists, err = c.VerifyMyPrivateAsset(ctx, "Org1MSP", "missingkey", myPrivateAsset)
	assert.False(t, exists, "should return false when key does not exist")
	assert.EqualError(t, err, "No private data hash with the Key: missingkey", "should error when key does not exist")

	exists, err = c.VerifyMyPrivateAsset(ctx, "Org1MSP", "myPrivateAssetkey", myPrivateAsset)
	assert.True(t, exists, "should return true when hash in world state matched hash from data collection")
	assert.Nil(t, err, "should not return error when hash in world state matched hash from data collection")
	stub.AssertCalled(t, "GetPrivateDataHash", "_implicit_org_Org1MSP", "myPrivateAssetkey")
}
