package main


import (

	"bytes"
	//"crypto/md5"
	//"crypto/sha256"
	//"encoding/hex"
	"encoding/json"
	//"errors"
	"fmt"
	"strconv"
	
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"

)

const TransactionCollateralObjectType string = "Collateral"
const CollateralTXObjectType string = "CollateralTX"

type TransactionCollateral struct {
	ObjectType   string               `json:"docType"`        // default set to "Collateral"
	TXKEY        string               `json:"TXKEY"`          // 交易日期：TXDATE(MTMYYYYMMDD)
	TXIDs        []string             `json:"TXIDs"`          // 交易序號資料
	Transactions []FXTradeCollateral  `json:"Transactions"`   // 當日交易資料
}

type FXTradeCollateral struct {
	ObjectType           string        `json:"docType"`             //docType is used to distinguish the various types of objects in state database
	TXID                 string        `json:"TXID"`                // 交易序號資料 ＝ OwnCptyID + TimeNow
	OwnCptyID            string        `json:"OwnCptyID"`
	CptyID               string        `json:"CptyID"`              //交易對手
	MTM                  float64       `json:"MTM"`       	        //(5)
	OurThreshold         int64         `json:"OwnThreshold"`        //本行門鑑金額 (4)
	CreditGuaranteeAmt   int64         `json:"CreditGuaranteeAmt"`  //信用擔保金額 (6)=(5)-(4)
	CreditGuaranteeBal   int64         `json:"CreditGuaranteeBal"`  //信用擔保餘額 (7)
	TXKinds              string        `json:"TXKinds"`             //返還/交付
	Collateral           int64         `json:"Collateral"`          //Collateral (8)=(6)-(7)
	CptyMTA              int64         `json:"CptyMTA"`             //交易對手最低轉讓金額
	MarginCall           int64         `json:"MarginCall"`          //MarginCall
}

