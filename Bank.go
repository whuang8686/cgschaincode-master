/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright SecurityNameship.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE/2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */
/*
Building Chaincode
Now let’s compile your chaincode.

go get -u --tags nopkcs11 github.com/hyperledger/fabric/core/chaincode/shim
go build --tags nopkcs11
*/

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	//"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"net/http"
	"io/ioutil"
	"net"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

const (
	millisPerSecond       = int64(time.Second / time.Millisecond)
	nanosPerMillisecond   = int64(time.Millisecond / time.Nanosecond)
	layout                = "2006/01/02"
	unitAmount            = int64(1000000)           //1單位=100萬
	perDayMillionInterest = float64(27.3972603)      //每1百萬面額，利率=1%，一天的利息
	perDayInterest        = float64(0.0000273972603) //每1元面額，利率=1%，一天的利息
	//InterestObjectType    = "Interest"
)


type Cpty struct {
	ObjectType           string          `json:"docType"`             //Cpty
	CptyID               string          `json:"CptyID"`
	CptyName             string          `json:"CptyName"`
	CptyStatus           string          `json:"CptyStatus"`          //Disable, Enable
	UpdateTime           string          `json:"updateTime"`          //更新時間	
}

type User struct {
	ObjectType           string          `json:"docType"`             //User
	UserID               string          `json:"UserID"`              //CptyID + TimeNow 0001 + 20190224105134
	CptyID               string          `json:"CptyID"`                             
	UserName             string          `json:"UserName"`
	Password             string          `json:"Password"`
	UserStatus           string          `json:"UserStatus"`          //Disable, Enable, Pending
	UpdateTime           string          `json:"updateTime"`          //更新時間	
}

type CptyISDA struct {
	ObjectType           string          `json:"docType"`             //CptyISDA
	CptyISDAID           string          `json:"CptyISDAID"`          //OwnCptyID + CptyID + TimeNow
	OwnCptyID            string          `json:"OwnCptyID"`
	CptyID               string          `json:"CptyID"`              //交易對手
	CptyIndependAmt      int64           `json:"CptyIndependAmt"`     //交易對手單獨提列金額
	OwnIndependAmt       int64           `json:"OwnIndependAmt"`      //本行單獨提列金額
	CptyThreshold        int64           `json:"CptyThreshold"`       //交易對手門檻金額 
	OwnThreshold         int64           `json:"OwnThreshold"`        //本行門鑑金額
	CptyMTA              int64           `json:"CptyMTA"`             //交易對手最低轉讓金額
	OwnMTA               int64           `json:"OwnMTA"`              //本行最低轉讓金額
	Rounding             int64           `json:"Rounding"`            //整數計算
	StartDate            string          `json:"StartDate"`           //合約起日
	EndDate              string          `json:"EndDate"`             //合約迄日
	USDCashPCT           float64         `json:"USDCashPCT"`          //USDCashPCT
	TWDCashPCT           float64         `json:"TWDCashPCT"`          //TWDCashPCT
	USDBondPCT           float64         `json:"USDBondPCT"`          //USDBondPCT
	TWDBondPCT           float64         `json:"TWDBondPCT"`          //TWDBondPCT
}

type CptyAsset struct {
	ObjectType           string          `json:"docType"`             //CptyAsset
	CptyAssetID          string          `json:"CptyAssetID"`         //OwnCptyID + TimeNow
	OwnCptyID            string          `json:"OwnCptyID"`
	USDBond              float64         `json:"USDBond"`             //USDBondPCT
	TWDBond              float64         `json:"TWDBond"`             //TWDBondPCT
	AUD                  float64         `json:"AUD"`                 //AUD 
	BRL                  float64         `json:"BRL"`                 //BRL
	CAD                  float64         `json:"CAD"`                 //CAD
	CHF                  float64         `json:"CHF"`                 //CHF
	CNY                  float64         `json:"CNY"`                 //CNY
	EUR                  float64         `json:"EUR"`                 //EUR
	GBP                  float64         `json:"GBP"`                 //GBP
	HKD                  float64         `json:"HKD"`                 //HKD
	INR                  float64         `json:"INR"`                 //INR
	JPY                  float64         `json:"JPY"`                 //JPY
	KRW                  float64         `json:"KRW"`                 //KRW
	MOP                  float64         `json:"MOP"`                 //MOP
	MYR                  float64         `json:"MYR"`                 //MYR
	NZD                  float64         `json:"NZD"`                 //NZD
	PHP                  float64         `json:"PHP"`                 //PHP
	SEK                  float64         `json:"SEK"`                 //SEK
	SGD                  float64         `json:"SGD"`                 //SGD
	THB                  float64         `json:"THB"`                 //THB
	TWD                  float64         `json:"TWD"`                 //TWD
	USD                  float64         `json:"USD"`                 //USD
	ZAR                  float64         `json:"ZAR"`                 //ZAR
}


type FXTrade struct {
	ObjectType           string          `json:"docType"`             //docType is used to distinguish the various types of objects in state database
	TXID                 string          `json:"TXID"`                //OwnCptyID + TXType + TimeNow
	TXType               string          `json:"TXType"`              //Transaction TXType BUY or SELL
	TXKinds              string          `json:"TXKinds"`             //交易種類SPOT or FW
	OwnCptyID            string          `json:"OwnCptyID"`
	CptyID               string          `json:"CptyID"`              //交易對手
	TradeDate            string          `json:"TradeDate"`           //交易日
	MaturityDate         string          `json:"MaturityDate"`        //到期日
	Contract             string          `json:"Contract"`            //交易合約 
	Curr1                string          `json:"Curr1"`               //Curr1
	Amount1              float64         `json:"Amount1"`             //Amount1
	Curr2                string          `json:"Curr2"`               //Curr2
	Amount2              float64         `json:"Amount2"`             //Amount2
	NetPrice             float64         `json:"NetPrice"`            //合約價
	isPutToQueue         bool            `json:"isPutToQueue"`        //isPutToQueue = true 代表資料檢核成功
	TXStatus             string          `json:"TXStatus"`            //Pending, Matched, Finished, Cancelled
	CreateTime           string          `json:"createTime"`          //建立時間
	UpdateTime           string          `json:"updateTime"`          //更新時間
	TXIndex              string          `json:"TXIndex"`             //Transaction Index(全部比對)
	TXHcode              string          `json:"TXHcode"`             //Transaction Hcode(更正交易序號)
	MatchedTXID          string          `json:"MatchedTXID"`         //比對序號 OwnCptyID(後來新增的)+TXType+TimeNow 0002S20180829063021
	TXMemo               string          `json:"TXMemo"`              //交易說明
	TXErrMsg             string          `json:"TXErrMsg"`            //交易錯誤說明
}

