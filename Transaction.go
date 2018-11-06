package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"net/http"
	"io/ioutil"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	//oraclizeapi "github.com/oraclize/fabric-api"
)

const TransactionObjectType string = "Transaction"
const QueuedTXObjectType string = "QueuedTX"
const HistoryTXObjectType string = "HistoryTX"
const TransactionMTMObjectType string = "MTM"
const MTMTXObjectType string = "MTMTX"
const timelayout string = "20060102150405"
const timelayout2 string = "2006/01/02 15:04:05"


/*
TXData1 = OwnCptyID + CptyID + TradeDate + MaturityDate + Contract + Curr1 + Amount1 
TXIndex = getSHA256(TXData1)
*/

type QueuedTransaction struct {
	ObjectType   string        `json:"docType"`        // default set to "QueuedTX"
	TXKEY        string        `json:"TXKEY"`          // 交易日期：TXDATE(YYYYMMDD)
	TXIDs        []string      `json:"TXIDs"`          // 交易序號資料
	TXIndexs     []string      `json:"TXIndexs"`       // 交易索引資料
	Transactions []FXTrade     `json:"Transactions"`   // 當日交易資料
}

type TransactionHistory struct {
	ObjectType   string        `json:"docType"`        // default set to "HistoryTX"
	TXKEY        string        `json:"TXKEY"`          // 交易日期：TXDATE(HYYYYMMDD)
	TXIDs        []string      `json:"TXIDs"`          // 交易序號資料
	TXIndexs     []string      `json:"TXIndexs"`       // 交易索引資料
	TXStatus     []string      `json:"TXStatus"`       // 交易狀態資料
	TXKinds      []string      `json:"TXKinds"`        // 交易種類資料
	Transactions []FXTrade     `json:"Transactions"`   // 當日交易資料
}

type TransactionMTM struct {
	ObjectType      string        `json:"docType"`           // default set to "MTM"
	TXKEY           string        `json:"TXKEY"`             // 交易日期：TXDATE(MTMYYYYMMDD)
	TXIDs           []string      `json:"TXIDs"`             // 交易序號資料
	TransactionsMTM []FXTradeMTM  `json:"TransactionsMTM"`   // 當日交易MTM資料
}

type FXTradeMTM struct {
	ObjectType   string        `json:"docType"`        //docType is used to distinguish the various types of objects in state database
	TXID         string        `json:"TXID"`           // 交易序號資料 ＝ OwnCptyID + TimeNow
	FXTXID       string        `json:"FXTXID"`         // FXTrade交易序號資料 ＝ OwnCptyID + TXType + TimeNow
	TXKinds      string        `json:"TXKinds"`        // Spot/FW
	OwnCptyID    string        `json:"OwnCptyID"`
	CptyID       string        `json:"CptyID"`         // 交易對手
	Contract     string        `json:"Contract"`       // 交易合約 
	NetPrice     float64       `json:"NetPrice"`       // 成交價
	ClosePrice   float64       `json:"ClosePrice"`     // 收盤價
	MTM          float64       `json:"MTM"`       	
	CreateTime   string        `json:"createTime"`     //建立時間
}

type MTMPrice struct {
	ObjectType   string        `json:"docType"`        // default set to "MTMPrice"
	TXKEY        string        `json:"TXKEY"`          // 交易日期：TXDATE(MTMPriceYYYYMMDD)
	AUDHKD       float64       `json:"AUDHKD"`         // AUD/HKD 
    AUDTWD       float64       `json:"AUDTWD"`         // AUD/TWD
    AUDUSD       float64       `json:"AUDUSD"`         // AUD/USD
    CADHKD       float64       `json:"CADHKD"`         // CAD/HKD
    CADTWD       float64       `json:"CADTWD"`         // CAD/TWD
    CHFHKD       float64       `json:"CHFHKD"`         // CHF/HKD
    CHFTWD       float64       `json:"CHFTWD"`         // CHF/TWD
    CNYHKD       float64       `json:"CNYHKD"`         // CNY/HKD
    CNYTWD       float64       `json:"CNYTWD"`         // CNY/TWD
    EURAUD       float64       `json:"EURAUD"`         // EUR/AUD
    EURCNY       float64       `json:"EURCNY"`         // EUR/CNY
    EURGBP       float64       `json:"EURGBP"`         // EUR/GBP
    EURHKD       float64       `json:"EURHKD"`         // EUR/HKD
    EURJPY       float64       `json:"EURJPY"`         // EUR/JPY
    EURTWD       float64       `json:"EURTWD"`         // EUR/TWD
    EURUSD       float64       `json:"EURUSD"`         // EUR/USD
    EURZAR       float64       `json:"EURZAR"`         // EUR/ZAR
    GBPHKD       float64       `json:"GBPHKD"`         // GBP/HKD
    GBPJPY       float64       `json:"GBPJPY"`         // GBP/JPY
    GBPTWD       float64       `json:"GBPTWD"`         // GBP/TWD
    GBPUSD       float64       `json:"GBPUSD"`         // GBP/USD
    HKDJPY       float64       `json:"HKDJPY"`         // HKD/JPY
    HKDTWD       float64       `json:"HKDTWD"`         // HKD/TWD
    JPYTWD       float64       `json:"JPYTWD"`         // JPY/TWD
    MYRHKD       float64       `json:"MYRHKD"`         // MYR/HKD
    MYRTWD       float64       `json:"MYRTWD"`         // MYR/TWD
    NZDHKD       float64       `json:"NZDHKD"`         // NZD/HKD
    NZDTWD       float64       `json:"NZDTWD"`         // NZD/TWD
    NZDUSD       float64       `json:"NZDUSD"`         // NZD/USD
    PHPTWD       float64       `json:"PHPTWD"`         // PHP/TWD
    SEKTWD       float64       `json:"SEKTWD"`         // SEK/TWD
    SGDHKD       float64       `json:"SGDHKD"`         // SGD/HKD
    SGDTWD       float64       `json:"SGDTWD"`         // SGD/TWD
    THBHKD       float64       `json:"THBHKD"`         // THB/HKD
    USDBRL       float64       `json:"USDBRL"`         // USD/BRL
    USDCAD       float64       `json:"USDCAD"`         // USD/CAD
    USDCHF       float64       `json:"USDCHF"`         // USD/CHF
    USDCNH       float64       `json:"USDCNH"`         // USD/CNH
    USDHKD       float64       `json:"USDHKD"`         // USD/HKD
    USDINR       float64       `json:"USDINR"`         // USD/INR
    USDJPY       float64       `json:"USDJPY"`         // USD/JPY
    USDKRW       float64       `json:"USDKRW"`         // USD/KRW
    USDMOP       float64       `json:"USDMOP"`         // USD/MOP
    USDMYR       float64       `json:"USDMYR"`         // USD/MYR
    USDPHP       float64       `json:"USDPHP"`         // USD/PHP
    USDSEK       float64       `json:"USDSEK"`         // USD/SEK
    USDSGD       float64       `json:"USDSGD"`         // USD/SGD
    USDTHB       float64       `json:"USDTHB"`         // USD/THB
    USDTWD       float64       `json:"USDTWD"`         // USD/TWD
    USDZAR       float64       `json:"USDZAR"`         // USD/ZAR
    ZARHKD       float64       `json:"ZARHKD"`         // ZAR/HKD
    ZARTWD       float64       `json:"ZARTWD"`         // ZAR/TWD
}