/*
peer chaincode invoke -n mycc -c '{"Args":["FXTradeCollateral", "20181019","0001"]}' -C myc 
peer chaincode query -n mycc -c '{"Args":["queryTables","{\"selector\":{\"docType\":\"MTMTX\",\"TXKEY\":\"MTM20180928\"}}"]}' -C myc
*/
func (s *SmartContract) FXTradeCollateral(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {
	
	//TimeNow := time.Now().Format(timelayout)

	//先前除當日資料
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	TXKEY := args[0]
	OwnCptyID := args[1]
	//TXID := args[1] + TimeNow 
	datadate := "Collateral" + args[1] 
	var recint int64= 0
	var recint1 int64= 0
	var i int64= 0
	//CollateralDate := TXKEY[0:4] + "/" + TXKEY[4:6] + "/" + TXKEY[6:8]
	
	fmt.Println("CollateralDate=",datadate+"\n")
	// Delete the key from the state in ledger
	errMsg := APIstub.DelState(datadate)
	if errMsg != nil {
		return shim.Error("Failed to DelState")
	}
    //查詢本行門鑑金額
	queryString1 := fmt.Sprintf("{\"selector\": {\"docType\":\"CptyISDA\",\"OwnCptyID\":\"%s\"}}", OwnCptyID)
	fmt.Println("queryString1= " + queryString1 + "\n") 
	ownthreshold := [10]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	resultsIterator1, err := APIstub.GetQueryResult(queryString1)
	defer resultsIterator1.Close()
    if err != nil {
        return shim.Error("Failed to GetQueryResult")
	}
	transactionArr1 := []CptyISDA{}
	for resultsIterator1.HasNext() {  
		queryResponse1,err := resultsIterator1.Next()
		if err != nil {
			return shim.Error("Failed to Next")
		}

		jsonByteObj := queryResponse1.Value
		cptyisda := CptyISDA{}
		json.Unmarshal(jsonByteObj, &cptyisda)
		transactionArr1 = append(transactionArr1, cptyisda)

		fmt.Println("transactionArr[recint].CptyISDA.CptyID= " + transactionArr1[recint1].CptyID  + "\n")
		CptyID, err := strconv.ParseInt(strings.Replace(transactionArr1[recint1].CptyID,"0","",-1) ,10, 64)
   		if err != nil {
			return shim.Error("Failed to strconv.Atoi")
		}
		fmt.Println("transactionArr[recint].val.CptyID= " + strings.Replace(transactionArr1[recint1].CptyID,"0","",-1) + "\n")		
			
		ownthreshold[CptyID-1] += transactionArr1[recint1].OwnThreshold
		recint1++
	}
	fmt.Println("transactionArr[recint].ok= \n")

    //取得MTM合計
	queryString := fmt.Sprintf("{\"selector\": {\"docType\":\"MTMTX\",\"TXKEY\":\"%s\"}}", "MTM" + TXKEY)
	fmt.Println("queryString= " + queryString + "\n") 
	resultsIterator, err := APIstub.GetQueryResult(queryString)
    defer resultsIterator.Close()
    if err != nil {
        return shim.Error("Failed to GetQueryResult")
	}
	transactionArr := []TransactionMTM{}
	
    //cpty := [5]string{"0001", "0002", "0003", "0004", "0005"}
    summtm := [10]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	for resultsIterator.HasNext() {

        queryResponse,err := resultsIterator.Next()
        if err != nil {
			return shim.Error("Failed to Next")
		}
		fmt.Println("queryResponse.Key= " + queryResponse.Key + "\n") 
	
		jsonByteObj := queryResponse.Value
		transaction := TransactionMTM{}
		json.Unmarshal(jsonByteObj, &transaction)
		transactionArr = append(transactionArr, transaction)

		for key := range transaction.Transactions {
			fmt.Println("transactionArr[recint].val.OwnCptyID= " + transaction.Transactions[key].OwnCptyID  + "\n")
			fmt.Println("transactionArr[recint].val.CptyID= " + strings.Replace(transaction.Transactions[key].CptyID,"0","",-1) + "\n")		
			fmt.Println("transactionArr[recint].val.MTM= " + strconv.FormatFloat(transaction.Transactions[key].MTM ,'f', 4, 64) + "\n")
			CptyID, err := strconv.ParseInt(strings.Replace(transaction.Transactions[key].CptyID,"0","",-1) ,10, 64)
   			if err != nil {
				return shim.Error("Failed to strconv.Atoi")
   			}
			fmt.Println("transactionArr[recint].val.CptyID= " + strconv.FormatInt(CptyID-1,16) + "\n")
			if transaction.Transactions[key].OwnCptyID == OwnCptyID {
				summtm[CptyID-1] += transaction.Transactions[key].MTM 
			}
		}	
        recint++
	}	
	for i = 0; i < 10 ; i++ {
		fmt.Println("array.ownthreshold= " + strconv.FormatInt(ownthreshold[i] ,10) + "\n")
		fmt.Println("array.summtm= " + strconv.FormatFloat(summtm[i] ,'f', 4, 64) + "\n")

		//queryArgs := [][]byte{[]byte("CreateFXTradeCollateral"), []byte("20181012"), []byte("0001"), []byte("0002")}
		//peer chaincode query -n mycc -c '{"Args":["queryMTMPrice","20181012"]}' -C myc      
		if summtm[i] > 0  {
			//response := APIstub.InvokeChaincode("mycc", queryArgs, "myc")
			//response := s.CreateFXTradeCollateral(APIstub, []string{"20181010","0001","0002"})
			//if response.Status != shim.OK {
			//	errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", response.Payload)
			//	fmt.Printf(errStr)
			//	return shim.Error(errStr)
			//} 
			CptyID := fmt.Sprintf("%03d", i+1)
			if err != nil {
				return shim.Error("Failed to convert CptyID")
			}
			err = CreateFXTradeCollateral(APIstub, TXKEY, strconv.FormatInt(i, 16) , OwnCptyID , CptyID, summtm[i], ownthreshold[i])
			if err != nil {
				return shim.Error("Failed to CreateFXTradeCollateral")
			}
		}
	}

	return shim.Success(nil)
}	

//peer chaincode invoke -n mycc -c '{"Args":["CreateFXTradeCollateral", "20181012","0001","0002"]}' -C myc 
func CreateFXTradeCollateral(APIstub shim.ChaincodeStubInterface, TXKEY string, TXID string, OwnCptyID string, CptyID string, MTM float64, OurThreshold int64) error {

	TimeNow := time.Now().Format(timelayout)
	
	//if len(args) < 3 {
	//	return shim.Error("Incorrect number of arguments. Expecting 3")
	//}

	TXKEY = "Collateral" + TXKEY
	//OwnCptyID := args[1]
	//CptyID := args[2]
	TXID = OwnCptyID + TimeNow + TXID


	fmt.Println("- start CreateFXTradeCollateral ", TXKEY, OwnCptyID, CptyID, MTM, OurThreshold)
/*


	newMTM, err := strconv.ParseFloat(MTM , 64)
	if err != nil {
		fmt.Println("MTM must be a numeric string.")
	} 
	OurThreshold, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		fmt.Println("OurThreshold must be a numeric string.")
	} else if OurThreshold < 0 {
		fmt.Println("OurThreshold must be a positive value.")
	}

	CreditGuaranteeAmt, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		fmt.Println("CreditGuaranteeAmt must be a numeric string.")
	} else if CreditGuaranteeAmt < 0 {
		fmt.Println("CreditGuaranteeAmt must be a positive value.")
	}

	CreditGuaranteeBal, err := strconv.ParseFloat(args[6], 64)
	if err != nil {
		fmt.Println("CreditGuaranteeBal must be a numeric string.")
	} else if CreditGuaranteeBal < 0 {
		fmt.Println("CreditGuaranteeBal must be a positive value.")
	}

	Collateral, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		fmt.Println("Collateral must be a numeric string.")
	} else if Collateral < 0 {
		fmt.Println("Collateral must be a positive value.")
	}

	CptyMTA, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		fmt.Println("CptyMTA must be a numeric string.")
	} else if CptyMTA < 0 {
		fmt.Println("CptyMTA must be a positive value.")
	}

	MarginCall, err := strconv.ParseFloat(args[9], 64)
	if err != nil {
		fmt.Println("MarginCall must be a numeric string.")
	} else if MarginCall < 0 {
		fmt.Println("MarginCall must be a positive value.")
	}

	fmt.Println("- start CreateFXTradeCollateral ", TXKEY, TXID, OwnCptyID, CptyID, MTM)
*/
    CollateralAsBytes, err := APIstub.GetState(TXKEY)
	if CollateralAsBytes == nil {
		fmt.Println("CollateralAsBytes is null ")
	}else
	{
		fmt.Println("CollateralAsBytes is not null ")
	}

	collateralTx := TransactionCollateral{}
	json.Unmarshal(CollateralAsBytes, &collateralTx)

	if err != nil {
		return err
	}
	collateralTx.ObjectType = CollateralTXObjectType
	collateralTx.TXKEY = TXKEY

	transactionCollateral := FXTradeCollateral{}
	transactionCollateral.ObjectType = TransactionCollateralObjectType
	transactionCollateral.TXID = TXID
	transactionCollateral.OwnCptyID = OwnCptyID
	transactionCollateral.CptyID  = CptyID

	
	collateralTx.TXIDs = append(collateralTx.TXIDs, TXID)
	collateralTx.Transactions = append(collateralTx.Transactions, transactionCollateral)

	CollateralAsBytes, err1 :=json.Marshal(collateralTx)
	fmt.Println("collateralTx= " + collateralTx.TXKEY  + "\n")
	fmt.Println("collateralTx= " + TXID + "\n")
	if err != nil {
		return err
	}
	err1 = APIstub.PutState(TXKEY, CollateralAsBytes)
	if err1 != nil {
		fmt.Println("PutState.TransactionMTMsBytes= " + err1.Error() + "\n")
		return err
	}

	return nil
}



//peer chaincode query -n mycc -c '{"Args":["queryCollateralTransactionStatus","Collateral20181022","0001"]}' -C myc
func (s *SmartContract) queryCollateralTransactionStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	CollateralTXKEY := args[0]
	CptyID := args[1]

	CollateralAsBytes, _ := APIstub.GetState(CollateralTXKEY)
	collateralTx := TransactionCollateral{}
	json.Unmarshal(CollateralAsBytes, &collateralTx)

	var doflg bool
	doflg = false
	var buffer bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString("{\"CollateralTXKEY\":")
	buffer.WriteString("\"")
	buffer.WriteString(collateralTx.TXKEY)
	buffer.WriteString("\"")
	buffer.WriteString(",\"Transactions\":[")
	bArrayMemberAlreadyWritten := false
	for key, val := range collateralTx.Transactions {
		if (val.OwnCptyID == CptyID || val.CptyID == CptyID || CptyID == "All") {
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"CollateralKey\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.Itoa(key + 1))
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXID\":")
			buffer.WriteString("\"")
			buffer.WriteString(collateralTx.Transactions[key].TXID)		
			buffer.WriteString("\"")
			buffer.WriteString(", \"OwnCptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(collateralTx.Transactions[key].OwnCptyID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"CptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(collateralTx.Transactions[key].CptyID)	
			buffer.WriteString("\"")
			buffer.WriteString(", \"MTM\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(collateralTx.Transactions[key].MTM,'f', 4, 64))
			buffer.WriteString("\"")			
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
			doflg = true
		}
	}
	buffer.WriteString("]")
	if doflg != true {
		//return shim.Error("Failed to find TransactionHistory ")
		buffer.WriteString(", \"Value\":")
		buffer.WriteString("Failed to find MTMTransaction")
	}
	buffer.WriteString("}]")
	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}
