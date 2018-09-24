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
	ObjectType           string          `json:"docType"`             //docType is used to distinguish the various types of objects in state database
	CptyID               string          `json:"CptyID"`
	CptyName             string          `json:"CptyName"`
	CptyStatus           string          `json:"CptyStatus"`          //Lock flag,defaut=false
}


type CptyISDA struct {
	ObjectType           string          `json:"docType"`             //docType is used to distinguish the various types of objects in state database
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
	EndDate              string          `json:"USD Cash Percentage"` //合約迄日
	USDCashPCT           float64         `json:"USDCashPCT"`          //USDCashPCT
	TWDCashPCT           float64         `json:"TWDCashPCT"`          //TWDCashPCT
	USDBondPCT           float64         `json:"USDBondPCT"`          //USDBondPCT
	TWDBondPCT           float64         `json:"TWDBondPCT"`          //TWDBondPCT
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
	
	isPutToQueue         bool             `json:"isPutToQueue"`       //isPutToQueue = true 代表資料檢核成功
	TXStatus             string           `json:"TXStatus"`           //Pending, Matched, Finished, Cancelled, PaymentError,
	CreateTime           string           `json:"createTime"`         //建立時間
	UpdateTime           string           `json:"updateTime"`         //更新時間
	TXIndex              string           `json:"TXIndex"`            //Transaction Index(全部比對)
	TXHcode              string           `json:"TXHcode"`            //Transaction Hcode(更正交易序號)
	MatchedTXID          string           `json:"MatchedTXID"`        //比對序號 OwnCptyID(後來新增的)+TXType+TimeNow 0002S20180829063021
	TXMemo               string           `json:"TXMemo"`             //交易說明
	TXErrMsg             string           `json:"TXErrMsg"`           //交易錯誤說明
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
	//CptyISDA
	} else if function == "createCptyISDA" {
		return s.createCptyISDA(APIstub, args)
	} else if function == "updateCptyISDA" {
		return s.updateCptyISDA(APIstub, args)
	} else if function == "deleteCptyISDA" {
		return s.deleteCptyISDA(APIstub, args)
	} else if function == "queryCptyISDA"  {
		return s.queryCptyISDA(APIstub, args)	
	} else if function == "queryAllCptyISDA" {
		return s.queryAllCptyISDA(APIstub, args)
	//FXTrade
	//} else if function == "createFXTrade" {
	//	return s.createFXTrade(APIstub, args)
	//} else if function == "updateFXTrade" {
	//	return s.updateFXTrade(APIstub, args)
	//} else if function == "deleteFXTrade" {
	//	return s.deleteFXTrade(APIstub, args)
	//} else if function == "queryFXTrade"  {
	//	return s.queryFXTrade(APIstub, args)	
	//} else if function == "queryAllFXTrade" {
	//	return s.queryAllFXTrade(APIstub, args)	

	} else if function == "fetchEURUSDviaOraclize" {
		return s.fetchEURUSDviaOraclize(APIstub)

	// Transaction Functions
    } else if function == "FXTradeTransfer" {
		return s.FXTradeTransfer(APIstub, args)
	} else if function == "CorrectFXTradeTransfer" {
		return s.CorrectFXTradeTransfer(APIstub, args)
	// Transaction MTM Functions
	} else if function == "FXTradeMTM" {
		return s.FXTradeMTM(APIstub, args)	


    } else if function == "queryTables" { 
		return s.queryTables(APIstub, args)
	} else if function == "queryTXIDTransactions" {
		return s.queryTXIDTransactions(APIstub, args)	
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
	} else if function == "updateQueuedTransactionHcode" {
	    return s.updateQueuedTransactionHcode(APIstub, args)
	 } else if function == "updateHistoryTransactionHcode" {
	    return s.updateHistoryTransactionHcode(APIstub, args)	
	} else {
    //map functions
		return s.mapFunction(APIstub, function, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
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

//peer chaincode invoke -n mycc -c '{"Args":["createCpty", "0001","CptyA","false"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createCpty", "0002","CptyB","false"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createCpty", "0003","CptyC","false"]}' -C myc
func (s *SmartContract) createCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	var Cpty = Cpty{ObjectType: "Cpty", CptyID: args[0], CptyName: args[1], CptyStatus: args[2]}
	CptyAsBytes, _ := json.Marshal(Cpty)
	err := APIstub.PutState(Cpty.CptyID, CptyAsBytes)
	if err != nil {
		return shim.Error("Failed to create state")
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["updateCpty", "0001","CptyAA","false"]}' -C myc
func (s *SmartContract) updateCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// 判斷是否有輸入值 

	CptyAsBytes, _ := APIstub.GetState(args[0])
	Cpty := Cpty{}

	json.Unmarshal(CptyAsBytes, &Cpty)
	Cpty.ObjectType = "Cpty"
	Cpty.CptyName = args[1]
	Cpty.CptyStatus = args[2]
	
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

	CptyAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(CptyAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["queryAllCpty","0001","9999"]}' -C myc
func (s *SmartContract) queryAllCpty(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

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

/*
peer chaincode invoke -n mycc -c '{"Args":["createCptyISDA", "0001","0002","0","0","25000000","8000000","3000000","500000","100000","2018/01/01","2020/12/31","1","0.95","0.96","0.89"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["createCptyISDA", "0001","0003","0","0","35000000","8000000","3000000","500000","100000","2018/01/01","2020/12/31","1","0.95","0.96","0.89"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["createCptyISDA", "0001","0004","0","0","45000000","8000000","3000000","500000","100000","2018/01/01","2020/12/31","1","0.95","0.96","0.89"]}' -C myc
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

//peer chaincode invoke -n mycc -c '{"Args":["createFXTrade", "000001","CptyA","CptyB","2018/01/01","2018/12/31","USD/TWD","USD","31206192","TWD","9310367.3832","29.835"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createFXTrade", "000002","CptyA","CptyB","2018/01/01","2018/12/31","USD/TWD","USD","31206192","TWD","9310367.3832","29.835"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createFXTrade", "000003","CptyA","CptyB","2018/01/01","2018/12/31","USD/TWD","USD","31206192","TWD","9310367.3832","29.835"]}' -C myc

/* func (s *SmartContract) createFXTrade(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11")
	}

	var newAmount1, newAmount2, newNetPrice float64

	newAmount1, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newAmount2, err = strconv.ParseFloat(args[9], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newNetPrice, err = strconv.ParseFloat(args[10], 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	var FXTrade = FXTrade{ObjectType: "FXTrade", TradeID: args[0], OwnCptyID: args[1], CptyID: args[2], TradeDate: args[3], MaturityDate: args[4], Contract: args[5], Curr1: args[6], Amount1: newAmount1, Curr2: args[8], Amount2: newAmount2, NetPrice: newNetPrice}
	FXTradeAsBytes, _ := json.Marshal(FXTrade)
	err1 := APIstub.PutState(FXTrade.TradeID, FXTradeAsBytes)
	if err1 != nil {
		return shim.Error("Failed to create state")
	}

	return shim.Success(nil)
} */

//peer chaincode invoke -n mycc -c '{"Args":["updateFXTrade", "000001","CptyA","CptyB","2018/01/01","2018/12/31","USD/JPY","USD","31206192","JPY","9310367.3832","29.835"]}' -C myc
/* func (s *SmartContract) updateFXTrade(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11")
	}

	// 判斷是否有輸入值 

	FXTradeAsBytes, _ := APIstub.GetState(args[0])
	FXTrade := FXTrade{}

	var newAmount1, newAmount2, newNetPrice float64

	newAmount1, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newAmount2, err = strconv.ParseFloat(args[9], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newNetPrice, err = strconv.ParseFloat(args[10], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(FXTradeAsBytes, &FXTrade)
	FXTrade.ObjectType = "FXTrade"
	FXTrade.OwnCptyID = args[1]
	FXTrade.CptyID = args[2]
	FXTrade.TradeDate = args[3]
	FXTrade.MaturityDate = args[4]
	FXTrade.Contract = args[5]
	FXTrade.Curr1 = args[6]
	FXTrade.Amount1 = newAmount1
	FXTrade.Curr2 = args[8]
	FXTrade.Amount2 = newAmount2
	FXTrade.NetPrice = newNetPrice
	
	FXTradeAsBytes, _ = json.Marshal(FXTrade)
	err1 := APIstub.PutState(args[0], FXTradeAsBytes)
	if err1 != nil {
		return shim.Error("Failed to change state")
	}

	return shim.Success(nil)
} */

//peer chaincode invoke -n mycc -c '{"Args":["deleteFXTrade", "000001"]}' -C myc
/* func (s *SmartContract) deleteFXTrade(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Delete the key from the state in ledger
	err := APIstub.DelState(args[0])
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
} */

//peer chaincode query -n mycc -c '{"Args":["queryFXTrade","000001"]}' -C myc
/* func (s *SmartContract) queryFXTrade(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	FXTradeAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(FXTradeAsBytes)
} */

//peer chaincode query -n mycc -c '{"Args":["queryAllFXTrade","000001","999999"]}' -C myc
/* func (s *SmartContract) queryAllFXTrade(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

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
} */

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