/*
 * The Init method is called when the Smart Contract "CGSecurity" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in sepaRate function // see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {

	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "CGSecurity"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "createCpty" {
		return s.createCpty(APIstub, args)
	} else if function == "updateCpty" {
		return s.updateCpty(APIstub, args)
    } else if function == "deleteCpty" {
		return s.deleteCpty(APIstub, args)
	} else if function == "queryCpty"  {
		return s.queryCpty(APIstub, args)	
	} else if function == "queryAllCpty" {
		return s.queryAllCpty(APIstub, args)	
	//User
	} else if function == "createUser" {
		return s.createUser(APIstub, args)		
	} else if function == "updateUser" {
		return s.updateUser(APIstub, args)			
	} else if function == "queryUser" {
		return s.queryUser(APIstub, args)			
	//CptyISDA
	} else if function == "createCptyISDA" {
		return s.createCptyISDA(APIstub, args)
	} else if function == "updateCptyISDA" {
		return s.updateCptyISDA(APIstub, args)
	} else if function == "deleteCptyISDA" {
		return s.deleteCptyISDA(APIstub, args)
	} else if function == "queryCptyISDA"  {
		return s.queryCptyISDA(APIstub, args)	
	} else if function == "queryCptyISDAByCpty"  {
		return s.queryCptyISDAByCpty(APIstub, args)			
	} else if function == "queryAllCptyISDA" {
		return s.queryAllCptyISDA(APIstub, args)
	//CptyAsset
    } else if function == "createCptyAsset" {
        return s.createCptyAsset(APIstub, args)
	} else if function == "updateCptyAsset" {
		return s.updateCptyAsset(APIstub, args)
    } else if function == "queryTXIDCptyAsset" {
		return s.queryTXIDCptyAsset(APIstub, args)		
	} else if function == "CreateCashFlow" {
		return s.CreateCashFlow(APIstub, args)				

	//} else if function == "fetchEURUSDviaOraclize" {
	//	return s.fetchEURUSDviaOraclize(APIstub)

    } else if function == "getrate" {
		return s.getrate(APIstub, args)

	// Transaction Functions
    } else if function == "FXTradeTransfer" {
		return s.FXTradeTransfer(APIstub, args)
	} else if function == "CorrectFXTradeTransfer" {
		return s.CorrectFXTradeTransfer(APIstub, args)
	// Transaction MTM Functions
	} else if function == "FXTradeMTM" {
		return s.FXTradeMTM(APIstub, args)	
	} else if function == "CreateFXTradeMTM" {
		return s.CreateFXTradeMTM(APIstub, args)		
	} else if function == "deleteFXTradeMTM" {
		return s.deleteFXTradeMTM(APIstub, args)
	} else if function == "QueryFXTradeMTM" {
		return s.QueryFXTradeMTM(APIstub, args)				
	// Transaction Settlment Functions
	} else if function == "FXTradeSettlment" {
		return s.FXTradeSettlment(APIstub, args)	
    // Transaction FXTradeCollateral Functions
    } else if function == "FXTradeCollateral" {
	    return s.FXTradeCollateral(APIstub, args)
	} else if function == "UpdateFXTradeCollateral" {
		return s.UpdateFXTradeCollateral(APIstub, args)	   
	} else if function == "CreateCollateralDetail" {
		return s.CreateCollateralDetail(APIstub, args)	 
	} else if function == "CollateralSettlment" {
		return s.CollateralSettlment(APIstub, args)		
	} else if function == "queryFXTradeCollateral" {
		return s.queryFXTradeCollateral(APIstub, args)				  
    // MTM Price Functions
    } else if function == "createMTMPrice" {
		return s.createMTMPrice(APIstub, args)
	} else if function == "createBondPrice" {
		return s.createBondPrice(APIstub, args)

    } else if function == "queryTables" { 
		return s.queryTables(APIstub, args)
	} else if function == "queryTXIDTransactions" {
		return s.queryTXIDTransactions(APIstub, args)	
	} else if function == "queryTXIDCollateral" {
		return s.queryTXIDCollateral(APIstub, args)			
	} else if function == "queryTXKEYTransactions" {
	    return s.queryTXKEYTransactions(APIstub, args)
	} else if function == "queryHistoryTXKEYTransactions" {
	    return s.queryHistoryTXKEYTransactions(APIstub, args)
	} else if function == "getHistoryForTransaction" {
	    return s.getHistoryForTransaction(APIstub, args)
	} else if function == "getHistoryTXIDForTransaction" {
	    return s.getHistoryTXIDForTransaction(APIstub, args)
	} else if function == "getHistoryForQueuedTransaction" {
	    return s.getHistoryForQueuedTransaction(APIstub, args)
	} else if function == "getHistoryTXIDForQueuedTransaction" {
	    return s.getHistoryTXIDForQueuedTransaction(APIstub, args)
	} else if function == "queryAllTransactions" {
	    return s.queryAllTransactions(APIstub, args)
	} else if function == "queryAllQueuedTransactions" {
	    return s.queryAllQueuedTransactions(APIstub, args)
	} else if function == "queryAllHistoryTransactions" {
	    return s.queryAllHistoryTransactions(APIstub, args)
	} else if function == "queryAllTransactionKeys" {
	    return s.queryAllTransactionKeys(APIstub, args)
	} else if function == "queryQueuedTransactionStatus" {
	    return s.queryQueuedTransactionStatus(APIstub, args)
	} else if function == "queryHistoryTransactionStatus" {
		return s.queryHistoryTransactionStatus(APIstub, args)
	} else if function == "queryMTMTransactionStatus" {
		return s.queryMTMTransactionStatus(APIstub, args)
	} else if function == "queryCollateralTransactionStatus" {
		return s.queryCollateralTransactionStatus(APIstub, args)		
	} else if function == "queryCollateralDetailStatus" {
		return s.queryCollateralDetailStatus(APIstub, args)			
	} else if function == "queryDayEndCollateralStatus" {
		return s.queryDayEndCollateralStatus(APIstub, args)					
	} else if function == "queryCptyISDAStatus" {
		return s.queryCptyISDAStatus(APIstub, args)		
    } else if function == "queryCptyAssetStatus" {
		return s.queryCptyAssetStatus(APIstub, args)	
	} else if function == "queryCashFlowStatus" {
		return s.queryCashFlowStatus(APIstub, args)			
	} else if function == "queryMTMPrice" {
		return s.queryMTMPrice(APIstub, args)		
	} else if function == "queryBondPrice" {
		return s.queryBondPrice(APIstub, args)			
	} else if function == "updateQueuedTransactionHcode" {
	    return s.updateQueuedTransactionHcode(APIstub, args)
	} else if function == "updateHistoryTransactionHcode" {
		return s.updateHistoryTransactionHcode(APIstub, args)	
	} else if function == "testEvent" {
		return s.testEvent(APIstub, args)	
			
	} else {
    //map functions
		return s.mapFunction(APIstub, function, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

//peer chaincode invoke -n mycc -c '{"Args":["testEvent"]}' -C myc
func (s *SmartContract) testEvent(stub shim.ChaincodeStubInterface, args []string) peer.Response{
	tosend := "Event send data is here!"
	err := stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
	   return shim.Error(err.Error())
	}
	return shim.Success(nil)
 }

func (s *SmartContract) mapFunction(stub shim.ChaincodeStubInterface, function string, args []string) peer.Response {
	switch function {

	case "put":
		if len(args) < 2 {
			return shim.Error("put operation must include two arguments: [key, value]")
		}
		key := args[0]
		value := args[1]

		if err := stub.PutState(key, []byte(value)); err != nil {
			fmt.Printf("Error putting state %s", err)
			return shim.Error(fmt.Sprintf("put operation failed. Error updating state: %s", err))
		}

		indexName := "compositeKeyTest"
		compositeKeyTestIndex, err := stub.CreateCompositeKey(indexName, []string{key})
		if err != nil {
			return shim.Error(err.Error())
		}

		valueByte := []byte{0x00}
		if err := stub.PutState(compositeKeyTestIndex, valueByte); err != nil {
			fmt.Printf("Error putting state with compositeKey %s", err)
			return shim.Error(fmt.Sprintf("put operation failed. Error updating state with compositeKey: %s", err))
		}

		return shim.Success(nil)

	case "remove":
		if len(args) < 1 {
			return shim.Error("remove operation must include one argument: [key]")
		}
		key := args[0]

		err := stub.DelState(key)
		if err != nil {
			return shim.Error(fmt.Sprintf("remove operation failed. Error updating state: %s", err))
		}
		return shim.Success(nil)

	case "get":
		if len(args) < 1 {
			return shim.Error("get operation must include one argument, a key")
		}
		key := args[0]
		value, err := stub.GetState(key)
		if err != nil {
			return shim.Error(fmt.Sprintf("get operation failed. Error accessing state: %s", err))
		}
		jsonVal, err := json.Marshal(string(value))
		return shim.Success(jsonVal)

	case "keys":
		if len(args) < 2 {
			return shim.Error("put operation must include two arguments, a key and value")
		}
		startKey := args[0]
		endKey := args[1]

		//sleep needed to test peer's timeout behavior when using iterators
		stime := 0
		if len(args) > 2 {
			stime, _ = strconv.Atoi(args[2])
		}

		keysIter, err := stub.GetStateByRange(startKey, endKey)
		if err != nil {
			return shim.Error(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
		}
		defer keysIter.Close()

		var keys []string
		for keysIter.HasNext() {
			//if sleeptime is specied, take a nap
			if stime > 0 {
				time.Sleep(time.Duration(stime) * time.Millisecond)
			}

			response, iterErr := keysIter.Next()
			if iterErr != nil {
				return shim.Error(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
			}
			keys = append(keys, response.Key)
		}

		for key, value := range keys {
			fmt.Printf("key %d contains %s\n", key, value)
		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return shim.Error(fmt.Sprintf("keys operation failed. Error marshaling JSON: %s", err))
		}

		return shim.Success(jsonKeys)
	case "query":
		query := args[0]
		keysIter, err := stub.GetQueryResult(query)
		if err != nil {
			return shim.Error(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
		}
		defer keysIter.Close()

		var keys []string
		for keysIter.HasNext() {
			response, iterErr := keysIter.Next()
			if iterErr != nil {
				return shim.Error(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
			}
			keys = append(keys, response.Key)
		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return shim.Error(fmt.Sprintf("query operation failed. Error marshaling JSON: %s", err))
		}

		return shim.Success(jsonKeys)
	case "history":
		key := args[0]
		keysIter, err := stub.GetHistoryForKey(key)
		if err != nil {
			return shim.Error(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
		}
		defer keysIter.Close()

		var keys []string
		for keysIter.HasNext() {
			response, iterErr := keysIter.Next()
			if iterErr != nil {
				return shim.Error(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
			}
			keys = append(keys, response.TxId)
		}

		for key, txID := range keys {
			fmt.Printf("key %d contains %s\n", key, txID)
		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return shim.Error(fmt.Sprintf("query operation failed. Error marshaling JSON: %s", err))
		}

		return shim.Success(jsonKeys)

	default:
		return shim.Success([]byte("Unsupported operation"))
	}
}

//peer chaincode invoke -n mycc -c '{"Args":["createCpty", "0001","CptyA","Active"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createCpty", "0002","CptyB","Active"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createCpty", "0003","CptyC","Active"]}' -C myc
func (s *SmartContract) createCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	
	TimeNow2 := time.Now().Format(timelayout2)

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	fmt.Sprintf(TimeNow2)
	
	var Cpty = Cpty{ObjectType: "Cpty", CptyID: args[0], CptyName: args[1], CptyStatus: args[2], UpdateTime:TimeNow2}
	CptyAsBytes, _ := json.Marshal(Cpty)
	err := APIstub.PutState(Cpty.CptyID, CptyAsBytes)
	if err != nil {
		return shim.Error("Failed to create state")
	}

	return shim.Success(nil)
}


//peer chaincode invoke -n mycc -c '{"Args":["updateCpty", "0001","Lock"]}' -C myc
func (s *SmartContract) updateCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// 判斷是否有輸入值 

	CptyAsBytes, _ := APIstub.GetState(args[0])
	Cpty := Cpty{}

	json.Unmarshal(CptyAsBytes, &Cpty)
	Cpty.ObjectType = "Cpty"
	Cpty.CptyStatus = args[1]
	Cpty.UpdateTime = time.Now().Format(timelayout2)
	
	CptyAsBytes, _ = json.Marshal(Cpty)
	err := APIstub.PutState(args[0], CptyAsBytes)
	if err != nil {
		return shim.Error("Failed to change state")
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["deleteCpty", "0003"]}' -C myc
func (s *SmartContract) deleteCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Delete the key from the state in ledger
	err := APIstub.DelState(args[0])
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

//peer chaincode query -n mycc -c '{"Args":["queryCpty","0001"]}' -C myc
func (s *SmartContract) queryCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
    var queryString string
	CptyID := args[0]
	if CptyID == "All" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"Cpty\"}}")
	}  else {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"Cpty\",\"CptyID\":\"%s\"}}", CptyID)	
	}	

	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("esultsIterator.Close")
 
    var buffer bytes.Buffer
    buffer.WriteString("[")
 
	bArrayMemberAlreadyWritten := false
	fmt.Printf("bArrayMemberAlreadyWritten := false\n")
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
         
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
		}
		fmt.Printf("resultsIterator.HasNext\n")
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")
        buffer.WriteString(", \"Record\":")
         
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
	}
	
    buffer.WriteString("]")
 
    return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryAllCpty"]}' -C myc
func (s *SmartContract) queryAllCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode invoke -n mycc -c '{"Args":["createUser", "BANK1","BANK1User1","BANK1User1","Active"]}' -C myc
func (s *SmartContract) createUser(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	
	TimeNow := time.Now().Format(timelayout)
	TimeNow2 := time.Now().Format(timelayout2)

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	UserID := args[0] + TimeNow 
	//fmt.Sprintf(TimeNow2)
	var User = User{ObjectType: "User", UserID: UserID, CptyID: args[0], UserName: args[1], Password: args[2], UserStatus: args[3], UpdateTime:TimeNow2}
	UserAsBytes, _ := json.Marshal(User)
	err := APIstub.PutState(User.UserID, UserAsBytes)
	if err != nil {
		return shim.Error("Failed to create state")
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["updateUser", "000120190224105134","Cpty1User1","Cpty1Pass2","Active"]}' -C myc
func (s *SmartContract) updateUser(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	UserAsBytes, _ := APIstub.GetState(args[0])
	User := User{}

	json.Unmarshal(UserAsBytes, &User)
	User.ObjectType = "User"
	User.UserName = args[1]
	User.Password = args[2]
	User.UserStatus = args[3]
	User.UpdateTime = time.Now().Format(timelayout2)
	
	UserAsBytes, _ = json.Marshal(User)
	err := APIstub.PutState(args[0], UserAsBytes)
	if err != nil {
		return shim.Error("Failed to change state")
	}

	return shim.Success(nil)
}

//peer chaincode query -n mycc -c '{"Args":["queryUser","BANK1","BANK1User1"]}' -C myc
func (s *SmartContract) queryUser(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
    var queryString string
	CptyID := args[0]
	UserName := args[1]
	if CptyID == "All" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"User\"}}")
	}  else {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"User\",\"CptyID\":\"%s\",\"UserName\":\"%s\"}}", CptyID, UserName)	
	}	

	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("esultsIterator.Close")
 
    var buffer bytes.Buffer
    buffer.WriteString("[")
 
	bArrayMemberAlreadyWritten := false
	fmt.Printf("bArrayMemberAlreadyWritten := false\n")
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
         
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
		}
		fmt.Printf("resultsIterator.HasNext\n")
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")
        buffer.WriteString(", \"Record\":")
         
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
	}
	
    buffer.WriteString("]")
 
    return shim.Success(buffer.Bytes())
}

/*
peer chaincode invoke -n mycc -c '{"Args":["createCptyISDA", "0001","0003","0","0","35000000","8000000","3000000","500000","100000","2018/01/01","2020/12/31","1","0.95","0.96","0.89"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["createCptyISDA", "0001","0002","0","0","25000000","6000000","3000000","500000","100000","2018/01/01","2020/12/31","1","0.95","0.96","0.89"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["createCptyISDA", "0001","0004","0","0","45000000","7000000","3000000","500000","100000","2018/01/01","2020/12/31","1","0.95","0.96","0.89"]}' -C myc
*/
func (s *SmartContract) createCptyISDA(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)

	if len(args) != 15 {
		return shim.Error("Incorrect number of arguments. Expecting 15")
	}

	var newCptyIndependAmt, newOwnIndependAmt, newCptyThreshold, newOwnThreshold, newCptyMTA, newOwnMTA, newRounding int64
	var newUSDCashPCT, newTWDCashPCT, newUSDBondPCT, newTWDBondPCT float64
	var OwnCptyID,CptyID,CptyISDAID string
	 
	OwnCptyID = args[0]
	CptyID = args[1]

	newCptyIndependAmt, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newOwnIndependAmt, err = strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCptyThreshold, err = strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newOwnThreshold, err = strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCptyMTA, err = strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newOwnMTA, err = strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newRounding, err = strconv.ParseInt(args[8], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDCashPCT, err = strconv.ParseFloat(args[11], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWDCashPCT, err = strconv.ParseFloat(args[12], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDBondPCT, err = strconv.ParseFloat(args[13], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWDBondPCT, err = strconv.ParseFloat(args[14], 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	CptyISDAID = OwnCptyID + CptyID + TimeNow 
	fmt.Println("createCptyISDA= " + CptyISDAID + "\n") 

	var CptyISDA = CptyISDA{ObjectType: "CptyISDA", CptyISDAID: CptyISDAID, OwnCptyID: OwnCptyID, CptyID: CptyID, CptyIndependAmt: newCptyIndependAmt, OwnIndependAmt: newOwnIndependAmt, CptyThreshold: newCptyThreshold, OwnThreshold: newOwnThreshold, CptyMTA: newCptyMTA, OwnMTA: newOwnMTA, Rounding: newRounding, StartDate: args[9], EndDate: args[10], USDCashPCT: newUSDCashPCT, TWDCashPCT: newTWDCashPCT, USDBondPCT: newUSDBondPCT, TWDBondPCT: newTWDBondPCT}
	CptyISDAAsBytes, _ := json.Marshal(CptyISDA)
	err1 := APIstub.PutState(CptyISDA.CptyISDAID, CptyISDAAsBytes)
	if err1 != nil {
		return shim.Error("Failed to create state")
		fmt.Println("createCptyISDA.PutState\n") 
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["updateCptyISDA", "0001","0002","0","0","25500000","8000000","3000000","500000","100000","2018/01/01","2020/12/31","1","0.95","0.96","0.89","0001000320180923051259"]}' -C myc
func (s *SmartContract) updateCptyISDA(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 16 {
		return shim.Error("Incorrect number of arguments. Expecting 16")
	}

	// 判斷是否有輸入值 

	CptyISDAAsBytes, _ := APIstub.GetState(args[15])
	CptyISDA := CptyISDA{}

	var newCptyIndependAmt, newOwnIndependAmt, newCptyThreshold, newOwnThreshold, newCptyMTA, newOwnMTA, newRounding int64
	var newUSDCashPCT, newTWDCashPCT, newUSDBondPCT, newTWDBondPCT float64

	newCptyIndependAmt, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newOwnIndependAmt, err = strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCptyThreshold, err = strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newOwnThreshold, err = strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCptyMTA, err = strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newOwnMTA, err = strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newRounding, err = strconv.ParseInt(args[8], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDCashPCT, err = strconv.ParseFloat(args[11], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWDCashPCT, err = strconv.ParseFloat(args[12], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDBondPCT, err = strconv.ParseFloat(args[13], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWDBondPCT, err = strconv.ParseFloat(args[14], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(CptyISDAAsBytes, &CptyISDA)
	CptyISDA.ObjectType = "CptyISDA"
	CptyISDA.OwnCptyID = args[0]
	CptyISDA.CptyID = args[1]
	CptyISDA.CptyIndependAmt = newCptyIndependAmt
	CptyISDA.OwnIndependAmt = newOwnIndependAmt
	CptyISDA.CptyThreshold = newCptyThreshold
	CptyISDA.OwnThreshold = newOwnThreshold
	CptyISDA.CptyMTA = newCptyMTA
	CptyISDA.OwnMTA = newOwnMTA
	CptyISDA.Rounding = newRounding
	CptyISDA.StartDate = args[9]
	CptyISDA.EndDate = args[10]
	CptyISDA.USDCashPCT = newUSDCashPCT
	CptyISDA.TWDCashPCT = newTWDCashPCT
	CptyISDA.USDBondPCT = newUSDBondPCT
	CptyISDA.TWDBondPCT = newTWDBondPCT
	
	CptyISDAAsBytes, _ = json.Marshal(CptyISDA)
	err1 := APIstub.PutState(args[15], CptyISDAAsBytes)
	if err1 != nil {
		return shim.Error("Failed to change state")
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["deleteCptyISDA", "00003"]}' -C myc
func (s *SmartContract) deleteCptyISDA(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Delete the key from the state in ledger
	err := APIstub.DelState(args[0])
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

//peer chaincode query -n mycc -c '{"Args":["queryCptyISDA","0001000320180923051259"]}' -C myc
func (s *SmartContract) queryCptyISDA(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	CptyISDAAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(CptyISDAAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["queryAllCptyISDA","00001","99999"]}' -C myc
func (s *SmartContract) queryAllCptyISDA(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryCptyISDAStatus","0001"]}' -C myc
func (s *SmartContract) queryCptyISDAStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var queryString,queryString2 string
	CptyID := args[0]

	if CptyID == "All" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CptyISDA\"}}")
	}  else {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CptyISDA\",\"OwnCptyID\":\"%s\"}}", CptyID)	
		queryString2 = fmt.Sprintf("{\"selector\":{\"docType\":\"CptyISDA\",\"CptyID\":\"%s\"}}", CptyID)	
	}	

	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("esultsIterator.Close")
 
    var buffer bytes.Buffer
    buffer.WriteString("[")
 
	bArrayMemberAlreadyWritten := false
	fmt.Printf("bArrayMemberAlreadyWritten := false\n")
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
         
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
		}
		fmt.Printf("resultsIterator.HasNext\n")
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")
        buffer.WriteString(", \"Record\":")
         
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
	}
	
	if CptyID != "All" {
		resultsIterator2, err := APIstub.GetQueryResult(queryString2)
    	if err != nil {
        	return shim.Error(err.Error())
    	}
		defer resultsIterator2.Close()
		bArrayMemberAlreadyWritten2 := false
    	for resultsIterator2.HasNext() {
        	queryResponse2, err := resultsIterator2.Next()
        	if err != nil {
            	return shim.Error(err.Error())
        	}
         
        	if bArrayMemberAlreadyWritten2 == true {
            buffer.WriteString(",")
        	}
			fmt.Printf("resultsIterator2.HasNext\n")
			buffer.WriteString("{\"Key\":")
        	buffer.WriteString("\"")
        	buffer.WriteString(queryResponse2.Key)
        	buffer.WriteString("\"")
			buffer.WriteString(", \"Record\":")
         
        	buffer.WriteString(string(queryResponse2.Value))
        	buffer.WriteString("}")
        	bArrayMemberAlreadyWritten2 = true
		}
    }
    buffer.WriteString("]")
 
    return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryCptyISDAByCpty","0001","0004"]}' -C myc
func (s *SmartContract) queryCptyISDAByCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	var queryString string
	OwnCptyID := args[0]
	CptyID := args[1]

	queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CptyISDA\",\"OwnCptyID\":\"%s\",\"CptyID\":\"%s\"}}", OwnCptyID, CptyID)
	
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("esultsIterator.Close")
 
    var buffer bytes.Buffer
    buffer.WriteString("[")
 
	bArrayMemberAlreadyWritten := false
	fmt.Printf("bArrayMemberAlreadyWritten := false\n")
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
         
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
		}
		fmt.Printf("resultsIterator.HasNext\n")
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")
        buffer.WriteString(", \"Record\":")
         
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
	}
    buffer.WriteString("]")
 
    return shim.Success(buffer.Bytes())
}

/*
peer chaincode invoke -n mycc -c '{"Args":["createCptyAsset", "0001","45000000","8000000","3000000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["createCptyAsset", "0002","45000000","8000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000","6000000"]}' -C myc
*/
func (s *SmartContract) createCptyAsset(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)

	if len(args) != 24 {
		return shim.Error("Incorrect number of arguments. Expecting 21")
	}

	var newUSDBond, newTWDBond float64
	var newAUD,newBRL,newCAD,newCHF,newCNY,newEUR,newGBP,newHKD,newINR,newJPY,newKRW float64
    var newMOP,newMYR,newNZD,newPHP,newSEK,newSGD,newTHB,newTWD,newUSD,newZAR float64
	var OwnCptyID,CptyAssetID string
	 
	OwnCptyID = args[0]

	newUSDBond, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWDBond, err = strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newAUD, err = strconv.ParseFloat(args[3], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newBRL, err = strconv.ParseFloat(args[4], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCAD, err = strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCHF, err = strconv.ParseFloat(args[6], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCNY, err = strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEUR, err = strconv.ParseFloat(args[8], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newGBP, err = strconv.ParseFloat(args[9], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newHKD, err = strconv.ParseFloat(args[10], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newINR, err = strconv.ParseFloat(args[11], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newJPY, err = strconv.ParseFloat(args[12], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newKRW, err = strconv.ParseFloat(args[13], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newMOP, err = strconv.ParseFloat(args[14], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newMYR, err = strconv.ParseFloat(args[15], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newNZD, err = strconv.ParseFloat(args[16], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newPHP, err = strconv.ParseFloat(args[17], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newSEK, err = strconv.ParseFloat(args[18], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newSGD, err = strconv.ParseFloat(args[19], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTHB, err = strconv.ParseFloat(args[20], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWD, err = strconv.ParseFloat(args[21], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSD, err = strconv.ParseFloat(args[22], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newZAR, err = strconv.ParseFloat(args[23], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	CptyAssetID = OwnCptyID + TimeNow 
	fmt.Println("createCptyAsset= " +  CptyAssetID + "\n") 

	var CptyAsset = CptyAsset{ObjectType: "CptyAsset", CptyAssetID: CptyAssetID, OwnCptyID: OwnCptyID, USDBond: newUSDBond, TWDBond: newTWDBond, AUD: newAUD,BRL: newBRL,CAD: newCAD,CHF: newCHF,CNY: newCNY,EUR: newEUR,GBP: newGBP,HKD: newHKD,INR: newINR,JPY: newJPY,KRW: newKRW,MOP: newMOP,MYR: newMYR,NZD: newNZD,PHP: newPHP,SEK: newSEK,SGD: newSGD,THB: newTHB,TWD: newTWD,USD: newUSD,ZAR: newZAR}
	CptyAssetAsBytes, _ := json.Marshal(CptyAsset)
	err1 := APIstub.PutState(CptyAsset.CptyAssetID, CptyAssetAsBytes)
	if err1 != nil {
		return shim.Error("Failed to create state")
		fmt.Println("createCptyAsset.PutState\n") 
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["updateCptyAsset", "0001","65000000","8800000","3000000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","500000","000120181016114811"]}' -C myc
func (s *SmartContract) updateCptyAsset(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 25 {
		return shim.Error("Incorrect number of arguments. Expecting 25")
	}


	// 判斷是否有輸入值 
	CptyAssetAsBytes, _ := APIstub.GetState(args[24])
	CptyAsset := CptyAsset{}

	var newUSDBond, newTWDBond float64
	var newAUD,newBRL,newCAD,newCHF,newCNY,newEUR,newGBP,newHKD,newINR,newJPY,newKRW float64
    var newMOP,newMYR,newNZD,newPHP,newSEK,newSGD,newTHB,newTWD,newUSD,newZAR float64



	newUSDBond, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWDBond, err = strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newAUD, err = strconv.ParseFloat(args[3], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newBRL, err = strconv.ParseFloat(args[4], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCAD, err = strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCHF, err = strconv.ParseFloat(args[6], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCNY, err = strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEUR, err = strconv.ParseFloat(args[8], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newGBP, err = strconv.ParseFloat(args[9], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newHKD, err = strconv.ParseFloat(args[10], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newINR, err = strconv.ParseFloat(args[11], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newJPY, err = strconv.ParseFloat(args[12], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newKRW, err = strconv.ParseFloat(args[13], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newMOP, err = strconv.ParseFloat(args[14], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newMYR, err = strconv.ParseFloat(args[15], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newNZD, err = strconv.ParseFloat(args[16], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newPHP, err = strconv.ParseFloat(args[17], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newSEK, err = strconv.ParseFloat(args[18], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newSGD, err = strconv.ParseFloat(args[19], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTHB, err = strconv.ParseFloat(args[20], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTWD, err = strconv.ParseFloat(args[21], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSD, err = strconv.ParseFloat(args[22], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newZAR, err = strconv.ParseFloat(args[23], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(CptyAssetAsBytes, &CptyAsset)
	CptyAsset.ObjectType = "CptyAsset"
	CptyAsset.OwnCptyID = args[0]

	CptyAsset.USDBond = newUSDBond
	CptyAsset.TWDBond = newTWDBond
	CptyAsset.AUD = newAUD
	CptyAsset.BRL = newBRL
	CptyAsset.CAD = newCAD
	CptyAsset.CHF = newCHF
	CptyAsset.CNY = newCNY
	CptyAsset.EUR = newEUR
	CptyAsset.GBP = newGBP
	CptyAsset.HKD = newHKD
	CptyAsset.INR = newINR
	CptyAsset.JPY = newJPY
	CptyAsset.KRW = newKRW
	CptyAsset.MOP = newMOP
	CptyAsset.MYR = newMYR
	CptyAsset.NZD = newNZD
	CptyAsset.PHP = newPHP
	CptyAsset.SEK = newSEK
	CptyAsset.SGD = newSGD
	CptyAsset.THB = newTHB
	CptyAsset.TWD = newTWD
	CptyAsset.USD = newUSD
	CptyAsset.ZAR = newZAR
	
	CptyAssetAsBytes, _ = json.Marshal(CptyAsset)
	err1 := APIstub.PutState(args[24], CptyAssetAsBytes)
	if err1 != nil {
		return shim.Error("Failed to change state")
	}

	return shim.Success(nil)
}

//peer chaincode query -n mycc -c '{"Args":["queryTXIDCptyAsset", "000120181016114811"]}' -C myc
func (s *SmartContract) queryTXIDCptyAsset(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	CptyAssetAsBytes, _ := APIstub.GetState(args[0])
	CptyAsset := CptyAsset{}
	json.Unmarshal(CptyAssetAsBytes, &CptyAsset)

	CptyAssetAsBytes, err := json.Marshal(CptyAsset)
	if err != nil {
		return shim.Error("Failed to query CptyAsset state")
	}

	return shim.Success(CptyAssetAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["queryCptyAssetStatus","0001"]}' -C myc
func (s *SmartContract) queryCptyAssetStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
    var queryString string
	CptyID := args[0]
	if CptyID == "All" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CptyAsset\"}}")
	}  else {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CptyAsset\",\"OwnCptyID\":\"%s\"}}", CptyID)	
	}	

	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("esultsIterator.Close")
 
    var buffer bytes.Buffer
    buffer.WriteString("[")
 
	bArrayMemberAlreadyWritten := false
	fmt.Printf("bArrayMemberAlreadyWritten := false\n")
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
         
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
		}
		fmt.Printf("resultsIterator.HasNext\n")
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")
        buffer.WriteString(", \"Record\":")
         
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
	}
	
    buffer.WriteString("]")
 
    return shim.Success(buffer.Bytes())
}

func GetOutboundIP() string {
	itf, _ := net.InterfaceByName("ens33") //here your interface
    item, _ := itf.Addrs()
    var ip net.IP
    for _, addr := range item {
        switch v := addr.(type) {
        case *net.IPNet:
            if !v.IP.IsLoopback() {
                if v.IP.To4() != nil {//Verify if IP is IPV4
                    ip = v.IP
                }
            }
        }
    }
    if ip != nil {
        return ip.String()
    } else {
        return ""
    }
}

/*
peer chaincode invoke -n mycc -c '{"Args":["createMTMPrice", "192.168.50.196","20181210"]}' -C myc
*/
func (s *SmartContract) createMTMPrice(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// http://127.0.0.2:8080/?datadate=20181025&curr1=USD&curr2=TWD
	// ipaddress := GetOutboundIP()
	// fmt.Println("getrate.GetOutboundIP= " + ipaddress + "\n")
	
	CurrencyPair := []string{"AUDHKD","AUDTWD","AUDUSD","CADHKD","CADTWD","CHFHKD","CHFTWD","CNYHKD","CNYTWD","EURAUD","EURCNY","EURGBP","EURHKD","EURJPY","EURTWD","EURUSD","EURZAR","GBPHKD","GBPJPY","GBPTWD","GBPUSD","HKDJPY","HKDTWD","JPYTWD","MYRHKD","MYRTWD","NZDHKD","NZDTWD","NZDUSD","PHPTWD","SEKTWD","SGDHKD","SGDTWD","THBHKD","USDBRL","USDCAD","USDCHF","USDCNH","USDHKD","USDINR","USDJPY","USDKRW","USDMOP","USDMYR","USDPHP","USDSEK","USDSGD","USDTHB","USDTWD","USDZAR","ZARHKD","ZARTWD"}
    PairMTM := []string{}
    for i:=0; i < len(CurrencyPair)  ; i++ {

		queryString := "http://" + args[0] + ":8080/?datadate=" + args[1] + "&curr1=" + SubString(CurrencyPair[i],0,3) + "&curr2=" + SubString(CurrencyPair[i],3,6)
		//fmt.Println("getrate.queryString= " + queryString + "\n")
		resp, err := http.Post(queryString,"application/x-www-form-urlencoded",strings.NewReader(""))
    	if err != nil {
        	fmt.Println(err)
    	}	

    	defer resp.Body.Close()
    	body, err := ioutil.ReadAll(resp.Body)
    	if err != nil {
        	fmt.Println(err)
		}
		fmt.Println("getrate= " + CurrencyPair[i] + "=" + string(body) + "\n")
		//PairMTM =  append(PairMTM, strings.Replace(string(body)," ","",-1)) 
		PairMTM =  append(PairMTM, strings.Replace(strings.Replace(string(body),"\n","",-1)," ","",-1)) 
	}	

    TXKEY := "MTMPrice" + args[1]
	
	var newAUDHKD, newAUDTWD, newAUDUSD, newCADHKD, newCADTWD, newCHFHKD, newCHFTWD, newCNYHKD, newCNYTWD, newEURAUD float64   
	newAUDHKD, err := strconv.ParseFloat(PairMTM[0], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newAUDTWD, err = strconv.ParseFloat(PairMTM[1], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newAUDUSD, err = strconv.ParseFloat(PairMTM[2], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCADHKD, err = strconv.ParseFloat(PairMTM[3], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCADTWD, err = strconv.ParseFloat(PairMTM[4], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCHFHKD, err = strconv.ParseFloat(PairMTM[5], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCHFTWD, err = strconv.ParseFloat(PairMTM[6], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCNYHKD, err = strconv.ParseFloat(PairMTM[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newCNYTWD, err = strconv.ParseFloat(PairMTM[8], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEURAUD, err = strconv.ParseFloat(PairMTM[9], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	var newEURCNY, newEURGBP, newEURHKD, newEURJPY, newEURTWD, newEURUSD, newEURZAR, newGBPHKD, newGBPJPY, newGBPTWD float64 
	newEURCNY, err = strconv.ParseFloat(PairMTM[10], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEURGBP, err = strconv.ParseFloat(PairMTM[11], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEURHKD, err = strconv.ParseFloat(PairMTM[12], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEURJPY, err = strconv.ParseFloat(PairMTM[13], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEURTWD, err = strconv.ParseFloat(PairMTM[14], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEURUSD, err = strconv.ParseFloat(PairMTM[15], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newEURZAR, err = strconv.ParseFloat(PairMTM[16], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newGBPHKD, err = strconv.ParseFloat(PairMTM[17], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newGBPJPY, err = strconv.ParseFloat(PairMTM[18], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newGBPTWD, err = strconv.ParseFloat(PairMTM[19], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	var newGBPUSD, newHKDJPY, newHKDTWD, newJPYTWD, newMYRHKD, newMYRTWD, newNZDHKD, newNZDTWD, newNZDUSD, newPHPTWD float64   
	newGBPUSD, err = strconv.ParseFloat(PairMTM[20], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newHKDJPY, err = strconv.ParseFloat(PairMTM[21], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newHKDTWD, err = strconv.ParseFloat(PairMTM[22], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newJPYTWD, err = strconv.ParseFloat(PairMTM[23], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newMYRHKD, err = strconv.ParseFloat(PairMTM[24], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newMYRTWD, err = strconv.ParseFloat(PairMTM[25], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newNZDHKD, err = strconv.ParseFloat(PairMTM[26], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newNZDTWD, err = strconv.ParseFloat(PairMTM[27], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newNZDUSD, err = strconv.ParseFloat(PairMTM[28], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newPHPTWD, err = strconv.ParseFloat(PairMTM[29], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	var newSEKTWD, newSGDHKD, newSGDTWD, newTHBHKD, newUSDBRL, newUSDCAD, newUSDCHF, newUSDCNH, newUSDHKD, newUSDINR float64    
	newSEKTWD, err = strconv.ParseFloat(PairMTM[30], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newSGDHKD, err = strconv.ParseFloat(PairMTM[31], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newSGDTWD, err = strconv.ParseFloat(PairMTM[32], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTHBHKD, err = strconv.ParseFloat(PairMTM[33], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDBRL, err = strconv.ParseFloat(PairMTM[34], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDCAD, err = strconv.ParseFloat(PairMTM[35], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDCHF, err = strconv.ParseFloat(PairMTM[36], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDCNH, err = strconv.ParseFloat(PairMTM[37], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDHKD, err = strconv.ParseFloat(PairMTM[38], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDINR, err = strconv.ParseFloat(PairMTM[39], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	var newUSDJPY, newUSDKRW, newUSDMOP, newUSDMYR, newUSDPHP, newUSDSEK, newUSDSGD, newUSDTHB, newUSDTWD, newUSDZAR float64    
	newUSDJPY, err = strconv.ParseFloat(PairMTM[40], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDKRW, err = strconv.ParseFloat(PairMTM[41], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDMOP, err = strconv.ParseFloat(PairMTM[42], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDMYR, err = strconv.ParseFloat(PairMTM[43], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDPHP, err = strconv.ParseFloat(PairMTM[44], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDSEK, err = strconv.ParseFloat(PairMTM[45], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDSGD, err = strconv.ParseFloat(PairMTM[46], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDTHB, err = strconv.ParseFloat(PairMTM[47], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDTWD, err = strconv.ParseFloat(PairMTM[48], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUSDZAR, err = strconv.ParseFloat(PairMTM[49], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
    
    var newZARHKD, newZARTWD  float64   
	newZARHKD, err = strconv.ParseFloat(PairMTM[50], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newZARTWD, err = strconv.ParseFloat(PairMTM[51], 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	var MTMPrice = MTMPrice{ObjectType: "MTMPrice", TXKEY: TXKEY, AUDHKD: newAUDHKD,AUDTWD: newAUDTWD,AUDUSD: newAUDUSD,CADHKD: newCADHKD,CADTWD: newCADTWD,CHFHKD: newCHFHKD,CHFTWD: newCHFTWD,CNYHKD: newCNYHKD,CNYTWD: newCNYTWD,EURAUD: newEURAUD,EURCNY: newEURCNY,EURGBP: newEURGBP,EURHKD: newEURHKD,EURJPY: newEURJPY,EURTWD: newEURTWD,EURUSD: newEURUSD,EURZAR: newEURZAR,GBPHKD: newGBPHKD,GBPJPY: newGBPJPY,GBPTWD: newGBPTWD,GBPUSD: newGBPUSD,HKDJPY: newHKDJPY,HKDTWD: newHKDTWD,JPYTWD: newJPYTWD,MYRHKD: newMYRHKD,MYRTWD: newMYRTWD,NZDHKD: newNZDHKD,NZDTWD: newNZDTWD,NZDUSD: newNZDUSD,PHPTWD: newPHPTWD,SEKTWD: newSEKTWD,SGDHKD: newSGDHKD,SGDTWD: newSGDTWD,THBHKD: newTHBHKD,USDBRL: newUSDBRL,USDCAD: newUSDCAD,USDCHF: newUSDCHF,USDCNH: newUSDCNH,USDHKD: newUSDHKD,USDINR: newUSDINR,USDJPY: newUSDJPY,USDKRW: newUSDKRW,USDMOP: newUSDMOP,USDMYR: newUSDMYR,USDPHP: newUSDPHP,USDSEK:newUSDSEK,USDSGD: newUSDSGD,USDTHB: newUSDTHB,USDTWD: newUSDTWD,USDZAR: newUSDZAR,ZARHKD: newZARHKD,ZARTWD: newZARTWD}
	MTMPriceAsBytes, _ := json.Marshal(MTMPrice)
	err1 := APIstub.PutState(MTMPrice.TXKEY, MTMPriceAsBytes)
	if err1 != nil {
		return shim.Error("Failed to create state")
		fmt.Println("createMTMPrice.PutState\n") 
	}

	return shim.Success(nil)
}


/*
peer chaincode invoke -n mycc -c '{"Args":["createBondPrice", "172.20.10.13","20190515"]}' -C myc
*/
func (s *SmartContract) createBondPrice(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// http://172.20.10.13:8080/?datadate=20181121&TWDBOND=A03108
	ipaddress := GetOutboundIP()
	fmt.Println("getrate.GetOutboundIP= " + ipaddress + "\n")
	
	TWDBond := []string{"A03108","A07110","A00104","A03107","A03114"}
	USDBond := []string{"US46625HJZ47","US71647NAM11","XS11335850562","US25152RXA66","BBG00FYBLQH5"}
	BondMTM := []string{}
	
    for i:=0; i < len(TWDBond)  ; i++ {

		queryString := "http://" + args[0] + ":8080/?datadate=" + args[1] + "&TWDBOND=" + TWDBond[i]
		//fmt.Println("getrate.queryString= " + queryString + "\n")
		resp, err := http.Post(queryString,"application/x-www-form-urlencoded",strings.NewReader(""))
    	if err != nil {
        	fmt.Println(err)
    	}	

    	defer resp.Body.Close()
    	body, err := ioutil.ReadAll(resp.Body)
    	if err != nil {
        	fmt.Println(err)
		}
		fmt.Println("getrate= " + TWDBond[i] + "=" + string(body) + "\n")
		//PairMTM =  append(PairMTM, strings.Replace(string(body)," ","",-1)) 
		BondMTM =  append(BondMTM, strings.Replace(strings.Replace(string(body),"\n","",-1)," ","",-1)) 
	}	
	for i:=0; i < len(USDBond)  ; i++ {

		queryString := "http://" + args[0] + ":8080/?datadate=" + args[1] + "&USDBOND=" + USDBond[i]
		//fmt.Println("getrate.queryString= " + queryString + "\n")
		resp, err := http.Post(queryString,"application/x-www-form-urlencoded",strings.NewReader(""))
    	if err != nil {
        	fmt.Println(err)
    	}	

    	defer resp.Body.Close()
    	body, err := ioutil.ReadAll(resp.Body)
    	if err != nil {
        	fmt.Println(err)
		}
		fmt.Println("getrate= " + USDBond[i] + "=" + string(body) + "\n")
		//PairMTM =  append(PairMTM, strings.Replace(string(body)," ","",-1)) 
		BondMTM =  append(BondMTM, strings.Replace(strings.Replace(string(body),"\n","",-1)," ","",-1)) 
	}	

    TXKEY := "BondPrice" + args[1]
	
	var newA03108, newA07110, newA00104, newA03107, newA03114 float64   
	newA03108, err := strconv.ParseFloat(BondMTM[0], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newA07110, err = strconv.ParseFloat(BondMTM[1], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newA00104, err = strconv.ParseFloat(BondMTM[2], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newA03107, err = strconv.ParseFloat(BondMTM[3], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newA03114, err = strconv.ParseFloat(BondMTM[4], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	var newUS46625HJZ47, newUS71647NAM11, newXS11335850562, newUS25152RXA66, newBBG00FYBLQH5 float64   
	newUS46625HJZ47, err = strconv.ParseFloat(BondMTM[5], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUS71647NAM11, err = strconv.ParseFloat(BondMTM[6], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newXS11335850562, err = strconv.ParseFloat(BondMTM[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newUS25152RXA66, err = strconv.ParseFloat(BondMTM[8], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newBBG00FYBLQH5, err = strconv.ParseFloat(BondMTM[9], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	                            
	var BondPrice = BondPrice{ObjectType: "BondPrice", TXKEY: TXKEY, A03108: newA03108,A07110: newA07110,A00104: newA00104,A03107: newA03107,A03114: newA03114,US46625HJZ47: newUS46625HJZ47,US71647NAM11: newUS71647NAM11,XS11335850562: newXS11335850562,US25152RXA66: newUS25152RXA66,BBG00FYBLQH5: newBBG00FYBLQH5}
	BondPriceAsBytes, _ := json.Marshal(BondPrice)
	err1 := APIstub.PutState(BondPrice.TXKEY, BondPriceAsBytes)
	if err1 != nil {
		return shim.Error("Failed to create state")
		fmt.Println("createBondPrice.PutState\n") 
	}

	return shim.Success(nil)
}

func queryMTMPriceByContract(APIstub shim.ChaincodeStubInterface, TXDATE string , Contract string) (float64) {

	TXKEY := "MTMPrice" + TXDATE 
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"MTMPrice\",\"TXKEY\":\"%s\"}}", TXKEY)	

	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult" + queryString + "\n")
    if err != nil {
        return 0
    }
	defer resultsIterator.Close()
	fmt.Printf("resultsIterator.Close")
    transactionArr := []MTMPrice{}
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return 0
        }
		jsonByteObj := queryResponse.Value
		transaction := MTMPrice{}
		json.Unmarshal(jsonByteObj, &transaction)
		transactionArr = append(transactionArr, transaction)
		//var newAUDHKD, newAUDTWD, newAUDUSD, newCADHKD, newCADTWD, newCHFHKD, newCHFTWD, newCNYHKD, newCNYTWD, newEURAUD float64   
        if Contract == "AUDHKD" {
          	return transactionArr[0].AUDHKD
		} 
		if Contract == "AUDTWD" {
			return transactionArr[0].AUDTWD
		}
		if Contract == "USDTWD" {
			return transactionArr[0].USDTWD
	    } 
		if Contract == "AUDUSD" {
			return transactionArr[0].AUDUSD
		} 
		if Contract == "CADHKD" {
			return transactionArr[0].CADHKD
		} 
		if Contract == "CHFHKD" {
			return transactionArr[0].CHFHKD
		} 
		if Contract == "CHFTWD" {
			return transactionArr[0].CHFTWD
		} 
		if Contract == "CNYHKD" {
			return transactionArr[0].CNYHKD
		} 
		if Contract == "CNYTWD" {
			return transactionArr[0].CNYTWD
		} 
		if Contract == "EURAUD" {
			return transactionArr[0].EURAUD
		} 
		//2
		//var newEURCNY, newEURGBP, newEURHKD, newEURJPY, newEURTWD, newEURUSD, newEURZAR, newGBPHKD, newGBPJPY, newGBPTWD float64 
	    if Contract == "EURCNY" {
          	return transactionArr[0].EURCNY
		} 
		if Contract == "EURGBP" {
			return transactionArr[0].EURGBP
		}
		if Contract == "EURHKD" {
			return transactionArr[0].EURHKD
	    } 
		if Contract == "EURJPY" {
			return transactionArr[0].EURJPY
		} 
		if Contract == "EURTWD" {
			return transactionArr[0].EURTWD
		} 
		if Contract == "EURUSD" {
			return transactionArr[0].EURUSD
		} 
		if Contract == "EURZAR" {
			return transactionArr[0].EURZAR
		} 
		if Contract == "GBPHKD" {
			return transactionArr[0].GBPHKD
		} 
		if Contract == "GBPJPY" {
			return transactionArr[0].GBPJPY
		} 
		if Contract == "GBPTWD" {
			return transactionArr[0].GBPTWD
		} 
		//3
		//var newGBPUSD, newHKDJPY, newHKDTWD, newJPYTWD, newMYRHKD, newMYRTWD, newNZDHKD, newNZDTWD, newNZDUSD, newPHPTWD float64   
        if Contract == "GBPUSD" {
          	return transactionArr[0].GBPUSD
		} 
		if Contract == "HKDJPY" {
			return transactionArr[0].HKDJPY
		}
		if Contract == "HKDTWD" {
			return transactionArr[0].HKDTWD
	    } 
		if Contract == "JPYTWD" {
			return transactionArr[0].JPYTWD
		} 
		if Contract == "MYRHKD" {
			return transactionArr[0].MYRHKD
		} 
		if Contract == "MYRTWD" {
			return transactionArr[0].MYRTWD
		} 
		if Contract == "NZDHKD" {
			return transactionArr[0].NZDHKD
		} 
		if Contract == "NZDTWD" {
			return transactionArr[0].NZDTWD
		} 
		if Contract == "NZDUSD" {
			return transactionArr[0].NZDUSD
		} 
		if Contract == "PHPTWD" {
			return transactionArr[0].PHPTWD
		} 
		//4
		//var newSEKTWD, newSGDHKD, newSGDTWD, newTHBHKD, newUSDBRL, newUSDCAD, newUSDCHF, newUSDCNH, newUSDHKD, newUSDINR float64    
        if Contract == "SEKTWD" {
          	return transactionArr[0].SEKTWD
		} 
		if Contract == "SGDHKD" {
			return transactionArr[0].SGDHKD
		}
		if Contract == "SGDTWD" {
			return transactionArr[0].SGDTWD
	    } 
		if Contract == "THBHKD" {
			return transactionArr[0].THBHKD
		} 
		if Contract == "USDBRL" {
			return transactionArr[0].USDBRL
		} 
		if Contract == "USDCAD" {
			return transactionArr[0].USDCAD
		} 
		if Contract == "USDCHF" {
			return transactionArr[0].USDCHF
		} 
		if Contract == "USDCNH" {
			return transactionArr[0].USDCNH
		} 
		if Contract == "USDHKD" {
			return transactionArr[0].USDHKD
		} 
		if Contract == "USDINR" {
			return transactionArr[0].USDINR
		} 
		//5
		//var newUSDJPY, newUSDKRW, newUSDMOP, newUSDMYR, newUSDPHP, newUSDSEK, newUSDSGD, newUSDTHB, newUSDTWD, newUSDZAR float64    
        if Contract == "USDJPY" {
          	return transactionArr[0].USDJPY
		} 
		if Contract == "USDKRW" {
			return transactionArr[0].USDKRW
		}
		if Contract == "USDMOP" {
			return transactionArr[0].USDMOP
	    } 
		if Contract == "USDMYR" {
			return transactionArr[0].USDMYR
		} 
		if Contract == "USDPHP" {
			return transactionArr[0].USDPHP
		} 
		if Contract == "USDSEK" {
			return transactionArr[0].USDSEK
		} 
		if Contract == "USDSGD" {
			return transactionArr[0].USDSGD
		} 
		if Contract == "USDTHB" {
			return transactionArr[0].USDTHB
		} 
		if Contract == "USDTWD" {
			return transactionArr[0].USDTWD
		} 
		if Contract == "USDZAR" {
			return transactionArr[0].USDZAR
		} 

		if Contract == "ZARHKD" {
			return transactionArr[0].ZARHKD
		} 
		if Contract == "ZARTWD" {
			return transactionArr[0].ZARTWD
		} 
	}
    return 0
}

//peer chaincode query -n mycc -c '{"Args":["queryMTMPrice","20190515"]}' -C myc
func (s *SmartContract) queryMTMPrice(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	TXKEY := "MTMPrice" + args[0]
	MTMPriceAsBytes, _ := APIstub.GetState(TXKEY)
	return shim.Success(MTMPriceAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["queryBondPrice","20181126"]}' -C myc
func (s *SmartContract) queryBondPrice(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	TXKEY := "BondPrice" + args[0]
	BondPriceAsBytes, _ := APIstub.GetState(TXKEY)
	return shim.Success(BondPriceAsBytes)
}


// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(msInt/millisPerSecond,
		(msInt%millisPerSecond)*nanosPerMillisecond), nil
}

func generateMaturity(issueDate string, years int, months int, days int) (string, error) {

	t, err := msToTime(makeTimestamp(issueDate))
	if err != nil {
		return "", err
	}

	maturityDate := t.AddDate(years, months, days)
	sDate := maturityDate.Format(layout)
	return sDate, nil

}

func makeTimestamp(aDate string) string {
	var stamp int64
	t, _ := time.Parse(layout, aDate)
	stamp = t.UnixNano() / int64(time.Millisecond)
	str := strconv.FormatInt(stamp, 10)
	return str
}

func getDateUnix(mydate string) int64 {
	tm2, _ := time.Parse(layout, mydate)
	return tm2.Unix()
}

func daySub(d1, d2 string) float64 {
	t1, _ := time.Parse(layout, d1)
	t2, _ := time.Parse(layout, d2)
	return float64(timeSub(t2, t1))
}

func timeSub(t1, t2 time.Time) int {
	t1 = t1.UTC().Truncate(24 * time.Hour)
	t2 = t2.UTC().Truncate(24 * time.Hour)
	return int(t1.Sub(t2).Hours() / 24)
}

func SubString(str string, begin, length int) (substr string) {
	// 將字串轉成[]rune
	rs := []rune(str)
	lth := len(rs)

	// 範圍判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回字串
	return string(rs[begin:end])
}

func UnicodeIndex(str, substr string) int {
	// 子字串在字串的位置
	result := strings.Index(str, substr)
	if result >= 0 {
		// 取得子字串之前的字串並轉換成[]byte
		prefix := []byte(str)[0:result]
		// 將字串轉換成[]rune
		rs := []rune(string(prefix))
		// 取得rs的長度，即子字串在字串的位置
		result = len(rs)
	}

	return result
}

func round(v float64, decimals int) float64 {
	var pow float64 = 1
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int((v*pow)+0.5)) / pow
}