/*
peer chaincode invoke -n mycc -c '{"Args":["FXTradeSettlment", "20181230"]}' -C myc 
peer chaincode query -n mycc -c '{"Args":["queryTables","{\"selector\":{\"docType\":\"MTMTX\",\"TXKEY\":\"MTM20180928\"}}"]}' -C myc
*/
func (s *SmartContract) FXTradeSettlment(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {
	
	//TimeNow := time.Now().Format(timelayout)

	//先前除當日資料
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	TXKEY := args[0]
	datadate := "MTM" + args[0] 
	MaturityDate := TXKEY[0:4] + "/" + TXKEY[4:6] + "/" + TXKEY[6:8]
	
	fmt.Println("MaturityDate=",MaturityDate+"\n")
	// Delete the key from the state in ledger
	errMsg := APIstub.DelState(datadate)
	if errMsg != nil {
		return shim.Error("Failed to DelState")
	}

	//queryString= {"selector": {"docType":"Transaction","MaturityDate":{"$gte":"20181201"}}}
	//queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Transaction\",\"MaturityDate\":\"%s\"}}", args[0])
	
	queryString := fmt.Sprintf("{\"selector\": {\"docType\":\"Transaction\",\"MaturityDate\":\"%s\",\"TXStatus\":\"Matched\"}}", MaturityDate)

	fmt.Println("queryString= " + queryString + "\n") 
	
	resultsIterator, err := APIstub.GetQueryResult(queryString)
    defer resultsIterator.Close()
    if err != nil {
        return shim.Error("Failed to GetQueryResult")
	}
	
	transactionArr := []FXTrade{}
	var recint int64= 0
    var CptyAssetID string

	for resultsIterator.HasNext() {

        queryResponse,err := resultsIterator.Next()
        if err != nil {
			return shim.Error("Failed to Next")
		}
		fmt.Println("queryResponse.Key= " + queryResponse.Key + "\n") 
	
		jsonByteObj := queryResponse.Value
		transaction := FXTrade{}
		json.Unmarshal(jsonByteObj, &transaction)
		transactionArr = append(transactionArr, transaction)

		//讀取CptyAsset資料  
		queryString2 := fmt.Sprintf("{\"selector\": {\"docType\":\"CptyAsset\",\"OwnCptyID\":\"%s\"}}", transactionArr[recint].OwnCptyID)
        resultsIterator2, err2 := APIstub.GetQueryResult(queryString2)
    	defer resultsIterator2.Close()
    	if err2 != nil {
        	return shim.Error("Failed to GetQueryResult")
		}
	    for resultsIterator2.HasNext() {
			queryResponse2,err2 := resultsIterator2.Next()
			if err2 != nil {
				return shim.Error("Failed to Next")
			}
			fmt.Println("queryResponse2.Key= " + queryResponse2.Key + "\n") 
			CptyAssetID = queryResponse2.Key 
		}	
		fmt.Println("CptyAssetID= " + CptyAssetID  + "\n") 
		assetAsBytes, err := APIstub.GetState(CptyAssetID)
		if err != nil {
			return shim.Error(err.Error())
		}
		assetTx := CptyAsset{}
		json.Unmarshal(assetAsBytes, &assetTx)
		fmt.Println("assetAsBytes.USD= " + strconv.FormatFloat(assetTx.USD,'f', 4, 64) + "\n") 
		fmt.Println("assetAsBytes.OwnCptyID= " + assetTx.OwnCptyID + "\n") 
		
		if transactionArr[recint].TXType ==  "B" {
			if transactionArr[recint].Curr1 ==  "AUD" {
				assetTx.AUD = assetTx.AUD + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "BRL" {
				assetTx.BRL = assetTx.BRL + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "CAD" {
				assetTx.CAD = assetTx.CAD + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "CHF" {
				assetTx.CHF = assetTx.CHF + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "CNY" {
				assetTx.CNY = assetTx.CNY + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "EUR" {
				assetTx.EUR = assetTx.EUR + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "GBP" {
				assetTx.GBP = assetTx.GBP + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "HKD" {
				assetTx.HKD = assetTx.HKD + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "INR" {
				assetTx.INR = assetTx.INR + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "JPY" {
				assetTx.JPY = assetTx.JPY + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "KRW" {
				assetTx.KRW = assetTx.KRW + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "KRW" {
				assetTx.KRW = assetTx.KRW + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "MOP" {
				assetTx.MOP = assetTx.MOP + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "NZD" {
				assetTx.NZD = assetTx.NZD + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "PHP" {
				assetTx.PHP = assetTx.PHP + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "SEK" {
				assetTx.SEK = assetTx.SEK + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "SGD" {
				assetTx.SGD = assetTx.SGD + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "THB" {
				assetTx.THB = assetTx.THB + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "TWD" {
				assetTx.TWD = assetTx.TWD + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "USD" {
				assetTx.USD = assetTx.USD + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "ZAR" {
				assetTx.ZAR = assetTx.ZAR + transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "AUD" {
				assetTx.AUD = assetTx.AUD - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "BRL" {
				assetTx.BRL = assetTx.BRL - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "CAD" {
				assetTx.CAD = assetTx.CAD - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "CHF" {
				assetTx.CHF = assetTx.CHF - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "CNY" {
				assetTx.CNY = assetTx.CNY - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "EUR" {
				assetTx.EUR = assetTx.EUR - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "GBP" {
				assetTx.GBP = assetTx.GBP - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "HKD" {
				assetTx.HKD = assetTx.HKD - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "INR" {
				assetTx.INR = assetTx.INR - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "JPY" {
				assetTx.JPY = assetTx.JPY - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "KRW" {
				assetTx.KRW = assetTx.KRW - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "KRW" {
				assetTx.KRW = assetTx.KRW - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "MOP" {
				assetTx.MOP = assetTx.MOP - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "NZD" {
				assetTx.NZD = assetTx.NZD - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "PHP" {
				assetTx.PHP = assetTx.PHP - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "SEK" {
				assetTx.SEK = assetTx.SEK - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "SGD" {
				assetTx.SGD = assetTx.SGD - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "THB" {
				assetTx.THB = assetTx.THB - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "TWD" {
				assetTx.TWD = assetTx.TWD - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "USD" {
				assetTx.USD = assetTx.USD - transactionArr[recint].Amount1
			}
			if transactionArr[recint].Curr1 ==  "ZAR" {
				assetTx.ZAR = assetTx.ZAR - transactionArr[recint].Amount1
			}
		}
		assetAsBytes, err = json.Marshal(assetTx)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = APIstub.PutState(CptyAssetID, assetAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
        //更新FXTrade狀態 
		err = updateTransactionStatusbyTXID(APIstub, queryResponse.Key, "Finished")

		//更新QueuedTransaction狀態 
		err = updateQueuedTransactionStatus(APIstub, SubString(queryResponse.Key, 5, 8) , queryResponse.Key, "Finished") 
		fmt.Println("updateQueuedTransactionStatus= " + SubString(queryResponse.Key, 5, 8) + "\n") 

		//更新TransactionHistory狀態 
		err = updateHistoryTransactionStatus(APIstub, "H" + SubString(queryResponse.Key, 5, 8) , queryResponse.Key, "Finished") 		

		recint += 1 
	}	
	
	return shim.Success(nil)
}	

/*
peer chaincode invoke -n mycc -c '{"Args":["deleteFXTradeMTM", "MTM20181012"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["FXTradeMTM", "0001B20181012155013","20181012"]}' -C myc 
peer chaincode invoke -n mycc -c '{"Args":["FXTradeMTM", "0002S20181012155025","20181012"]}' -C myc 
peer chaincode query -n mycc -c '{"Args":["queryTables","{\"selector\":{\"docType\":\"MTMTX\",\"TXKEY\":\"MTM20181012\"}}"]}' -C myc
*/
func (s *SmartContract) FXTradeMTM(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {
	
	TimeNow := time.Now().Format(timelayout)

	//queryString= {"selector": {"docType":"Transaction","MaturityDate":{"$gte":"2018/10/12"}}}
    //queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Transaction\",\"MaturityDate\":\"%s\"}}", args[0])
	queryString := fmt.Sprintf("{\"selector\": {\"docType\":\"Transaction\",\"TXID\":\"%s\"}}",  args[0])

	fmt.Println("queryString= " + queryString + "\n") 
	resultsIterator, err := APIstub.GetQueryResult(queryString)
    defer resultsIterator.Close()
    if err != nil {
        return shim.Error("Failed to GetQueryResult")
	}
	transactionArr := []FXTrade{}
	var recint int64= 0

    for recint = 0; resultsIterator.HasNext(); recint++ {

		queryResponse,err := resultsIterator.Next()
        if err != nil {
			return shim.Error("Failed to Next")
		}
		fmt.Println("queryResponse.Key= " + queryResponse.Key + "\n") 
	
		jsonByteObj := queryResponse.Value
		transaction := FXTrade{}
		json.Unmarshal(jsonByteObj, &transaction)
		transactionArr = append(transactionArr, transaction)
		//fmt.Println("queryResponse.Value.NetPrice= " + strconv.FormatFloat(transactionArr[i].NetPrice, 'f', 4, 64) + "\n") 
		//計算Spot MTM
		//if transactionArr[recint].TXKinds == "SPOT" {

		   TXKEY := args[1]
		   TXID := transactionArr[recint].OwnCptyID + TimeNow 
		   FXTXID :=  queryResponse.Key
		   TXType :=  transactionArr[recint].TXType
		   TXKinds := transactionArr[recint].TXKinds
		   Contract  := strings.Replace(transactionArr[recint].Contract,"/","",-1) 
		   fmt.Println("PutState.TransactionMContractTMsBytes= " + Contract + "\n")
		   OwnCptyID := transactionArr[recint].OwnCptyID
		   CptyID := transactionArr[recint].CptyID
		   Amount1 := strconv.FormatFloat(transactionArr[recint].Amount1, 'f', 4, 64)
		   Curr2 := transactionArr[recint].Curr2
		   
		   NetPrice := strconv.FormatFloat(transactionArr[recint].NetPrice, 'f', 6, 64)
        
		   fmt.Println("PutState.TransactionMTMsBytes= " + NetPrice + "\n")
           
		   response := s.CreateFXTradeMTM(APIstub, []string{TXKEY, TXID, FXTXID, TXType, TXKinds, Contract, OwnCptyID, CptyID, Amount1, Curr2, NetPrice, strconv.FormatInt(recint, 16)})
		   // if the transfer failed break out of loop and return error
		   if response.Status != shim.OK {
			   return shim.Error("Transfer failed: " + response.Message)
		   }
		   if response.Status == shim.OK {
			   fmt.Println("response.Status\n")
		   }
		//}
	}	
	totalrecintString := strconv.FormatInt(recint,16)
	return shim.Success([]byte(totalrecintString))
}	

//peer chaincode invoke -n mycc -c '{"Args":["QueryFXTradeMTM", "20181012"]}' -C myc 
func (s *SmartContract) QueryFXTradeMTM(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {
	
	TXKEY := args[0]
	MaturityDate := TXKEY[0:4] + "/" + TXKEY[4:6] + "/" + TXKEY[6:8]

	//queryString= {"selector": {"docType":"Transaction","MaturityDate":{"$gte":"2018/10/12"}}}
    //queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Transaction\",\"MaturityDate\":\"%s\"}}", args[0])
	queryString := fmt.Sprintf("{\"selector\": {\"docType\":\"Transaction\",\"MaturityDate\":{\"$gte\":\"%s\"}}}", MaturityDate)

	fmt.Println("queryString= " + queryString + "\n") 
	resultsIterator, err := APIstub.GetQueryResult(queryString)
    defer resultsIterator.Close()
    if err != nil {
		fmt.Printf("- getQueryResultForQueryString resultsIterator error")
        return shim.Error("Failed to GetQueryResult")
    }
    // buffer is a JSON array containing QueryRecords
    var buffer bytes.Buffer
	buffer.WriteString("[")
	fmt.Printf("- getQueryResultForQueryString start buffer")
    bArrayMemberAlreadyWritten := false
    for resultsIterator.HasNext() {
        queryResponse,
        err := resultsIterator.Next()
        if err != nil {
            return shim.Error("Failed to Next")
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
	
	return shim.Success(buffer.Bytes())
}	

func (s *SmartContract) CreateFXTradeMTM(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)
	TimeNow2 := time.Now().Format(timelayout2)
	
	if len(args) < 9 {
		return shim.Error("Incorrect number of arguments. Expecting 9")
	}

	TXKEY := "MTM" + args[0] 
	FXTXID := args[2]
	TXType := args[3]
	TXKinds := args[4]
	Contract := args[5]
	OwnCptyID := args[6]
	CptyID := args[7]
	Amount1, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		fmt.Println("Amount1 must be a numeric string.")
	} else if Amount1 < 0 {
		fmt.Println("Amount1 must be a positive value.")
	}
    Curr2 := args[9]
	NetPrice, err := strconv.ParseFloat(args[10], 64)
	if err != nil {
		fmt.Println("NetPrice must be a numeric string.")
	} else if NetPrice < 0 {
		fmt.Println("NetPrice must be a positive value.")
	}

	TXID := args[6] + TimeNow + args[11]

	fmt.Println("- start CreateFXTradeMTM ", TXKEY, TXID, FXTXID, TXKinds, OwnCptyID, CptyID, NetPrice)
	
	MTMAsBytes, err := APIstub.GetState(TXKEY)
	if MTMAsBytes == nil {
		fmt.Println("MTMAsBytes is null ")
	}else
	{
		fmt.Println("MTMAsBytes is not null ")
	}

	mtmTx := TransactionMTM{}
	json.Unmarshal(MTMAsBytes, &mtmTx)

	if err != nil {
		return shim.Error(err.Error())
	}
	mtmTx.ObjectType = TransactionMTMObjectType
	mtmTx.TXKEY = TXKEY

	transactionMTM := FXTradeMTM{}
	transactionMTM.ObjectType = MTMTXObjectType
	transactionMTM.TXID = TXID
	transactionMTM.FXTXID = FXTXID
	transactionMTM.TXKinds = TXKinds
	transactionMTM.Contract = Contract
	transactionMTM.OwnCptyID = OwnCptyID
	transactionMTM.CptyID  = CptyID
	transactionMTM.NetPrice = NetPrice
	ClosePrice := queryMTMPriceByContract(APIstub, args[0] ,  Contract)
	transactionMTM.ClosePrice = ClosePrice
    fmt.Println("ClosePrice= " + strconv.FormatFloat(ClosePrice,'f', 4, 64)   + "\n")
	NetPrice = NetPrice
	//Step1 : Amount1 * (收盤Forward Rate - 交易NetPrice) USD/TWD OwnCptyID的方向 
	var MTM float64= 0.00
	if TXType == "S" {
		MTM = Amount1 * (NetPrice - ClosePrice)
	} else if TXType == "B" {
		MTM = Amount1 * (ClosePrice - NetPrice)
	}	
	
	//Step2 : TWD  / (USD/TWD) = USD XXX MTM計算都轉成USD
	fmt.Println("Step2= " + "USD" + SubString(Contract,3,6)   + "\n")
	if Curr2 != "USD" {
		MTM = MTM / queryMTMPriceByContract(APIstub, args[0] ,  "USD" + SubString(Contract,3,6))
	}
	transactionMTM.MTM = MTM
	transactionMTM.CreateTime = TimeNow2
	mtmTx.TXIDs = append(mtmTx.TXIDs, TXID)
	mtmTx.TransactionsMTM = append(mtmTx.TransactionsMTM, transactionMTM)

	MTMAsBytes, err1 :=json.Marshal(mtmTx)
	fmt.Println("mtmTx= " + mtmTx.TXKEY  + "\n")
	fmt.Println("mtmTx= " + TXID + "\n")
	if err != nil {
		 return shim.Error(err.Error())
	}
	err1 = APIstub.PutState(TXKEY, MTMAsBytes)
	if err1 != nil {
		fmt.Println("PutState.TransactionMTMsBytes= " + err1.Error() + "\n")
		return shim.Error(err1.Error())
	}

	return shim.Success(nil)
}


//peer chaincode invoke -n mycc -c '{"Args":["deleteFXTradeMTM", "MTM20181012"]}' -C myc
func (s *SmartContract) deleteFXTradeMTM(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	startKey := args[0] 
	endKey := args[0] + "9999999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
    defer resultsIterator.Close()
    if err != nil {
        return shim.Error("Failed to GetQueryResult")
	}
	for resultsIterator.HasNext() {
		queryResponse,err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Failed to Next")
		}
		fmt.Println("queryResponse.Key= " + queryResponse.Key + "\n") 
		err = APIstub.DelState(queryResponse.Key)
		if err != nil {
			return shim.Error("Failed to delete state")
		}
	}

	

	// Delete the key from the state in ledger
	
	return shim.Success(nil)
}


/*
peer chaincode invoke -n mycc -c '{"Args":["submitApproveTransaction", "BANK004B00400000000120180415070724","0","BANKCBC"]}' -C myc -v 9.0
peer chaincode invoke -n mycc -c '{"Args":["submitApproveTransaction", "BANK002S00200000000120180415065316","0","BANKCBC"]}' -C myc -v 9.0

*/

/* func (s *SmartContract) submitApproveTransaction(
	APIstub shim.ChaincodeStubInterface,
	args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)

	err := checkArgArrayLength(args, 2)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(args[0]) <= 0 {
		return shim.Error("TXID must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("Admin must be a non-empty string")
	}
	//BK004S00400000000120180610041355
	TXID := strings.ToUpper(args[0])
	TXKEY := SubString(TimeNow, 0, 8)
	HTXKEY := "H" + SubString(TimeNow, 0, 8)
	TXDAY := SubString(TXID, 18, 8)
	if TXDAY < TXKEY {
		TXKEY = TXDAY
		HTXKEY = "H" + TXKEY
	}

	ApproveFlag := approved0
	ValueAsBytes, err := APIstub.GetState("approveflag")
	if err == nil {
		ApproveFlag = string(ValueAsBytes)
	}
	fmt.Printf("1.ApproveFlag=%s\n", ApproveFlag)
	transaction, err := getTransactionStructFromID(stub, TXID)
	if err != nil {
		return shim.Error("TXID transacton does not found.")
	}

	MatchedTXID := transaction.MatchedTXID
	SecurityID := transaction.SecurityID
	SecurityAmount := transaction.SecurityAmount
	Payment := transaction.Payment
	TXType := transaction.TXType
	TXFrom := transaction.TXFrom
	TXTo := transaction.TXTo
	BankFrom := transaction.BankFrom
	BankTo := transaction.BankTo

	isApproved := true
	NewStatus := "Finished"
	if ApproveFlag == approved0 {
		isApproved = true
		NewStatus = "Finished"
	} else if ApproveFlag == approved1 {
		isApproved = true
		NewStatus = "Waiting4Payment"
	} else if ApproveFlag == approved2 {
		isApproved = true
		NewStatus = "PaymentError"
	} else if ApproveFlag == approved21 {
		isApproved = false
		NewStatus = "Cancelled1"
	} else if ApproveFlag == approved22 {
		isApproved = true
		NewStatus = "Finished"
	} else if ApproveFlag == approved3 {
		isApproved = false
		NewStatus = "Cancelled2"
	} else if ApproveFlag == approved5 {
		_, _, securityamount, _, errMsg := checkAccountBalance(stub, SecurityID, Payment, SecurityAmount, TXFrom, TXType, -1)
		fmt.Printf("0-1.Account securityamount=%s\n", securityamount)
		fmt.Printf("0-2.Transaction SecurityAmount=%s\n", SecurityAmount)
		fmt.Printf("0-3.Approved errMsg=%s\n", errMsg)
		if errMsg != "" {
			isApproved = false
			NewStatus = "Cancelled2"
		} else if securityamount < SecurityAmount {
			isApproved = true
			NewStatus = "PaymentError"
		} else {
			isApproved = true
			NewStatus = "Finished"
		}
	}

	fmt.Printf("1.Approved TXID=%s\n", TXID)
	fmt.Printf("2.Approved MatchedTXID=%s\n", MatchedTXID)
	fmt.Printf("3.Approved TXKEY=%s\n", TXKEY)
	fmt.Printf("4.Approved HTXKEY=%s\n", HTXKEY)

	if isApproved != true {
		err := updateQueuedTransactionApproveStatus(stub, TXKEY, TXID, MatchedTXID, NewStatus)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = updateHistoryTransactionApproveStatus(stub, HTXKEY, TXID, MatchedTXID, NewStatus)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = updateTransactionStatus(stub, TXID, NewStatus, MatchedTXID)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = updateTransactionStatus(stub, MatchedTXID, NewStatus, TXID)
		if err != nil {
			return shim.Error(err.Error())
		}

		if TXType == "S" {
			senderBalance, receiverBalance, senderPendingBalance, receiverPendingBalance, err := resetAccountBalance(stub, SecurityID, SecurityAmount, Payment, TXFrom, TXTo)
			senderBalance, receiverBalance, err = resetSecurityAmount(stub, SecurityID, Payment, SecurityAmount, TXFrom, TXTo)
			if BankFrom != BankTo {
				err = updateBankTotals(stub, TXFrom, SecurityID, TXFrom, Payment, SecurityAmount, false)
				if err != nil {
					return shim.Error(err.Error())
				}
				err = updateBankTotals(stub, TXTo, SecurityID, TXTo, Payment, SecurityAmount, true)
				if err != nil {
					return shim.Error(err.Error())
				}
			}
			if (senderBalance < 0) || (receiverBalance < 0) || (senderPendingBalance < 0) || (receiverPendingBalance < 0) {
				return shim.Error("senderBalance,receiverBalance,senderPendingBalance,receiverPendingBalance <0")
			}

		}
		if TXType == "B" {
			senderBalance, receiverBalance, senderPendingBalance, receiverPendingBalance, err := resetAccountBalance(stub, SecurityID, SecurityAmount, Payment, TXTo, TXFrom)
			senderBalance, receiverBalance, err = resetSecurityAmount(stub, SecurityID, Payment, SecurityAmount, TXTo, TXFrom)
			if BankFrom != BankTo {
				err = updateBankTotals(stub, TXFrom, SecurityID, TXFrom, Payment, SecurityAmount, true)
				if err != nil {
					return shim.Error(err.Error())
				}
				err = updateBankTotals(stub, TXTo, SecurityID, TXTo, Payment, SecurityAmount, false)
				if err != nil {
					return shim.Error(err.Error())
				}
			}
			if (senderBalance < 0) || (receiverBalance < 0) || (senderPendingBalance < 0) || (receiverPendingBalance < 0) {
				return shim.Error("senderBalance,receiverBalance,senderPendingBalance,receiverPendingBalance <0")
			}

		}

	} else if isApproved == true {
		fmt.Printf("5.Approved TXID=%s\n", TXID)
		fmt.Printf("6.Approved MatchedTXID=%s\n", MatchedTXID)
		fmt.Printf("7.Approved TXKEY=%s\n", TXKEY)
		fmt.Printf("8.Approved TXKEY=%s\n", HTXKEY)
		err := updateQueuedTransactionApproveStatus(stub, TXKEY, TXID, MatchedTXID, NewStatus)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = updateHistoryTransactionApproveStatus(stub, HTXKEY, TXID, MatchedTXID, NewStatus)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = updateTransactionStatus(stub, TXID, NewStatus, MatchedTXID)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = updateTransactionStatus(stub, MatchedTXID, NewStatus, TXID)
		if err != nil {
			return shim.Error(err.Error())
		}

	}

	return shim.Success(nil)

} */

//peer chaincode invoke -n mycc -c '{"Args":["submitEndDayTransaction", "BANK004S00400000000120180414121032" , "BANKCBC" ]}' -C myc -v 1.0

/* func (s *SmartContract) submitEndDayTransaction(
	APIstub shim.ChaincodeStubInterface,
	args []string) peer.Response {
	//var MatchedTXID string
	//MatchedTXID = ""
	TimeNow := time.Now().Format(timelayout)

	err := checkArgArrayLength(args, 2)
	if err != nil {
		return shim.Error(err.Error())
	}

	if len(args[0]) <= 0 {
		return shim.Error("TXID  must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("Admin  must be a non-empty string")
	}

	TXID := strings.ToUpper(args[0])
	//Admin := strings.ToUpper(args[1])
	//TimeNow := time.Now().Format(timelayout)
	TXKEY := SubString(TimeNow, 0, 8)
	HTXKEY := "H" + SubString(TimeNow, 0, 8)
	TXDAY := SubString(TXID, 18, 8)
	if TXDAY < TXKEY {
		TXKEY = TXDAY
		HTXKEY = "H" + TXKEY
	}

	MatchedTXID, err2 := updateEndDayTransactionStatus(stub, TXID)
	if err2 != nil {
		return shim.Error(err2.Error())
	}

	err = updateEndDayQueuedTransactionStatus(stub, TXKEY, TXID, MatchedTXID)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = updateEndDayHistoryTransactionStatus(stub, HTXKEY, TXID, MatchedTXID)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)

} */

/* FXTrade交易比對
peer chaincode invoke -n mycc -c '{"Args":["FXTradeTransfer", "B","0001","0002","2018/01/01","2018/12/30","USD/TWD","USD","1000000","TWD","1000000","26","SPOT","true"]}' -C myc 
peer chaincode invoke -n mycc -c '{"Args":["FXTradeTransfer", "S","0002","0001","2018/01/01","2018/12/30","USD/TWD","USD","1000000","TWD","1000000","30","SPOT","true"]}' -C myc 
*/

func (s *SmartContract) FXTradeTransfer(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)

	newTX, isPutInQueue, errMsg := validateTransaction(APIstub, args)

	if errMsg != "" {
		//return shim.Error(err.Error())
		fmt.Println("errMsg= " + errMsg + "\n")
		newTX.TXErrMsg = errMsg
		newTX.TXStatus = "Cancelled"
		newTX.TXMemo = "交易被取消"
	}
	TXID := newTX.TXID
	TXIndex := newTX.TXIndex
	TXType := newTX.TXType  
	OwnCptyID := newTX.OwnCptyID
	//CptyID := newTX.CptyID
	//TradeDate := newTX.TradeDate
	//MaturityDate := newTX.MaturityDate
	//Contract := newTX.Contract
	//Curr1 := newTX.Curr1
	//Amount1 := newTX.Amount1
	TXStatus := newTX.TXStatus

	var doflg bool
	var TXKinds string
	doflg = false
	TXKEY := SubString(TimeNow, 0, 8) //20060102150405
	HTXKEY := "H" + TXKEY
	TXDAY := SubString(TXID, 5, 8)
	if TXDAY < TXKEY {
		TXKEY = TXDAY
		HTXKEY = "H" + TXKEY
	}

	//ApproveFlag := approved0
	//ValueAsBytes, err := APIstub.GetState("approveflag")
	//if err == nil {
	//	ApproveFlag = string(ValueAsBytes)
	//}

	fmt.Printf("1.isPutInQueue=%s\n", isPutInQueue)

	if isPutInQueue == true {
		newTX.isPutToQueue = true
		queueAsBytes, err := APIstub.GetState(TXKEY)
		fmt.Println("isPutInQueue == true=" + TXKEY + "\n")
		if err != nil {
			//return shim.Error(err.Error())
			fmt.Println("1.queueAsBytes= " + err.Error() + "\n")
			newTX.TXErrMsg = TXKEY + ":QueueID does not exits."
			newTX.TXStatus = "Cancelled"
			newTX.TXMemo = "交易被取消"
		}
		queuedTx := QueuedTransaction{}
		json.Unmarshal(queueAsBytes, &queuedTx)

		historyAsBytes, err := APIstub.GetState(HTXKEY)
		if err != nil {
			fmt.Println("2.historyAsBytes,errMsg= " + err.Error() + "\n")
			newTX.TXErrMsg = HTXKEY + ":HistoryID does not exits."
			newTX.TXStatus = "Cancelled"
			newTX.TXMemo = "交易被取消"
		}
		historyNewTX := TransactionHistory{}
		json.Unmarshal(historyAsBytes, &historyNewTX)

		if queueAsBytes == nil {
			fmt.Println("queueAsBytes == nil,queueAsBytes\n")
			queuedTx.ObjectType = QueuedTXObjectType
			queuedTx.TXKEY = TXKEY
			queuedTx.Transactions = append(queuedTx.Transactions, newTX)
			queuedTx.TXIndexs = append(queuedTx.TXIndexs, TXIndex)
			queuedTx.TXIDs = append(queuedTx.TXIDs, TXID)
			if historyAsBytes == nil {
				fmt.Println("historyAsBytes\n")
				historyNewTX.ObjectType = HistoryTXObjectType
				historyNewTX.TXKEY = HTXKEY
				historyNewTX.Transactions = append(historyNewTX.Transactions, newTX)
				historyNewTX.TXIndexs = append(historyNewTX.TXIndexs, TXIndex)
				historyNewTX.TXIDs = append(historyNewTX.TXIDs, TXID)
				historyNewTX.TXStatus = append(historyNewTX.TXStatus, newTX.TXStatus)
				historyNewTX.TXKinds = append(historyNewTX.TXKinds, TXKinds)
			}
		} else if queueAsBytes != nil {
			fmt.Println("queueAsBytes != nil,queueAsBytes\n")
			for key, val := range queuedTx.Transactions {
				fmt.Println("key " + val.TXID + "\n")
				if val.TXIndex == TXIndex && val.TXStatus == TXStatus && val.OwnCptyID != OwnCptyID && val.TXType != TXType && val.TXID != TXID {
					fmt.Println("1.TXIndex= " + TXIndex + "\n")
					fmt.Println("2.OwnCptyID= " + OwnCptyID + "\n")
					fmt.Println("3.TXType= " + TXType + "\n")
					fmt.Println("4.TXID= " + TXID + "\n")
					fmt.Println("5.val.TXID= " + val.TXID + "\n")
					fmt.Println("6.TXStatus= " + TXStatus + "\n")
					fmt.Println("7.val.TXStatus= " + val.TXStatus + "\n")

					if TXStatus == "Pending" && val.TXStatus == "Pending" {
						if doflg == true {
							//return shim.Error("doflg eq to true")
							newTX.TXErrMsg = "doflg can not equle to true."
							newTX.TXStatus = "Cancelled"
							newTX.TXMemo = "交易被取消"
							break
						}
						newTX.MatchedTXID = val.TXID
						queuedTx.Transactions[key].MatchedTXID = TXID
						historyNewTX.Transactions[key].MatchedTXID = TXID
						err = updateTransactionStatus(APIstub, val.TXID, "Matched", TXID)
						if err != nil {
							//return shim.Error(err.Error())
							fmt.Println("updateTransactionStatus,err=\n")
							newTX.TXErrMsg = "Failed to execute updateTransactionStatus = Matched."
							newTX.TXStatus = "Cancelled"
							newTX.TXMemo = "交易被取消"
							break
						}
						queuedTx.Transactions[key].TXStatus = "Matched"
						historyNewTX.Transactions[key].TXStatus = "Matched"
						historyNewTX.TXStatus[key] = "Matched"
						newTX.TXStatus = "Matched"
						queuedTx.Transactions[key].TXMemo = ""
						historyNewTX.Transactions[key].TXMemo = ""
						newTX.TXMemo = ""
						
						doflg = true
						break
					}
				} 
			}
		}
		if queueAsBytes != nil {
			if historyAsBytes != nil {
				queuedTx.ObjectType = QueuedTXObjectType
				queuedTx.TXKEY = TXKEY
				queuedTx.Transactions = append(queuedTx.Transactions, newTX)
				queuedTx.TXIndexs = append(queuedTx.TXIndexs, TXIndex)
				queuedTx.TXIDs = append(queuedTx.TXIDs, TXID)

				historyNewTX.ObjectType = HistoryTXObjectType
				historyNewTX.TXKEY = HTXKEY
				historyNewTX.Transactions = append(historyNewTX.Transactions, newTX)
				historyNewTX.TXIndexs = append(historyNewTX.TXIndexs, TXIndex)
				historyNewTX.TXIDs = append(historyNewTX.TXIDs, TXID)
				historyNewTX.TXStatus = append(historyNewTX.TXStatus, newTX.TXStatus)
				historyNewTX.TXKinds = append(historyNewTX.TXKinds, TXKinds)
			}
		}		
		QueuedAsBytes, err := json.Marshal(queuedTx)
		fmt.Println("PutState.QueuedAsBytes=ok\n")
		err = APIstub.PutState(TXKEY, QueuedAsBytes)
		if err != nil {
			fmt.Println("PutState.QueuedAsBytes= " + err.Error() + "\n")
			return shim.Error(err.Error())
		}
		historyAsBytes, err = json.Marshal(historyNewTX)
		fmt.Println("PutState.historyAsBytes=ok\n")
		err = APIstub.PutState(HTXKEY, historyAsBytes)
		if err != nil {
			fmt.Println("PutState.historyAsBytes= " + err.Error() + "\n")
			return shim.Error(err.Error())
		}
	}
	fmt.Println("newTX.TXID= " + newTX.TXID + "\n")
	TransactionAsBytes, err := json.Marshal(newTX)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutState(TXID, TransactionAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}


func validateTransaction(APIstub shim.ChaincodeStubInterface,args []string) (FXTrade, bool, string) {
	TimeNow := time.Now().Format(timelayout)
	TimeNow2 := time.Now().Format(timelayout2)
	var err error
	var TXData,  TXIndex, TXID string
	
	transaction := FXTrade{}
	transaction.ObjectType = TransactionObjectType
	transaction.TXStatus = "Cancelled"
	transaction.TXMemo = "尚未比對"
	transaction.TXErrMsg = ""
	transaction.TXHcode = ""
	transaction.CreateTime = TimeNow2
	transaction.UpdateTime = TimeNow2

	err = checkArgArrayLength(args, 13)
	if err != nil {
		return transaction, false, "The args-length must be 13."
	}
	if len(args[0]) <= 0 {
		return transaction, false, "TXType must be a non-empty string."
	}
	if len(args[1]) <= 0 {
		return transaction, false, "OwnCptyID must be a non-empty string."
	}
	if len(args[2]) <= 0 {
		return transaction, false, "CptyID must be a non-empty string."
	}
	if len(args[3]) <= 0 {
		return transaction, false, "TradeDate must be a non-empty string."
	}
	if len(args[4]) <= 0 {
		return transaction, false, "MaturityDate must be a non-empty string."
	}
	if len(args[5]) <= 0 {
		return transaction, false, "Contract must be a non-empty string."
	}
	if len(args[6]) <= 0 {
		return transaction, false, "Curr1 must be a non-empty string."
	}
	if len(args[7]) <= 0 {
		return transaction, false, "Amount1 must be a non-empty string."
	}
	if len(args[8]) <= 0 {
		return transaction, false, "Curr2 must be a non-empty string."
	}
	if len(args[9]) <= 0 {
		return transaction, false, "Amount2 must be a non-empty string."
	}
    if len(args[10]) <= 0 {
		return transaction, false, "NetPrice must be a non-empty string."
	}
	if len(args[11]) <= 0 {
		return transaction, false, "TXKinds must be a non-empty string."
	}
	if len(args[12]) <= 0 {
		return transaction, false, "isPutToQueue flag must be a non-empty string."
	}

	isPutToQueue, err := strconv.ParseBool(strings.ToLower(args[12]))
	if err != nil {
		return transaction, false, "isPutToQueue must be a boolean string."
	}
	transaction.isPutToQueue = isPutToQueue
	TXType := SubString(strings.ToUpper(args[0]), 0, 1)
	if (TXType != "B") && (TXType != "S") {
		return transaction, false, "TXType must be a B or S."
	}

	transaction.TXType = TXType
	OwnCptyID := strings.ToUpper(args[1])
	CptyID := strings.ToUpper(args[2])
	//if verifyIdentity(stub, OwnCptyID) != "" {
	//	return transaction, false, "OwnCptyID does not exits in the CptyList."
	//}
	//if verifyIdentity(stub, CptyID) != "" {
	//	return transaction, false, "CptyID does not exits in the CptyList."
	//}
	TXID = OwnCptyID + TXType + TimeNow
	transaction.TXID = TXID

	if OwnCptyID == CptyID {
		return transaction, false, "OwnCptyID can not equal to CptyID."
	}
	transaction.OwnCptyID = OwnCptyID
	transaction.CptyID = CptyID
	TradeDate := args[3]
	transaction.TradeDate = TradeDate

	MaturityDate := args[4]
	transaction.MaturityDate = MaturityDate

	Contract := strings.ToUpper(args[5])
	transaction.Contract = Contract

	Curr1 := strings.ToUpper(args[6])
	transaction.Curr1 = Curr1

	Amount1, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return transaction, false, "Amount1 must be a numeric string."
	} else if Amount1 < 0 {
		return transaction, false, "Amount1 must be a positive value"
	}
	transaction.Amount1 = Amount1

	Curr2 := strings.ToUpper(args[8])
	transaction.Curr2 = Curr2

	Amount2, err := strconv.ParseFloat(args[9], 64)
	if err != nil {
		return transaction, false, "Amount2 must be a numeric string."
	} else if Amount2 < 0 {
		return transaction, false, "Amount2 must be a positive value"
	}
	transaction.Amount2 = Amount2

	NetPrice, err := strconv.ParseFloat(args[10], 64)
	if err != nil {
		return transaction, false, "NetPrice must be a numeric string."
	} else if NetPrice < 0 {
		return transaction, false, "NetPrice must be a positive value"
	}
	transaction.NetPrice = NetPrice
    transaction.TXKinds = strings.ToUpper(args[11])
	//TXData = OwnCptyID + CptyID + TradeDate + MaturityDate + Contract + Curr1 + Amount1 
	if TXType == "S" {
		TXData = OwnCptyID + CptyID + TradeDate + MaturityDate + Contract + Curr1 + strconv.FormatFloat(Amount1, 'f', 4, 64)
		TXIndex = getSHA256(TXData)
	}
	fmt.Println("6.validateTransaction= " + TimeNow + "\n")
	if TXType == "B" {
		TXData = CptyID + OwnCptyID + TradeDate + MaturityDate + Contract + Curr1 + strconv.FormatFloat(Amount1, 'f', 4, 64)
		TXIndex = getSHA256(TXData)
	}
	transaction.TXIndex = TXIndex
	transaction.TXStatus = "Pending"
	return transaction, true, ""

}

func getTransactionStructFromID(APIstub shim.ChaincodeStubInterface,TXID string) (*FXTrade, error) {

	var errMsg string
	newTX := &FXTrade{}
	TXAsBytes, err := APIstub.GetState(TXID)
	if err != nil {
		return newTX, err
	} else if TXAsBytes == nil {
		errMsg = fmt.Sprintf("Error: Transaction ID does not exist: %s", TXID)
		return newTX, errors.New(errMsg)
	}
	err = json.Unmarshal([]byte(TXAsBytes), newTX)
	if err != nil {
		return newTX, err
	}
	return newTX, nil
}

func getQueueStructFromID(
	APIstub shim.ChaincodeStubInterface,
	TXKEY string) (*QueuedTransaction, error) {

	var errMsg string
	queue := &QueuedTransaction{}
	queueAsBytes, err := APIstub.GetState(TXKEY)
	if err != nil {
		return nil, err
	} else if queueAsBytes == nil {
		errMsg = fmt.Sprintf("Error: QueuedTransaction ID does not exist: %s", TXKEY)
		return nil, errors.New(errMsg)
	}
	err = json.Unmarshal([]byte(queueAsBytes), queue)
	if err != nil {
		return nil, err
	}
	return queue, nil
}

func getHistoryTransactionStructFromID(
	APIstub shim.ChaincodeStubInterface,
	TXKEY string) (*TransactionHistory, error) {

	var errMsg string
	newTX := &TransactionHistory{}
	TXAsBytes, err := APIstub.GetState(TXKEY)
	if err != nil {
		return newTX, err
	} else if TXAsBytes == nil {
		errMsg = fmt.Sprintf("Error: TXKEY does not exist: %s", TXKEY)
		return newTX, errors.New(errMsg)
	}
	err = json.Unmarshal([]byte(TXAsBytes), newTX)
	if err != nil {
		return newTX, err
	}
	return newTX, nil
}

func getQueueArrayFromQuery(
	APIstub shim.ChaincodeStubInterface) ([]QueuedTransaction, int, error) {
	TimeNow := time.Now().Format(timelayout)

	//startKey := "20180000" //20180326
	//endKey := "20181231"
	var doflg bool
	var sumLen int
	doflg = false
	sumLen = 0
	TXKEY := SubString(TimeNow, 0, 8)

	resultsIterator, err := APIstub.GetStateByRange(TXKEY, TXKEY)
	if err != nil {
		return nil, sumLen, err
	}
	defer resultsIterator.Close()

	queueArr := []QueuedTransaction{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, sumLen, err
		}
		jsonByteObj := queryResponse.Value
		queue := QueuedTransaction{}
		for key, val := range queue.Transactions {
			if val.TXStatus != "" {
				json.Unmarshal(jsonByteObj, &queue)
				queueArr = append(queueArr, queue)
				doflg = true
				sumLen = key

			}
		}
	}
	if doflg != true {
		return nil, sumLen, err
	}
	return queueArr, sumLen, nil
}

func getSHA256(myData string) string {
	// ID generation
	moveOutInFundID := sha256.New()
	moveOutInFundID.Write([]byte(myData))
	moveOutInFundIDString := fmt.Sprintf("%x", moveOutInFundID.Sum(nil))
	return moveOutInFundIDString
}

func getMD5Str(myData string) string {
	h := md5.New()
	h.Write([]byte(myData))
	cipherStr := h.Sum(nil)
	encodeStr := hex.EncodeToString(cipherStr)
	return encodeStr
}

func updateTransactionStatus(APIstub shim.ChaincodeStubInterface, TXID string, TXStatus string, MatchedTXID string) error {
	fmt.Printf("1 updateTransactionStatus TXID = %s, TXStatus = %s, MatchedTXID = %s\n", TXID, TXStatus, MatchedTXID)
	TimeNow2 := time.Now().Format(timelayout2)
	transaction, err := getTransactionStructFromID(APIstub, TXID)
	transaction.TXStatus = TXStatus
	var TXMemo, TXErrMsg string
	TXMemo = ""
	TXErrMsg = ""

	if TXStatus == "Waiting4Payment" {
		TXMemo = "等待回應"
		transaction.TXMemo = TXMemo
		transaction.TXStatus = "Waiting4Payment"
	}
	if TXStatus == "PaymentError" {
		TXMemo = "款不足等待補款"
		transaction.TXMemo = TXMemo
		transaction.TXStatus = "PaymentError"
	}
	if TXStatus == "Cancelled" {
		TXMemo = "交易取消"
		transaction.TXMemo = TXMemo
		transaction.TXStatus = "Cancelled"
	}
	if TXStatus == "Cancelled1" {
		TXMemo = "交易取消"
		transaction.TXMemo = TXMemo
		transaction.TXStatus = "Cancelled"
	}
	if TXStatus == "Cancelled2" {
		TXMemo = "系統錯誤"
		transaction.TXMemo = TXMemo
		transaction.TXStatus = "Cancelled"
	}
	if TXStatus == "Matched" {
		TXMemo = ""
		transaction.TXMemo = TXMemo
		transaction.TXStatus = "Matched"
	}
	if TXStatus == "Finished" {
		TXMemo = ""
		transaction.TXMemo = TXMemo
		transaction.TXErrMsg = TXErrMsg
		transaction.TXStatus = "Finished"
	}

	fmt.Printf("3 updateTransactionStatus MatchedTXID = %s\n", MatchedTXID)
	transaction.MatchedTXID = MatchedTXID
	fmt.Printf("4 updateTransactionStatus transaction MatchedTXID = %s\n", transaction.MatchedTXID)

	transaction.UpdateTime = TimeNow2
	transactionAsBytes, err := json.Marshal(transaction)
	if err != nil {
		return err
	}
	err = APIstub.PutState(TXID, transactionAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateTransactionStatusbyTXID(APIstub shim.ChaincodeStubInterface, TXID string, TXStatus string) error {
	fmt.Printf("1 updateTransactionStatusbyTXID TXID = %s, TXStatus = %s\n", TXID, TXStatus)
	TimeNow2 := time.Now().Format(timelayout2)
	transaction, err := getTransactionStructFromID(APIstub, TXID)
	transaction.TXStatus = TXStatus
	var TXMemo, TXErrMsg string
	TXMemo = ""
	TXErrMsg = ""

	if TXStatus == "Finished" {
		TXMemo = ""
		transaction.TXMemo = TXMemo
		transaction.TXErrMsg = TXErrMsg
		transaction.TXStatus = "Finished"
	}

	transaction.UpdateTime = TimeNow2
	transactionAsBytes, err := json.Marshal(transaction)
	if err != nil {
		return err
	}
	err = APIstub.PutState(TXID, transactionAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateQueuedTransactionStatus(APIstub shim.ChaincodeStubInterface, TXKEY string, TXID string, TXStatus string) error {

	TimeNow2 := time.Now().Format(timelayout2)
	queuedTX, err := getQueueStructFromID(APIstub, TXKEY)
	if err != nil {
		return err
	}
	var doflg bool
	doflg = false

	for key, val := range queuedTX.TXIDs {
		if val == TXID {
			queuedTX.Transactions[key].TXStatus = TXStatus
			queuedTX.Transactions[key].UpdateTime = TimeNow2
			doflg = true
			break
		}
	}
	if doflg != true {
		return errors.New("Failed to find Queued TXID ")
	}

	queuedAsBytes, err := json.Marshal(queuedTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(TXKEY, queuedAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateHistoryTransactionStatus(APIstub shim.ChaincodeStubInterface, HTXKEY string, TXID string, TXStatus string) error {

	TimeNow2 := time.Now().Format(timelayout2)
	historyTX, err := getHistoryTransactionStructFromID(APIstub, HTXKEY)
	if err != nil {
		return err
	}
	var doflg bool
	doflg = false

	for key, val := range historyTX.TXIDs {
		if val == TXID {
			historyTX.Transactions[key].TXStatus = TXStatus
			historyTX.Transactions[key].UpdateTime = TimeNow2
			historyTX.TXStatus[key] = TXStatus
			doflg = true
			break
		}
	}
	if doflg != true {
		return errors.New("Failed to find History TXID ")
	}

	historyAsBytes, err := json.Marshal(historyTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(HTXKEY, historyAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateQueuedTransactionApproveStatus(APIstub shim.ChaincodeStubInterface, TXKEY string, TXID string, MatchedTXID string, TXStatus string) error {

	TimeNow2 := time.Now().Format(timelayout2)
	queuedTX, err := getQueueStructFromID(APIstub, TXKEY)
	if err != nil {
		return err
	}
	var doflg1, doflg2 bool
	doflg1 = false
	doflg2 = false
	var OldTXStatus, NewTXStatus, TXMemo, TXErrMsg string
	OldTXStatus = ""
	NewTXStatus = TXStatus
	TXMemo = ""
	TXErrMsg = ""

	if TXStatus == "Waiting4Payment" {
		TXMemo = "等待回應"
		NewTXStatus = "Waiting4Payment"
	}
	if TXStatus == "PaymentError" {
		TXMemo = "款不足等待補款"
		NewTXStatus = "PaymentError"
	}
	if TXStatus == "Cancelled" {
		TXMemo = "交易取消"
		NewTXStatus = "Cancelled"
	}
	if TXStatus == "Cancelled1" {
		TXMemo = "交易取消"
		NewTXStatus = "Cancelled"
	}
	if TXStatus == "Cancelled2" {
		TXMemo = "系統錯誤"
		NewTXStatus = "Cancelled"
	}
	if TXStatus == "Matched" {
		TXMemo = ""
		NewTXStatus = "Matched"
	}
	if TXStatus == "Finished" {
		TXMemo = ""
		NewTXStatus = "Finished"
	}

	for key, val := range queuedTX.TXIDs {
		OldTXStatus = queuedTX.Transactions[key].TXStatus
		fmt.Printf("1.ATXIDs: %s\n", TXID)
		fmt.Printf("2.AMatchedTXID: %s\n", MatchedTXID)
		fmt.Printf("3.OldTXStatus: %s\n", OldTXStatus)
		if TXStatus == "Finished" {
			queuedTX.Transactions[key].TXErrMsg = TXErrMsg
		}
		if val == TXID {
			fmt.Printf("3.AQkey: %d\n", key)
			fmt.Printf("4.AQval: %s\n", val)
			if OldTXStatus == "Finished" || OldTXStatus == "Cancelled" {
				doflg1 = true
				break
			}
			queuedTX.Transactions[key].TXStatus = NewTXStatus
			queuedTX.Transactions[key].TXMemo = TXMemo
			queuedTX.Transactions[key].UpdateTime = TimeNow2
			doflg1 = true
		}
		if val == MatchedTXID {
			fmt.Printf("5.AQkey: %d\n", key)
			fmt.Printf("6.AQval: %s\n", val)
			if OldTXStatus == "Finished" || OldTXStatus == "Cancelled" {
				doflg2 = true
				break
			}
			queuedTX.Transactions[key].TXStatus = NewTXStatus
			queuedTX.Transactions[key].TXMemo = TXMemo
			queuedTX.Transactions[key].UpdateTime = TimeNow2
			doflg2 = true
		}
		if doflg1 == true && doflg2 == true {
			break
		}
	}
	if doflg1 != true || doflg2 != true {
		return errors.New("Failed to find Approve-Queued TXID ")
	}

	queuedAsBytes, err := json.Marshal(queuedTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(TXKEY, queuedAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateHistoryTransactionApproveStatus(APIstub shim.ChaincodeStubInterface, HTXKEY string, TXID string, MatchedTXID string, TXStatus string) error {

	TimeNow2 := time.Now().Format(timelayout2)
	historyTX, err := getHistoryTransactionStructFromID(APIstub, HTXKEY)
	if err != nil {
		return err
	}
	var doflg1, doflg2 bool
	doflg1 = false
	doflg2 = false
	var OldTXStatus, NewTXStatus, TXMemo, TXErrMsg string
	OldTXStatus = ""
	NewTXStatus = TXStatus
	TXMemo = ""
	TXErrMsg = ""

	if TXStatus == "Waiting4Payment" {
		TXMemo = "等待回應"
		NewTXStatus = "Waiting4Payment"
	}
	if TXStatus == "PaymentError" {
		TXMemo = "款不足等待補款"
		NewTXStatus = "PaymentError"
	}
	if TXStatus == "Cancelled" {
		TXMemo = "交易取消"
		NewTXStatus = "Cancelled"
	}
	if TXStatus == "Cancelled1" {
		TXMemo = "交易取消"
		NewTXStatus = "Cancelled"
	}
	if TXStatus == "Cancelled2" {
		TXMemo = "系統錯誤"
		NewTXStatus = "Cancelled"
	}
	if TXStatus == "Matched" {
		TXMemo = ""
		NewTXStatus = "Matched"
	}
	if TXStatus == "Finished" {
		TXMemo = ""
		NewTXStatus = "Finished"
	}

	for key, val := range historyTX.TXIDs {
		OldTXStatus = historyTX.Transactions[key].TXStatus
		fmt.Printf("7.hTXIDs: %s\n", TXID)
		fmt.Printf("8.hMatchedTXID: %s\n", MatchedTXID)
		fmt.Printf("8.OldTXStatus: %s\n", OldTXStatus)
		if TXStatus == "Finished" {
			historyTX.Transactions[key].TXErrMsg = TXErrMsg
		}
		if val == TXID {
			fmt.Printf("9.AHkey: %d\n", key)
			fmt.Printf("10.AHval: %s\n", val)
			if OldTXStatus == "Finished" || OldTXStatus == "Cancelled" {
				doflg1 = true
				break
			}
			historyTX.Transactions[key].TXStatus = NewTXStatus
			historyTX.Transactions[key].TXMemo = TXMemo
			historyTX.Transactions[key].UpdateTime = TimeNow2
			historyTX.TXStatus[key] = NewTXStatus
			doflg1 = true
		}
		if val == MatchedTXID {
			fmt.Printf("11.AHkey: %d\n", key)
			fmt.Printf("12.AHval: %s\n", val)
			if OldTXStatus == "Finished" || OldTXStatus == "Cancelled" {
				doflg2 = true
				break
			}
			historyTX.Transactions[key].TXStatus = NewTXStatus
			historyTX.Transactions[key].TXMemo = TXMemo
			historyTX.Transactions[key].UpdateTime = TimeNow2
			historyTX.TXStatus[key] = NewTXStatus
			doflg2 = true
		}
		if doflg1 == true && doflg2 == true {
			break
		}
	}
	if doflg1 != true || doflg2 != true {
		return errors.New("Failed to find Approve-History TXID ")
	}

	historyAsBytes, err := json.Marshal(historyTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(HTXKEY, historyAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateTransactionTXHcode(APIstub shim.ChaincodeStubInterface, TXID string, TXHcode string) error {
	fmt.Printf("updateTransactionTXHcode: TXID=%s,TXHcode=%s\n", TXID, TXHcode)

	TimeNow2 := time.Now().Format(timelayout2)
	transaction, err := getTransactionStructFromID(APIstub, TXID)
	if err != nil {
		return err
	}
	transaction.TXHcode = TXHcode
	transaction.TXStatus = "Cancelled"
	transaction.TXMemo = "交易更正"
	transaction.UpdateTime = TimeNow2

	transactionAsBytes, err := json.Marshal(transaction)
	if err != nil {
		return err
	}
	err = APIstub.PutState(TXID, transactionAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateQueuedTransactionTXHcode(APIstub shim.ChaincodeStubInterface, TXKEY string, TXID string, TXHcode string) error {
	fmt.Printf("updateQueuedTransactionTXHcode: TXKEY=%s,TXID=%s,TXHcode=%s\n", TXKEY, TXID, TXHcode)
	TimeNow2 := time.Now().Format(timelayout2)
	queuedTX, err := getQueueStructFromID(APIstub, TXKEY)
	if err != nil {
		return err
	}
	var doflg bool
	doflg = false

	for key, val := range queuedTX.TXIDs {
		fmt.Printf("1.Qkey: %d\n", key)
		fmt.Printf("2.Qval: %s\n", val)
		if val == TXID {
			fmt.Printf("3.Qkey: %d\n", key)
			fmt.Printf("4.Qval: %s\n", val)
			queuedTX.Transactions[key].TXHcode = TXHcode
			queuedTX.Transactions[key].TXStatus = "Cancelled"
			queuedTX.Transactions[key].TXMemo = "交易更正"
			queuedTX.Transactions[key].UpdateTime = TimeNow2
			doflg = true
			break
		}
	}
	if doflg != true {
		return errors.New("5.Failed to find Queued TXID ")
	}

	queuedAsBytes, err := json.Marshal(queuedTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(TXKEY, queuedAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func updateHistoryTransactionTXHcode(APIstub shim.ChaincodeStubInterface, HTXKEY string, TXID string, TXHcode string) error {
	fmt.Printf("updateHistoryTransactionTXHcode: HTXKEY=%s,TXID=%s,TXHcode=%s\n", HTXKEY, TXID, TXHcode)
	TimeNow2 := time.Now().Format(timelayout2)
	historyTX, err := getHistoryTransactionStructFromID(APIstub, HTXKEY)
	if err != nil {
		return err
	}
	//TimeNow := time.Now().Format(timelayout)

	var doflg bool
	doflg = false

	for key, val := range historyTX.TXIDs {
		fmt.Printf("1.Hkey: %d\n", key)
		fmt.Printf("2.Hval: %s\n", val)
		if val == TXID {
			fmt.Printf("3.Hkey: %d\n", key)
			fmt.Printf("4.Hval: %s\n", val)
			historyTX.Transactions[key].TXHcode = TXHcode
			historyTX.Transactions[key].TXStatus = "Cancelled"
			historyTX.Transactions[key].TXMemo = "交易更正"
			historyTX.TXStatus[key] = "Cancelled"
			historyTX.Transactions[key].UpdateTime = TimeNow2
			doflg = true
			break
		}
	}
	if doflg != true {
		return errors.New("Failed to find History TXID ")
	}

	historyAsBytes, err := json.Marshal(historyTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(HTXKEY, historyAsBytes)
	if err != nil {
		return err
	}
	return nil
}

func (s *SmartContract) updateQueuedTransactionHcode(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	TXKEY := args[0]
	TXID := args[1]
	TXHcode := args[2]

	fmt.Printf("updateQueuedTransactionHcode: TXKEY=%s,TXID=%s,TXHcode=%s\n", TXKEY, TXID, TXHcode)
	TimeNow2 := time.Now().Format(timelayout2)
	queuedTX, err := getQueueStructFromID(APIstub, TXKEY)
	if err != nil {
		return shim.Error(err.Error())
	}
	var doflg bool
	doflg = false

	for key, val := range queuedTX.TXIDs {
		fmt.Printf("1.Qkey: %d\n", key)
		fmt.Printf("2.Qval: %s\n", val)
		if val == TXID {
			fmt.Printf("3.Qkey: %d\n", key)
			fmt.Printf("4.Qval: %s\n", val)
			queuedTX.Transactions[key].TXHcode = TXHcode
			queuedTX.Transactions[key].TXStatus = "Cancelled"
			queuedTX.Transactions[key].TXMemo = "交易更正"
			queuedTX.Transactions[key].UpdateTime = TimeNow2
			doflg = true
			break
		}
	}
	if doflg != true {
		return shim.Error("Failed to find Queued TXID ")
	}

	queuedAsBytes, err := json.Marshal(queuedTX)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutState(TXKEY, queuedAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queuedAsBytes)
}

func (s *SmartContract) updateHistoryTransactionHcode(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	HTXKEY := args[0]
	TXID := args[1]
	TXHcode := args[2]

	fmt.Printf("updateHistoryTransactionHcode: HTXKEY=%s,TXID=%s,TXHcode=%s\n", HTXKEY, TXID, TXHcode)
	TimeNow2 := time.Now().Format(timelayout2)
	historyTX, err := getHistoryTransactionStructFromID(APIstub, HTXKEY)
	if err != nil {
		return shim.Error(err.Error())
	}

	var doflg bool
	doflg = false

	for key, val := range historyTX.TXIDs {
		fmt.Printf("1.Hkey: %d\n", key)
		fmt.Printf("2.Hval: %s\n", val)
		if val == TXID {
			fmt.Printf("3.Hkey: %d\n", key)
			fmt.Printf("4.Hval: %s\n", val)
			historyTX.Transactions[key].TXHcode = TXHcode
			historyTX.Transactions[key].TXStatus = "Cancelled"
			historyTX.Transactions[key].TXMemo = "交易更正"
			historyTX.TXStatus[key] = "Cancelled"
			historyTX.Transactions[key].UpdateTime = TimeNow2
			doflg = true
			break
		}
	}
	if doflg != true {
		return shim.Error("Failed to find History TXID ")
	}

	historyAsBytes, err := json.Marshal(historyTX)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutState(HTXKEY, historyAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(historyAsBytes)
}

/*
peer chaincode invoke -n mycc -c '{"Args":["FXTradeTransfer", "B","0001","0002","2018/01/01","2018/12/31","USD/TWD","USD","1000000","TWD","1000000","30","SPOT","true"]}' -C myc 
peer chaincode query -n mycc -c '{"Args":["queryQueuedTransactionStatus","20180906","Pending","0001"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["CorrectFXTradeTransfer", "B","0001","0002","2018/09/01","2018/11/30","USD/TWD","USD","1000000","ZAR","1000000","30","SPOT","true","0001B20180921152043"]}' -C myc 
OwnCptyID不能修改，如要修改直接刪除後再新增
*/
 func (s *SmartContract) CorrectFXTradeTransfer(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)

	newTX, isPutInQueue, errMsg := validateCorrectTransaction(APIstub, args)
	if errMsg != "" {
		//return shim.Error(err.Error())
		newTX.TXErrMsg = errMsg
		newTX.TXStatus = "Cancelled"
		newTX.TXMemo = "交易被取消"
	}
	TXIndex := newTX.TXIndex
	TXID := newTX.TXID
	TXType := newTX.TXType
	OwnCptyID := newTX.OwnCptyID
	TXStatus := newTX.TXStatus
	TXHcode := newTX.TXHcode

	var doflg bool
	var TXKinds string
	doflg = false
	TXKEY := SubString(TimeNow, 0, 8) //A0710220180326
	HTXKEY := "H" + TXKEY
	TXDAY := SubString(TXID, 5, 8)
	if TXDAY < TXKEY {
		TXKEY = TXDAY
		HTXKEY = "H" + TXKEY
	}

	fmt.Printf("1.StartCorrect=%s\n")

	if isPutInQueue == true {
		newTX.isPutToQueue = true
		fmt.Printf("2.TXKEYCorrect=%s\n", TXKEY)

		queueAsBytes, err := APIstub.GetState(TXKEY)
		if err != nil {
			//return shim.Error(err.Error())
			newTX.TXErrMsg = TXKEY + ":QueueID does not exits."
			newTX.TXStatus = "Cancelled"
			newTX.TXMemo = "交易被取消"
		}
		queuedTx := QueuedTransaction{}
		json.Unmarshal(queueAsBytes, &queuedTx)
		fmt.Printf("3.HTXKEYCorrect=%s\n", HTXKEY)

		historyAsBytes, err := APIstub.GetState(HTXKEY)
		if err != nil {
			//return shim.Error(err.Error())
			newTX.TXErrMsg = TXKEY + ":HistoryID does not exits."
			newTX.TXStatus = "Cancelled"
			newTX.TXMemo = "交易被取消"
		}
		historyNewTX := TransactionHistory{}
		json.Unmarshal(historyAsBytes, &historyNewTX)

		fmt.Println("01.CTXIndex= " + TXIndex + "\n")
		fmt.Println("02.OwnCptyID= " + OwnCptyID + "\n")
		fmt.Println("03.CTXType= " + TXType + "\n")
		fmt.Println("04.CTXID= " + TXID + "\n")
		fmt.Println("05.CTXStatus= " + TXStatus + "\n")
		fmt.Println("06.Cval.TXHcode= " + TXHcode + "\n")

		if queueAsBytes == nil {
			queuedTx.ObjectType = QueuedTXObjectType
			queuedTx.TXKEY = TXKEY
			queuedTx.Transactions = append(queuedTx.Transactions, newTX)
			queuedTx.TXIndexs = append(queuedTx.TXIndexs, TXIndex)
			queuedTx.TXIDs = append(queuedTx.TXIDs, TXID)
			if historyAsBytes == nil {
				historyNewTX.ObjectType = HistoryTXObjectType
				historyNewTX.TXKEY = HTXKEY
				historyNewTX.Transactions = append(historyNewTX.Transactions, newTX)
				historyNewTX.TXIndexs = append(historyNewTX.TXIndexs, TXIndex)
				historyNewTX.TXIDs = append(historyNewTX.TXIDs, TXID)
				historyNewTX.TXStatus = append(historyNewTX.TXStatus, newTX.TXStatus)
			}
		} else if queueAsBytes != nil {
			for key, val := range queuedTx.Transactions {
				if val.TXIndex == TXIndex && val.TXStatus == TXStatus && val.OwnCptyID != OwnCptyID && val.TXType != TXType && val.TXID != TXID {
					fmt.Println("1.TXIndex= " + TXIndex + "\n")
					fmt.Println("2.OwnCptyID= " + OwnCptyID + "\n")
					fmt.Println("3.TXType= " + TXType + "\n")
					fmt.Println("4.TXID= " + TXID + "\n")
					fmt.Println("5.val.TXID= " + val.TXID + "\n")
					fmt.Println("6.TXStatus= " + TXStatus + "\n")
					fmt.Println("7.val.TXStatus= " + val.TXStatus + "\n")

					if TXStatus == "Pending" && val.TXStatus == "Pending" {
						if doflg == true {
							//return shim.Error("doflg eq to true")
							newTX.TXErrMsg = "doflg can not equle to true."
							newTX.TXStatus = "Cancelled"
							newTX.TXMemo = "交易被取消"
							break
						}
						newTX.MatchedTXID = val.TXID
						queuedTx.Transactions[key].MatchedTXID = TXID
						historyNewTX.Transactions[key].MatchedTXID = TXID
						err = updateTransactionStatus(APIstub, val.TXID, "Matched", TXID)
						if err != nil {
							//return shim.Error(err.Error())
							newTX.TXErrMsg = "Failed to execute updateTransactionStatus = Matched."
							newTX.TXStatus = "Cancelled"
							newTX.TXMemo = "交易被取消"
							break
						}
						queuedTx.Transactions[key].TXStatus = "Matched"
						historyNewTX.Transactions[key].TXStatus = "Matched"
						historyNewTX.TXStatus[key] = "Matched"
						newTX.TXStatus = "Matched"
						queuedTx.Transactions[key].TXMemo = ""
						historyNewTX.Transactions[key].TXMemo = ""
						newTX.TXMemo = ""
					
					} else {
						queuedTx.Transactions[key].TXStatus = "Finished"
						historyNewTX.Transactions[key].TXStatus = "Finished"
						historyNewTX.TXStatus[key] = "Finished"
						newTX.TXStatus = "Finished"
						queuedTx.Transactions[key].TXMemo = ""
						historyNewTX.Transactions[key].TXMemo = ""
						newTX.TXMemo = ""
						err := updateTransactionStatus(APIstub, val.TXID, "Finished", TXID)
						if err != nil {
							//return shim.Error(err.Error())
							newTX.TXErrMsg = "Failed to execute updateTransactionStatus = Finished."
							//newTX.TXStatus = "Cancelled"
							break
						}
					}
					doflg = true
					break
					
				}
			}
		}
		if queueAsBytes != nil {
			if historyAsBytes != nil {
				queuedTx.ObjectType = QueuedTXObjectType
				queuedTx.TXKEY = TXKEY
				queuedTx.Transactions = append(queuedTx.Transactions, newTX)
				queuedTx.TXIndexs = append(queuedTx.TXIndexs, TXIndex)
				queuedTx.TXIDs = append(queuedTx.TXIDs, TXID)

				historyNewTX.ObjectType = HistoryTXObjectType
				historyNewTX.TXKEY = HTXKEY
				historyNewTX.Transactions = append(historyNewTX.Transactions, newTX)
				historyNewTX.TXIndexs = append(historyNewTX.TXIndexs, TXIndex)
				historyNewTX.TXIDs = append(historyNewTX.TXIDs, TXID)
				historyNewTX.TXStatus = append(historyNewTX.TXStatus, newTX.TXStatus)
				historyNewTX.TXKinds = append(historyNewTX.TXKinds, TXKinds)
			}
		}
		QueuedAsBytes, err := json.Marshal(queuedTx)
		err = APIstub.PutState(TXKEY, QueuedAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		historyAsBytes, err = json.Marshal(historyNewTX)
		err = APIstub.PutState(HTXKEY, historyAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	TransactionAsBytes, err := json.Marshal(newTX)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutState(TXID, TransactionAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	} 

	return shim.Success(nil)

}  

func validateCorrectTransaction(
	APIstub shim.ChaincodeStubInterface,
	args []string) (FXTrade, bool, string) {

	TimeNow := time.Now().Format(timelayout)
	TimeNow2 := time.Now().Format(timelayout2)
	var err error
	var TXData, TXIndex, TXID string
	//TimeNow := time.Now().Format(timelayout)
	transaction := FXTrade{}
	transaction.ObjectType = TransactionObjectType
	transaction.TXStatus = "Cancelled"
	transaction.TXMemo = "交易更正"
	transaction.TXErrMsg = ""
	transaction.TXHcode = ""
	transaction.UpdateTime = TimeNow2
	fmt.Println("TimeNow is %s", TimeNow)

	err = checkArgArrayLength(args, 14)
	if err != nil {
		return transaction, false, "The args-length must be 13."
	}
	if len(args[0]) <= 0 {
		return transaction, false, "TXType must be a non-empty string."
	}
	if len(args[1]) <= 0 {
		return transaction, false, "OwnCptyID must be a non-empty string."
	}
	if len(args[2]) <= 0 {
		return transaction, false, "CptyID must be a non-empty string."
	}
	if len(args[3]) <= 0 {
		return transaction, false, "TradeDate must be a non-empty string."
	}
	if len(args[4]) <= 0 {
		return transaction, false, "MaturityDate must be a non-empty string."
	}
	if len(args[5]) <= 0 {
		return transaction, false, "Contract must be a non-empty string."
	}
	if len(args[6]) <= 0 {
		return transaction, false, "Curr1 must be a non-empty string."
	}
	if len(args[7]) <= 0 {
		return transaction, false, "Amount1 must be a non-empty string."
	}
	if len(args[8]) <= 0 {
		return transaction, false, "Curr2 must be a non-empty string."
	}
	if len(args[9]) <= 0 {
		return transaction, false, "Amount2 must be a non-empty string."
	}
    if len(args[10]) <= 0 {
		return transaction, false, "NetPrice must be a non-empty string."
	}
	if len(args[11]) <= 0 {
		return transaction, false, "TXKinds must be a non-empty string."
	}
	if len(args[12]) <= 0 {
		return transaction, false, "isPutToQueue flag must be a non-empty string."
	}
	if len(args[13]) <= 0 {
		return transaction, false, "TXID flag must be a non-empty string."
	}

	TXID = strings.ToUpper(args[13])
	sourceTX, err := getTransactionStructFromID(APIstub, TXID)
	if (sourceTX.TXStatus != "Pending") && (sourceTX.TXStatus != "Matched") {
		return transaction, false, "Failed to find Transaction Pending/Matched TXStatus."
	}
	//if sourceTX.TXStatus == "Cancelled" {
	//	return transaction, false, "TXStatus of transaction was Cancelled. TXHcode:" + sourceTX.TXHcode
	//}

	TXType := args[0]
	if (TXType != "B") && (TXType != "S") {
		return transaction, false, "TXType must be a B or S."
	}
	transaction.TXType = TXType
	OwnCptyID := strings.ToUpper(args[1])
	CptyID := strings.ToUpper(args[2])
	TXHcode := OwnCptyID + TXType + TimeNow
	transaction.TXID = TXHcode
	
	if OwnCptyID == CptyID {
		return transaction, false, "OwnCptyID equal to CptyID."
	}

	//if verifyIdentity(stub, OwnCptyID) != "" {
	//	return transaction, false, "OwnCptyID does not exits in the BankList."
	//}
    //if verifyIdentity(stub, CptyID) != "" {
	//	return transaction, false, "CptyID does not exits in the CptyList."
	//}
	
	transaction.OwnCptyID = OwnCptyID
	transaction.CptyID = CptyID
	TradeDate := args[3]
	transaction.TradeDate = TradeDate
	
	MaturityDate := args[4]
	transaction.MaturityDate = MaturityDate

	Contract := strings.ToUpper(args[5])
	transaction.Contract = Contract

	Curr1 := strings.ToUpper(args[6])
	transaction.Curr1 = Curr1

	Amount1, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return transaction, false, "Amount1 must be a numeric string."
	} else if Amount1 < 0 {
		return transaction, false, "Amount1 must be a positive value"
	}
	transaction.Amount1 = Amount1

	Curr2 := strings.ToUpper(args[8])
	transaction.Curr2 = Curr2

	Amount2, err := strconv.ParseFloat(args[9], 64)
	if err != nil {
		return transaction, false, "Amount2 must be a numeric string."
	} else if Amount2 < 0 {
		return transaction, false, "Amount2 must be a positive value"
	}
	transaction.Amount2 = Amount2

	NetPrice, err := strconv.ParseFloat(args[10], 64)
	if err != nil {
		return transaction, false, "NetPrice must be a numeric string."
	} else if NetPrice < 0 {
		return transaction, false, "NetPrice must be a positive value"
	}
	transaction.NetPrice = NetPrice

	//TXData = OwnCptyID + CptyID + TradeDate + MaturityDate + Contract + Curr1 + Amount1 
	if TXType == "S" {
		TXData = OwnCptyID + CptyID + TradeDate + MaturityDate + Contract + Curr1 + strconv.FormatFloat(Amount1, 'f', 4, 64)
		TXIndex = getSHA256(TXData)
	}
	fmt.Println("6.validateTransaction= " + TimeNow + "\n")
	if TXType == "B" {
		TXData = CptyID + OwnCptyID + TradeDate + MaturityDate + Contract + Curr1 + strconv.FormatFloat(Amount1, 'f', 4, 64)
		TXIndex = getSHA256(TXData)
	}

	transaction.TXIndex = TXIndex
	transaction.TXHcode = TXID
	transaction.TXStatus = "Pending"

	err2 := updateTransactionTXHcode(APIstub, TXID, TXHcode)
	if err2 != nil {
		//return transaction, false, err2
		transaction.TXMemo = "更正失敗"
		transaction.TXErrMsg = TXID + ":updateTransactionTXHcode execution failed."
		return transaction, false, TXID + ":updateTransactionTXHcode execution failed."
	}

	return transaction, true, ""

} 

/* func updateEndDayTransactionStatus(APIstub shim.ChaincodeStubInterface, TXID string) (string, error) {
	var MatchedTXID string
	MatchedTXID = ""
	TimeNow2 := time.Now().Format(timelayout2)
	transaction, err := getTransactionStructFromID(APIstub, TXID)
	if transaction.TXStatus != "Pending" && transaction.TXStatus != "Waiting4Payment" && transaction.TXStatus != "PaymentError" {
		return MatchedTXID, errors.New("Failed to find Transaction Pending OR Waiting4Payment TXStatus")
	}
	TXStatus := transaction.TXStatus
	TXMemo := ""
	if TXStatus == "Waiting4Payment" {
		TXMemo = "日終交易取消"
	}
	if TXStatus == "PaymentError" {
		TXMemo = "款不足"
	}
	if TXStatus == "Pending" {
		TXMemo = "尚未比對"
	}

	transaction.TXStatus = "Cancelled"
	transaction.TXMemo = TXMemo
	transaction.UpdateTime = TimeNow2
	transactionAsBytes, err := json.Marshal(transaction)
	if err != nil {
		return MatchedTXID, err
	}
	err = APIstub.PutState(TXID, transactionAsBytes)
	if err != nil {
		return MatchedTXID, err
	}
	if (TXStatus == "Waiting4Payment") || (TXStatus == "PaymentError") {
		MatchedTXID = transaction.MatchedTXID
		transaction2, _ := getTransactionStructFromID(stub, MatchedTXID)
		if transaction2 != nil {
			transaction2.TXStatus = "Cancelled"
			transaction2.TXMemo = TXMemo
			transaction2.UpdateTime = TimeNow2
			transaction2AsBytes, err := json.Marshal(transaction2)
			if err != nil {
				return MatchedTXID, err
			}
			err = APIstub.PutState(MatchedTXID, transaction2AsBytes)
			if err != nil {
				return MatchedTXID, err
			}
		}

		TXType := transaction.TXType
		SecurityID := transaction.SecurityID
		SecurityAmount := transaction.SecurityAmount
		Payment := transaction.Payment
		TXFrom := transaction.TXFrom
		TXTo := transaction.TXTo
		BankFrom := transaction.BankFrom
		BankTo := transaction.BankTo

		if TXType == "S" {
			senderBalance, receiverBalance, senderPendingBalance, receiverPendingBalance, err := resetAccountBalance(stub, SecurityID, SecurityAmount, Payment, TXFrom, TXTo)
			senderBalance, receiverBalance, err = resetSecurityAmount(stub, SecurityID, Payment, SecurityAmount, TXFrom, TXTo)
			if BankFrom != BankTo {
				err = updateBankTotals(stub, TXFrom, SecurityID, TXFrom, Payment, SecurityAmount, false)
				if err != nil {
					return MatchedTXID, err
				}
				err = updateBankTotals(stub, TXTo, SecurityID, TXTo, Payment, SecurityAmount, true)
				if err != nil {
					return MatchedTXID, err
				}
			}
			if (senderBalance < 0) || (receiverBalance < 0) || (senderPendingBalance < 0) || (receiverPendingBalance < 0) {
				return MatchedTXID, errors.New("senderBalance,receiverBalance,senderPendingBalance,receiverPendingBalance <0")
			}

		}
		if TXType == "B" {
			senderBalance, receiverBalance, senderPendingBalance, receiverPendingBalance, err := resetAccountBalance(stub, SecurityID, SecurityAmount, Payment, TXTo, TXFrom)
			senderBalance, receiverBalance, err = resetSecurityAmount(stub, SecurityID, Payment, SecurityAmount, TXTo, TXFrom)
			if BankFrom != BankTo {
				err = updateBankTotals(stub, TXFrom, SecurityID, TXFrom, Payment, SecurityAmount, true)
				if err != nil {
					return MatchedTXID, err
				}
				err = updateBankTotals(stub, TXTo, SecurityID, TXTo, Payment, SecurityAmount, false)
				if err != nil {
					return MatchedTXID, err
				}
			}
			if (senderBalance < 0) || (receiverBalance < 0) || (senderPendingBalance < 0) || (receiverPendingBalance < 0) {
				return MatchedTXID, errors.New("senderBalance,receiverBalance,senderPendingBalance,receiverPendingBalance <0")
			}

		}
	}

	return MatchedTXID, nil
} */

/* func updateEndDayQueuedTransactionStatus(APIstub shim.ChaincodeStubInterface, TXKEY string, TXID string, MatchedTXID string) error {

	TimeNow2 := time.Now().Format(timelayout2)
	queuedTX, err := getQueueStructFromID(APIstub, TXKEY)
	if err != nil {
		return err
	}

	var doflg bool
	doflg = false

	for key, val := range queuedTX.TXIDs {
		fmt.Printf("qkey1: %d\n", key)
		fmt.Printf("qval1: %s\n", val)

		TXStatus := queuedTX.Transactions[key].TXStatus
		TXMemo := ""
		if TXStatus == "Waiting4Payment" {
			TXMemo = "日終交易取消"
		}
		if TXStatus == "PaymentError" {
			TXMemo = "款不足"
		}
		if TXStatus == "Pending" {
			TXMemo = "尚未比對"
		}

		if val == TXID {
			fmt.Printf("qkey2: %d\n", key)
			fmt.Printf("qval2: %s\n", val)

			if (queuedTX.Transactions[key].TXStatus == "Pending") || (queuedTX.Transactions[key].TXStatus == "Waiting4Payment") || (queuedTX.Transactions[key].TXStatus == "PaymentError") {
				queuedTX.Transactions[key].TXStatus = "Cancelled"
				queuedTX.Transactions[key].TXMemo = TXMemo
				queuedTX.Transactions[key].UpdateTime = TimeNow2
				doflg = true
			}
		}
		if val == MatchedTXID {
			fmt.Printf("qkey3: %d\n", key)
			fmt.Printf("qval3: %s\n", val)

			if (queuedTX.Transactions[key].TXStatus == "Pending") || (queuedTX.Transactions[key].TXStatus == "Waiting4Payment") || (queuedTX.Transactions[key].TXStatus == "PaymentError") {
				queuedTX.Transactions[key].TXStatus = "Cancelled"
				queuedTX.Transactions[key].TXMemo = TXMemo
				queuedTX.Transactions[key].UpdateTime = TimeNow2
				doflg = true
			}
		}
	}
	if doflg != true {
		return errors.New("Failed to find Queued Pending OR Waiting4Payment OR PaymentError TXStatus ")
	}

	queuedAsBytes, err := json.Marshal(queuedTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(TXKEY, queuedAsBytes)
	if err != nil {
		return err
	}
	return nil
} */

func updateEndDayHistoryTransactionStatus(APIstub shim.ChaincodeStubInterface, HTXKEY string, TXID string, MatchedTXID string) error {

	TimeNow2 := time.Now().Format(timelayout2)
	historyTX, err := getHistoryTransactionStructFromID(APIstub, HTXKEY)
	if err != nil {
		return err
	}

	var doflg bool
	doflg = false

	for key, val := range historyTX.TXIDs {
		fmt.Printf("hkey1: %d\n", key)
		fmt.Printf("hval1: %s\n", val)

		TXStatus := historyTX.Transactions[key].TXStatus
		TXMemo := ""
		if TXStatus == "Waiting4Payment" {
			TXMemo = "日終交易取消"
		}
		if TXStatus == "PaymentError" {
			TXMemo = "款不足"
		}
		if TXStatus == "Pending" {
			TXMemo = "尚未比對"
		}

		if val == TXID {
			fmt.Printf("hkey2: %d\n", key)
			fmt.Printf("hval2: %s\n", val)
			if (historyTX.Transactions[key].TXStatus == "Pending") || (historyTX.Transactions[key].TXStatus == "Waiting4Payment") || (historyTX.Transactions[key].TXStatus == "PaymentError") {
				historyTX.Transactions[key].TXStatus = "Cancelled"
				historyTX.Transactions[key].TXMemo = TXMemo
				historyTX.Transactions[key].UpdateTime = TimeNow2
				historyTX.TXStatus[key] = "Cancelled"
				doflg = true
			}
		}
		if val == MatchedTXID {
			fmt.Printf("hkey3: %d\n", key)
			fmt.Printf("hval3: %s\n", val)
			if (historyTX.Transactions[key].TXStatus == "Pending") || (historyTX.Transactions[key].TXStatus == "Waiting4Payment") || (historyTX.Transactions[key].TXStatus == "PaymentError") {
				historyTX.Transactions[key].TXStatus = "Cancelled"
				historyTX.Transactions[key].TXMemo = TXMemo
				historyTX.Transactions[key].UpdateTime = TimeNow2
				historyTX.TXStatus[key] = "Cancelled"
				doflg = true
			}
		}
	}
	if doflg != true {
		return errors.New("Failed to find History Pending OR Waiting4Payment TXStatus ")
	}

	historyAsBytes, err := json.Marshal(historyTX)
	if err != nil {
		return err
	}
	err = APIstub.PutState(HTXKEY, historyAsBytes)
	if err != nil {
		return err
	}
	return nil
}

//peer chaincode query -n mycc -c '{"Args":["queryTXIDTransactions", "0001B20181009112451"]}' -C myc
//peer chaincode query -n mycc -c '{"Args":["queryTXIDTransactions", "0002S20180903060425"]}' -C myc
//用TXID去查詢FXTrade，回傳一筆   
func (s *SmartContract) queryTXIDTransactions(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	NewTXAsBytes, _ := APIstub.GetState(args[0])
	NewTX := FXTrade{}
	json.Unmarshal(NewTXAsBytes, &NewTX)

	NewTXAsBytes, err := json.Marshal(NewTX)
	if err != nil {
		return shim.Error("Failed to query NewTX state")
	}

	return shim.Success(NewTXAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["queryTXKEYTransactions", "2019"]}' -C myc
//用TXKEY去查詢QueuedTransaction，回傳第一筆
func (s *SmartContract) queryTXKEYTransactions(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	QueuedTXAsBytes, _ := APIstub.GetState(args[0])
	QueuedTX := QueuedTransaction{}
	json.Unmarshal(QueuedTXAsBytes, &QueuedTX)

	QueuedTXAsBytes, err := json.Marshal(QueuedTX.Transactions)
	if err != nil {
		return shim.Error("Failed to query QueuedTX state")
	}

	return shim.Success(QueuedTXAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["queryHistoryTXKEYTransactions", "H20180829"]}' -C myc
//用TXKEY去查詢TransactionHistory，回傳第一筆
func (s *SmartContract) queryHistoryTXKEYTransactions(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	HistoryNewTXAsBytes, _ := APIstub.GetState(args[0])
	HistoryNewTX := TransactionHistory{}
	json.Unmarshal(HistoryNewTXAsBytes, &HistoryNewTX)

	HistoryNewTXAsBytes, err := json.Marshal(HistoryNewTX.Transactions)
	if err != nil {
		return shim.Error("Failed to query HistoryNewTX state")
	}

	return shim.Success(HistoryNewTXAsBytes)
}

//peer chaincode query -n mycc -c '{"Args":["getHistoryForTransaction", "0001B20180905161403"]}' -C myc
//peer chaincode query -n mycc -c '{"Args":["getHistoryForTransaction", "0001B20180903060152"]}' -C myc

//用TXKEY去查詢TransactionHistory，回傳多筆
func (s *SmartContract) getHistoryForTransaction(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	TXID := args[0]

	fmt.Printf("- start getHistoryForTransaction: %s\n", TXID)

	resultsIterator, err := APIstub.GetHistoryForKey(TXID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForTransaction returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["getHistoryTXIDForTransaction","0002S20180903060425","7de652b09b8ee1eea3b0c7b17c0b515877014b37a227ee2fa76acc81b90a0bf3"]}' -C myc
func (s *SmartContract) getHistoryTXIDForTransaction(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	TransactionID := args[0]
	TXID := args[1]

	fmt.Printf("- start getHistoryTXIDForTransaction: %s\n", TransactionID)
	fmt.Printf("- start getHistoryTXIDForTransaction: %s\n", TXID)

	resultsIterator, err := APIstub.GetHistoryForKey(TransactionID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if response.TxId == TXID {
			buffer.WriteString("{\"TxId\":")
			buffer.WriteString("\"")
			buffer.WriteString(response.TxId)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Value\":")
			// if it was a delete operation on given key, then we need to set the
			//corresponding value null. Else, we will write the response.Value
			//as-is (as the Value itself a JSON marble)
			if response.IsDelete {
				buffer.WriteString("null")
			} else {
				buffer.WriteString(string(response.Value))
			}

			buffer.WriteString(", \"Timestamp\":")
			buffer.WriteString("\"")
			buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
			buffer.WriteString("\"")

			buffer.WriteString(", \"IsDelete\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatBool(response.IsDelete))
			buffer.WriteString("\"")
			buffer.WriteString("}")

			break
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryTXIDForTransaction returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

//**peer chaincode query -n mycc -c '{"Args":["getHistoryForQueuedTransaction", "H20180903"]}' -C myc

func (s *SmartContract) getHistoryForQueuedTransaction(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	TXKEY := args[0]

	fmt.Printf("- start getHistoryForQueuedTransaction: %s\n", TXKEY)

	resultsIterator, err := APIstub.GetHistoryForKey(TXKEY)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForQueuedTransaction returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["getHistoryTXIDForQueuedTransaction","20180610","a4723f60d5c85d29a2107382fb8e3c8c1624924b970efa04f313727a0dfaa0ff"]}' -C myc
func (s *SmartContract) getHistoryTXIDForQueuedTransaction(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	QueuedTransactionID := args[0]
	TXID := args[1]

	fmt.Printf("- start getHistoryTXIDForQueuedTransaction: %s\n", QueuedTransactionID)
	fmt.Printf("- start getHistoryTXIDForQueuedTransaction: %s\n", TXID)

	resultsIterator, err := APIstub.GetHistoryForKey(QueuedTransactionID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if response.TxId == TXID {
			buffer.WriteString("{\"TxId\":")
			buffer.WriteString("\"")
			buffer.WriteString(response.TxId)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Value\":")
			// if it was a delete operation on given key, then we need to set the
			//corresponding value null. Else, we will write the response.Value
			//as-is (as the Value itself a JSON marble)
			if response.IsDelete {
				buffer.WriteString("null")
			} else {
				buffer.WriteString(string(response.Value))
			}

			buffer.WriteString(", \"Timestamp\":")
			buffer.WriteString("\"")
			buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
			buffer.WriteString("\"")

			buffer.WriteString(", \"IsDelete\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatBool(response.IsDelete))
			buffer.WriteString("\"")
			buffer.WriteString("}")

			break
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryTXIDForQueuedTransaction returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryAllTransactions", "CPTYAB20180828133108","CPTYAB20180828143108"]}' -C myc

func (s *SmartContract) queryAllTransactions(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	//TXID = OwnCptyID + TXType + TimeNow
	//0001B20180828133108
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
		// Record is a JSON object, so we write as/is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryTables","{\"selector\":{\"docType\":\"Transaction\",\"TXHcode\":\"0001B20180906123212\"}}"]}' -C myc
//peer chaincode query -n mycc -c '{"Args":["queryTables","{\"selector\":{\"docType\":\"Transaction\",\"MaturityDate\":\"2018/12/31\"}}"]}' -C myc
func (s *SmartContract) queryTables(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(APIstub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func getQueryResultForQueryString(APIstub shim.ChaincodeStubInterface, queryString string)([] byte, error) {
	
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
    resultsIterator, err := APIstub.GetQueryResult(queryString)
    defer resultsIterator.Close()
    if err != nil {
		fmt.Printf("- getQueryResultForQueryString resultsIterator error")
        return nil, err
    }
    // buffer is a JSON array containing QueryRecords
    var buffer bytes.Buffer
	buffer.WriteString("[")
	fmt.Printf("- getQueryResultForQueryString start buffer")
    bArrayMemberAlreadyWritten := false
    for resultsIterator.HasNext() {
        queryResponse,
        err := resultsIterator.Next()
        if err != nil {
            return nil, err
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
    fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
    return buffer.Bytes(), nil
}

//peer chaincode query -n mycc -c '{"Args":["queryAllQueuedTransactions", "20180918","20180919"]}' -C myc
func (s *SmartContract) queryAllQueuedTransactions(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	//TXKEY = SubString(TimeNow,0,8)
	//20180406
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
		// Record is a JSON object, so we write as/is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryAllHistoryTransactions", "H20180918","H20180919"]}' -C myc 
func (s *SmartContract) queryAllHistoryTransactions(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	//TXKEY = SubString(TimeNow,0,8)
	//20180406
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
		// Record is a JSON object, so we write as/is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryAllTransactionKeys", "0001" , "9999"]}' -C myc
func (s *SmartContract) queryAllTransactionKeys(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 2 {
		return shim.Error("Keys operation must include two arguments, startKey and endKey")
	}
	startKey := args[0]
	endKey := args[1]

	//sleep needed to test peer's timeout behavior when using iterators
	stime := 0
	if len(args) > 2 {
		stime, _ = strconv.Atoi(args[2])
	}

	keysIter, err := APIstub.GetStateByRange(startKey, endKey)
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

}

//peer chaincode query -n mycc -c '{"Args":["queryQueuedTransactionStatus","20180928","Pending","0001"]}' -C myc
func (s *SmartContract) queryQueuedTransactionStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	TXKEY := args[0]
	TXStatus := args[1]
	CptyID := args[2]

	QueuedAsBytes, _ := APIstub.GetState(TXKEY)
	QueuedTX := QueuedTransaction{}
	json.Unmarshal(QueuedAsBytes, &QueuedTX)

	var doflg bool
	doflg = false
	var buffer bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString("{\"TXKEY\":")
	buffer.WriteString("\"")
	buffer.WriteString(QueuedTX.TXKEY)
	buffer.WriteString("\"")
	buffer.WriteString(",\"Transactions\":[")
	bArrayMemberAlreadyWritten := false
	for key, val := range QueuedTX.Transactions {

		if (val.TXStatus == TXStatus || TXStatus == "All") && (val.OwnCptyID == CptyID || val.CptyID == CptyID || CptyID == "All") {
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"QueuedKey\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.Itoa(key + 1))
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXID\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXType\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXType)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXKinds\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXKinds)
			buffer.WriteString("\"")
			buffer.WriteString(", \"OwnCptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].OwnCptyID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"CptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].CptyID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TradeDate\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TradeDate)
			buffer.WriteString("\"")
			buffer.WriteString(", \"MaturityDate\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].MaturityDate)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Contract\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].Contract)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Curr1\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].Curr1)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Amount1\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(QueuedTX.Transactions[key].Amount1,'f', 4, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"Curr2\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].Curr2)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Amount2\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(QueuedTX.Transactions[key].Amount2,'f', 4, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"NetPrice\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(QueuedTX.Transactions[key].NetPrice,'f', 4, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXStatus\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXStatus)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXMemo\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXMemo)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXErrMsg\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXErrMsg)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXHcode\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXHcode)
			buffer.WriteString("\"")
			buffer.WriteString(", \"MatchedTXID\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].MatchedTXID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"CreateTime\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].CreateTime)
			buffer.WriteString("\"")
			buffer.WriteString(", \"UpdateTime\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].UpdateTime)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXIndex\":")
			buffer.WriteString("\"")
			buffer.WriteString(QueuedTX.Transactions[key].TXIndex)
			buffer.WriteString("\"")
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
			doflg = true
		}
	}
	buffer.WriteString("]")
	if doflg != true {
		//return shim.Error("Failed to find QueuedTransaction ")
		buffer.WriteString(", \"Value\":")
		buffer.WriteString("Failed to find QueuedTransaction")
	}
	buffer.WriteString("}]")
	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryHistoryTransactionStatus","H20180609","Finished"]}' -C myc
func (s *SmartContract) queryHistoryTransactionStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	HTXKEY := args[0]
	TXStatus := args[1]
	CptyID := args[2]

	HistoryAsBytes, _ := APIstub.GetState(HTXKEY)
	HistoryTX := TransactionHistory{}
	json.Unmarshal(HistoryAsBytes, &HistoryTX)

	var doflg bool
	doflg = false
	var buffer bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString("{\"HTXKEY\":")
	buffer.WriteString("\"")
	buffer.WriteString(HistoryTX.TXKEY)
	buffer.WriteString("\"")
	buffer.WriteString(",\"Transactions\":[")
	bArrayMemberAlreadyWritten := false
	for key, val := range HistoryTX.Transactions {
		if (val.TXStatus == TXStatus || TXStatus == "All") && (val.OwnCptyID == CptyID || val.CptyID == CptyID || CptyID == "All") {
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"QueuedKey\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.Itoa(key + 1))
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXID\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TXID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXType\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TXType)
			buffer.WriteString("\"")
			buffer.WriteString(", \"OwnCptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].OwnCptyID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"CptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].CptyID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TradeDate\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TradeDate)
			buffer.WriteString("\"")
			buffer.WriteString(", \"MaturityDate\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].MaturityDate)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Contract\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].Contract)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Curr1\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].Curr1)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Amount1\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(HistoryTX.Transactions[key].Amount1,'f', 4, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"Curr2\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].Curr2)
			buffer.WriteString("\"")
			buffer.WriteString(", \"Amount2\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(HistoryTX.Transactions[key].Amount2,'f', 4, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"NetPrice\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(HistoryTX.Transactions[key].NetPrice,'f', 4, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXStatus\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TXStatus)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXMemo\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TXMemo)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXErrMsg\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TXErrMsg)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXHcode\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TXHcode)
			buffer.WriteString("\"")
			buffer.WriteString(", \"MatchedTXID\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].MatchedTXID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"CreateTime\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].CreateTime)
			buffer.WriteString("\"")
			buffer.WriteString(", \"UpdateTime\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].UpdateTime)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXIndex\":")
			buffer.WriteString("\"")
			buffer.WriteString(HistoryTX.Transactions[key].TXIndex)
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
		buffer.WriteString("Failed to find HistoryTransaction")
	}
	buffer.WriteString("}]")
	fmt.Printf("%s", buffer.String())

	return shim.Success(buffer.Bytes())
}

//peer chaincode query -n mycc -c '{"Args":["queryMTMTransactionStatus","MTM20180928","0001"]}' -C myc
func (s *SmartContract) queryMTMTransactionStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	MTMTXKEY := args[0]
	CptyID := args[1]

	MTMAsBytes, _ := APIstub.GetState(MTMTXKEY)
	MTMTX := TransactionMTM{}
	json.Unmarshal(MTMAsBytes, &MTMTX)

	var doflg bool
	doflg = false
	var buffer bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString("{\"MTMTXKEY\":")
	buffer.WriteString("\"")
	buffer.WriteString(MTMTX.TXKEY)
	buffer.WriteString("\"")
	buffer.WriteString(",\"Transactions\":[")
	bArrayMemberAlreadyWritten := false
	for key, val := range MTMTX.TransactionsMTM {
		if (val.OwnCptyID == CptyID || val.CptyID == CptyID || CptyID == "All") {
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"MTMKey\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.Itoa(key + 1))
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXID\":")
			buffer.WriteString("\"")
			buffer.WriteString(MTMTX.TransactionsMTM[key].TXID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"FXTXID\":")
			buffer.WriteString("\"")
			buffer.WriteString(MTMTX.TransactionsMTM[key].FXTXID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"TXKinds\":")
			buffer.WriteString("\"")
			buffer.WriteString(MTMTX.TransactionsMTM[key].TXKinds)
			buffer.WriteString("\"")
			buffer.WriteString(", \"OwnCptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(MTMTX.TransactionsMTM[key].OwnCptyID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"CptyID\":")
			buffer.WriteString("\"")
			buffer.WriteString(MTMTX.TransactionsMTM[key].CptyID)
			buffer.WriteString("\"")
			buffer.WriteString(", \"NetPrice\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(MTMTX.TransactionsMTM[key].NetPrice,'f', 6, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"ClosePrice\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(MTMTX.TransactionsMTM[key].ClosePrice,'f', 6, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"MTM\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatFloat(MTMTX.TransactionsMTM[key].MTM,'f', 4, 64))
			buffer.WriteString("\"")
			buffer.WriteString(", \"CreateTime\":")
			buffer.WriteString("\"")
			buffer.WriteString(MTMTX.TransactionsMTM[key].CreateTime)
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

/*
peer chaincode query -n mycc -c '{"Args":["queryMTMTransactionStatus","MTM20180928",""]}' -C myc
peer chaincode query -n mycc -c '{"Args":["queryMTMTransactionStatus","","0001"]}' -C myc
peer chaincode query -n mycc -c '{"Args":["queryMTMTransactionStatus","MTM20180928","0002"]}' -C myc

func (s *SmartContract) queryMTMTransactionStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	var queryString,queryString2 string

	MTMKEY := args[0]
	CptyID := args[1]
	
	//Query MTMDate 
	if CptyID == "" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"MTM\",\"TXKEY\":\"%s\"}}", MTMKEY)
	//Query Cpty	
	} else if MTMKEY == "" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"MTM\",\"OwnCptyID\":\"%s\"}}", CptyID)
		queryString2 = fmt.Sprintf("{\"selector\":{\"docType\":\"MTM\",\"CptyID\":\"%s\"}}", CptyID)
	}  else {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"MTM\",\"TXKEY\":\"%s\",\"OwnCptyID\":\"%s\"}}", MTMKEY,CptyID)	
		queryString2 = fmt.Sprintf("{\"selector\":{\"docType\":\"MTM\",\"TXKEY\":\"%s\",\"CptyID\":\"%s\"}}", MTMKEY,CptyID)
	}	
	fmt.Printf("queryMTMTransactionStatus.queryString:\n%s\n", queryString)
	fmt.Printf("queryMTMTransactionStatus.queryString2:\n%s\n", queryString2)

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
	
	if CptyID != "" {
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
        	buffer.WriteString(", \"Record\":")
         
        	buffer.WriteString(string(queryResponse2.Value))
        	buffer.WriteString("}")
        	bArrayMemberAlreadyWritten2 = true
		}
    }
    buffer.WriteString("]")
 
    return shim.Success(buffer.Bytes())
}
*/

func checkArgArrayLength(
	args []string,
	expectedArgLength int) error {

	argArrayLength := len(args)
	if argArrayLength != expectedArgLength {
		errMsg := fmt.Sprintf(
			"Incorrect number of arguments: Received %d, expecting %d",
			argArrayLength,
			expectedArgLength)
		return errors.New(errMsg)
	}
	return nil
}

//peer chaincode query -n mycc -c '{"Args":["fetchEURUSDviaOraclize"]}' -C myc
/*
func (s *SmartContract) fetchEURUSDviaOraclize(APIstub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("============= START : Calling the oraclize chaincode =============")
	var datasource = "URL"                                                                  // Setting the Oraclize datasource
	var query = "json(https://min-api.cryptocompare.com/data/price?fsym=EUR&tsyms=USD).USD" // Setting the query
	//var query ="await request.get('https://min-api.cryptocompare.com/data/price?fsym=EUR&tsyms=USD')"
	fmt.Printf("proof: %s", query)
	result, proof := oraclizeapi.OraclizeQuery_sync(APIstub, datasource, query, oraclizeapi.TLSNOTARY)
	fmt.Printf("proof: %s", proof)
	fmt.Printf("\nresult: %s\n", result)
	fmt.Println("Do something with the result...")
	//var request = require('request')
    
	fmt.Println("============= END : Calling the oraclize chaincode =============")
	return shim.Success(nil)
} 
*/

//peer chaincode query -n mycc -c '{"Args":["getrate","192.168.50.89","20181025"]}' -C myc
func (s *SmartContract) getrate(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// http://127.0.0.2:8080/?datadate=20181025&curr1=USD&curr2=TWD
	queryString := "http://" + args[0] + ":8080/?datadate=" + args[1] + "&curr1=USD&curr2=TWD" 
	fmt.Println("getrate.queryString= " + queryString + "\n")
	resp, err := http.Post(queryString,
						   "application/x-www-form-urlencoded",
						   strings.NewReader(""))
    if err != nil {
        fmt.Println(err)
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
	}
	fmt.Println("getrate= " + string(body) + "\n")
    return shim.Success(nil)
} 
