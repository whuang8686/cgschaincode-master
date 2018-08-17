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


type Bank struct {
	ObjectType           string          `json:"docType"` //docType is used to distinguish the various types of objects in state database
	BankID               string          `json:"BankID"`
	BankName             string          `json:"BankName"`
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
	if function == "createBank" {
		return s.createBank(APIstub, args)
	} else if function == "updateBank" {
		return s.updateBank(APIstub, args)
    } else if function == "deleteBank" {
		return s.deleteBank(APIstub, args)
	} else if function == "queryBank"  {
		return s.queryBank(APIstub, args)	
	} else if function == "queryAllBank" {
		return s.queryAllBank(APIstub, args)			
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

//peer chaincode invoke -n mycc -c '{"Args":["createBank", "0001","CptyA"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createBank", "0002","CptyB"]}' -C myc
//peer chaincode invoke -n mycc -c '{"Args":["createBank", "0003","CptyC"]}' -C myc
func (s *SmartContract) createBank(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	var Bank = Bank{ObjectType: "bank", BankID: args[0], BankName: args[1]}
	BankAsBytes, _ := json.Marshal(Bank)
	err2 := APIstub.PutState(Bank.BankID, BankAsBytes)
	if err2 != nil {
		return shim.Error("Failed to create state")
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["updateBank", "0001","CptyAA"]}' -C myc
func (s *SmartContract) updateBank(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// 判斷是否有輸入值 

	BankAsBytes, _ := APIstub.GetState(args[0])
	Bank := Bank{}

	json.Unmarshal(BankAsBytes, &Bank)
	Bank.ObjectType = "bank"
	Bank.BankName = args[1]
	
	BankAsBytes, _ = json.Marshal(Bank)
	err := APIstub.PutState(args[0], BankAsBytes)
	if err != nil {
		return shim.Error("Failed to change state")
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["deleteBank", "0003"]}' -C myc
func (s *SmartContract) deleteBank(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

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

//peer chaincode query -n mycc -c '{"Args":["queryBank","0001"]}' -C myc
func (s *SmartContract) queryBank(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	BankAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(BankAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["queryAllBank","0000","9999"]}' -C myc
func (s *SmartContract) queryAllBank(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

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
