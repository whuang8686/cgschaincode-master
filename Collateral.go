package main


import (

	//"bytes"
	//"crypto/md5"
	//"crypto/sha256"
	//"encoding/hex"
	"encoding/json"
	//"errors"
	"fmt"
	"strconv"
	
	"strings"
	//"time"

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


/*
		var TXKEY,TXID string

		TXKEY = "Collateral" + SubString(TimeNow, 0, 8)
		TXID = transactionArr[recint].Transactions[0].OwnCptyID + TimeNow + strconv.FormatInt(recint,16)

		CollateralsBytes, err := APIstub.GetState(TXKEY)
		collateralTx := TransactionCollateral{}
		json.Unmarshal(CollateralAsBytes, &collateralTx)
        //新增 
		if CollateralsBytes == nil { 
			collateralTx.ObjectType = CollateralTXObjectType
			collateralTx.TXKEY = TXKEY

		   transactionCollateral := FXTradeCollateral{}
		   transactionMTM.ObjectType = TransactionCollateralObjectType
		   transactionMTM.TXID = TXID

		   transactionMTM.FXTXID = queryResponse.Key
		   transactionMTM.TXKinds = transactionArr[recint].TXKinds
		   transactionMTM.OwnCptyID = transactionArr[recint].OwnCptyID
		   transactionMTM.CptyID  = transactionArr[recint].CptyID
           transactionMTM.NetPrice = transactionArr[recint].NetPrice
		   transactionMTM.ClosePrice = 30.123 //get from API 

		   NetPrice = transactionArr[recint].NetPrice
		   ClosePrice = 30.123     //get from API
		   MTM = ClosePrice - NetPrice
		   transactionMTM.MTM = MTM
		   mtmTx.TXIDs = append(mtmTx.TXIDs, TXID)
		   mtmTx.Transactions = append(mtmTx.Transactions, transactionMTM)

		   fmt.Println("queryResponse.TXKEY= " + TXKEY + "\n") 
		   fmt.Println("queryResponse.TXID= " + TXID + "\n") 

		   TransactionMTMsBytes, err1 :=json.Marshal(mtmTx)
		   if err != nil {
				return shim.Error(err.Error())
		   }
		   err1 = APIstub.PutState(TXKEY, TransactionMTMsBytes)
		   if err1 != nil {
			   fmt.Println("PutState.TransactionMTMsBytes= " + err1.Error() + "\n")
			   return shim.Error(err1.Error())

		}
*/
        recint++
	}	
	for i := 0; i < 10 ; i++ {
		fmt.Println("array.ownthreshold= " + strconv.FormatInt(ownthreshold[i] ,10) + "\n")
		fmt.Println("array.summtm= " + strconv.FormatFloat(summtm[i] ,'f', 4, 64) + "\n")
	}

	return shim.Success(nil)
}	