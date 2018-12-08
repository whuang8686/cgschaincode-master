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
	"math"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"

)

const CollateralTXObjectType string = "Collateral"


type TransactionCollateral struct {
	ObjectType           string        `json:"docType"`             // default set to "Collateral"
	TXID                 string        `json:"TXID"`                // 交易序號資料 
	TXDATE               string        `json:"TXDATE"`              // 交易日期：TXDATE(YYYYMMDD) 
	OwnCptyID            string        `json:"OwnCptyID"`
	CptyID               string        `json:"CptyID"`              // 交易對手
	MTM                  float64       `json:"MTM"`       	        // (5)
	OurThreshold         int64         `json:"OwnThreshold"`        // 本行門鑑金額 (4)
	CreditGuaranteeAmt   int64         `json:"CreditGuaranteeAmt"`  // 信用擔保金額 (6)=(5)-(4)
	CreditGuaranteeBal   int64         `json:"CreditGuaranteeBal"`  // 信用擔保餘額 (7)
	TXKinds              string        `json:"TXKinds"`             // 返還/交付
	Collateral           int64         `json:"Collateral"`          // Collateral (8)=(6)-(7)
	CptyMTA              int64         `json:"CptyMTA"`             // 交易對手最低轉讓金額
	MarginCall           int64         `json:"MarginCall"`          // MarginCall
	TXStatus             string        `json:"TXStatus"`            // Pending, Finished, Suspended
	CreateTime           string        `json:"createTime"`          // 建立時間
}

type CollateralDetail struct {
	ObjectType           string        `json:"docType"`             // default set to "CollateralDetail"
	TXID                 string        `json:"TXID"`                // 交易序號資料 
	TXDATE               string        `json:"TXDATE"`              // 交易日期：TXDATE(YYYMMDD) 
	OwnCptyID            string        `json:"OwnCptyID"`
	CptyID               string        `json:"CptyID"`              // 交易對手
	Curr                 string        `json:"Curr"`                // 幣別
	CollateralType       string        `json:"CollateralType"`      // 擔保品種類 Bond,Cash
	CollateralDetail     string        `json:"CollateralDetail"`    // Bond放債券代碼,Cash放幣別
	Amount               float64       `json:"Amount"`       	    // 擔保品金額
	Discount             float64       `json:"Discount"`       	    // 折扣率
	DiscountAmount       float64       `json:"DiscountAmount"`      // 折扣後金額
	FXTXID               string        `json:"FXTXID"`              // 交易序號資料 TransactionCollateral
	CreateTime           string        `json:"createTime"`          // 建立時間
}

type CashFlow struct {
	ObjectType           string        `json:"docType"`             // default set to "CashFlow"
	TXID                 string        `json:"TXID"`                // 交易序號資料 
	TXDATE               string        `json:"TXDATE"`              // 交易日期：TXDATE(YYYMMDD) 
	OwnCptyID            string        `json:"OwnCptyID"`
	CptyID               string        `json:"CptyID"`              // 交易對手
	CashFlowType         string        `json:"CashFlowType"`        // CashFlow種類：FXTrade,Collateral
	CashFlowDetail       string        `json:"CashFlowDetail`       // USDBond,TWDBond,AUD
	CashFlowAmount       float64       `json:"CashFlowAmount"`      // CashFlow金額
	FXTXID               string        `json:"FXTXID"`              // 交易序號資料 
	CreateTime           string        `json:"createTime"`          // 建立時間
}

/*
peer chaincode invoke -n mycc -c '{"Args":["FXTradeCollateral", "20181026","0001"]}' -C myc 
peer chaincode query -n mycc -c '{"Args":["queryTables","{\"selector\":{\"docType\":\"MTMTX\",\"TXKEY\":\"MTM20180928\"}}"]}' -C myc
*/
func (s *SmartContract) FXTradeCollateral(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {
	
	//TimeNow := time.Now().Format(timelayout)

	//先前除當日資料
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	TXDATE := args[0]
	OwnCptyID := args[1]

	var recint int64= 0
	var recint1 int64= 0
	var i int64= 0
	var MarginCall float64=0
	var TXKinds string

    //查詢本行門鑑金額
	queryString1 := fmt.Sprintf("{\"selector\": {\"docType\":\"CptyISDA\",\"OwnCptyID\":\"%s\"}}", OwnCptyID)
	fmt.Println("queryString1= " + queryString1 + "\n") 
	ownthreshold := [10]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	cptymta := [10]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	rounding := [10]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
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
			
		ownthreshold[CptyID-1] = transactionArr1[recint1].OwnThreshold
		cptymta[CptyID-1] = transactionArr1[recint1].CptyMTA
		rounding[CptyID-1] += transactionArr1[recint1].Rounding

		recint1++
	}
	fmt.Println("transactionArr[recint].ok= \n")

    //取得MTM合計
	queryString := fmt.Sprintf("{\"selector\": {\"docType\":\"MTM\",\"TXKEY\":\"%s\"}}", "MTM" + TXDATE)
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

		for key := range transaction.TransactionsMTM {
			fmt.Println("transactionArr[recint].val.OwnCptyID= " + transaction.TransactionsMTM[key].OwnCptyID  + "\n")
			fmt.Println("transactionArr[recint].val.CptyID= " + strings.Replace(transaction.TransactionsMTM[key].CptyID,"0","",-1) + "\n")		
			fmt.Println("transactionArr[recint].val.MTM= " + strconv.FormatFloat(transaction.TransactionsMTM[key].MTM ,'f', 4, 64) + "\n")
			CptyID, err := strconv.ParseInt(strings.Replace(transaction.TransactionsMTM[key].CptyID,"0","",-1) ,10, 64)
   			if err != nil {
				return shim.Error("Failed to strconv.Atoi")
   			}
			fmt.Println("transactionArr[recint].val.CptyID= " + strconv.FormatInt(CptyID-1,16) + "\n")
			if transaction.TransactionsMTM[key].OwnCptyID == OwnCptyID {
				summtm[CptyID-1] += transaction.TransactionsMTM[key].MTM 
			}
		}	
        recint++
	}	
	for i = 0; i < 10 ; i++ {
		fmt.Println("array.ownthreshold= " + strconv.FormatInt(ownthreshold[i] ,10) + "\n")
		fmt.Println("array.summtm= " + strconv.FormatFloat(summtm[i] ,'f', 4, 64) + "\n")

		//queryArgs := [][]byte{[]byte("CreateFXTradeCollateral"), []byte("20181012"), []byte("0001"), []byte("0002")}
		//peer chaincode query -n mycc -c '{"Args":["queryMTMPrice","20181012"]}' -C myc      
		if summtm[i] != 0  {
			//response := APIstub.InvokeChaincode("mycc", queryArgs, "myc")
			//response := s.CreateFXTradeCollateral(APIstub, []string{"20181010","0001","0002"})
			//if response.Status != shim.OK {
			//	errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", response.Payload)
			//	fmt.Printf(errStr)
			//	return shim.Error(errStr)
			//} 
			CptyID := fmt.Sprintf("%04d", i+1)
			if err != nil {
				return shim.Error("Failed to convert CptyID")
			}
			//計算CreditGuaranteeAmt信用擔保金額=MTM-OwnThreshold本行之門檻金額
			CreditGuaranteeAmt := summtm[i]  - float64(ownthreshold[i])
			//計算信用擔保餘額CreditGuaranteeBal by 前一天
			CreditGuaranteeBal := float64(0)
			//計算TXKinds （1)信用擔保金額 > 信用擔保餘額 = Cpty交付 (2)信用擔保金額 < 信用擔保餘額 = Cpty返還
			if (CreditGuaranteeAmt > CreditGuaranteeBal) {
				TXKinds = "交易對手交付"
			} else {
                TXKinds = "返還交易對手"
			}
            //計算Collateral = 信用擔保金額 - 信用擔保餘額
			Collateral := CreditGuaranteeAmt - CreditGuaranteeBal
			//計算MarginCall ＝ Collateral > CptyMTA，取整數計算(Cpty交付金額Rounding進位，我付款Rounding捨去)
            if (math.Abs(Collateral) > float64(cptymta[i])) {
				if TXKinds == "交易對手交付" {
					MarginCall = math.Ceil(Collateral / float64(rounding[i])) * float64(rounding[i])
				} else{
					MarginCall = math.Floor(Collateral / float64(rounding[i])) * float64(rounding[i])
				}
			}

			err = CreateFXTradeCollateral(APIstub, TXDATE, strconv.FormatInt(i, 16) , OwnCptyID , CptyID, summtm[i], ownthreshold[i], int64(CreditGuaranteeAmt), int64(CreditGuaranteeBal), TXKinds, int64(Collateral), int64(cptymta[i]), int64(MarginCall))
			if err != nil {
				return shim.Error("Failed to CreateFXTradeCollateral")
			}
		}
	}

	return shim.Success(nil)
}	

//peer chaincode invoke -n mycc -c '{"Args":["CreateFXTradeCollateral", "20181026","0001","0002"]}' -C myc 
func CreateFXTradeCollateral(APIstub shim.ChaincodeStubInterface, TXDATE string, TXID string, OwnCptyID string, CptyID string, MTM float64, OurThreshold int64, CreditGuaranteeAmt int64,CreditGuaranteeBal int64,TXKinds string ,Collateral int64,CptyMTA int64, MarginCall int64)  error {

	TimeNow := time.Now().Format(timelayout)
	TimeNow2 := time.Now().Format(timelayout2)
   
	TXID = OwnCptyID + CptyID + TimeNow + TXID
	TXStatus := "Pending"
	fmt.Println("- start CreateFXTradeCollateral ", TXDATE, TXID, OwnCptyID, CptyID, MTM, OurThreshold)

	var TransactionCollateral = TransactionCollateral{ObjectType: CollateralTXObjectType, TXID: TXID, TXDATE: TXDATE, OwnCptyID: OwnCptyID, CptyID: CptyID, MTM: MTM, OurThreshold: OurThreshold, CreditGuaranteeAmt:CreditGuaranteeAmt,CreditGuaranteeBal:CreditGuaranteeBal ,TXKinds:TXKinds ,Collateral:Collateral ,CptyMTA:CptyMTA ,MarginCall:MarginCall ,TXStatus:TXStatus ,CreateTime:TimeNow2}
	CollateralAsBytes, _ := json.Marshal(TransactionCollateral)
	err1 := APIstub.PutState(TransactionCollateral.TXID, CollateralAsBytes)
	if err1 != nil {
		return err1
		fmt.Println("CreateFXTradeCollateral.PutState\n") 
	}

	return nil
}

//peer chaincode invoke -n mycc -c '{"Args":["UpdateFXTradeCollateral", "00010005201812061156064","Finished"]}' -C myc 
func (s *SmartContract) UpdateFXTradeCollateral(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	TransactionCollateralAsBytes, _ := APIstub.GetState(args[0])
	TransactionCollateral := TransactionCollateral{}

	json.Unmarshal(TransactionCollateralAsBytes, &TransactionCollateral)
	TransactionCollateral.TXStatus = args[1]
	
	TransactionCollateralAsBytes, _ = json.Marshal(TransactionCollateral)
	err := APIstub.PutState(args[0], TransactionCollateralAsBytes)
	if err != nil {
		return shim.Error("Failed to change state")
	}

	return shim.Success(nil)
}

//peer chaincode invoke -n mycc -c '{"Args":["queryFXTradeCollateral", "20181208","00010002201812041256341"]}' -C myc 
//CollateralDetail的FXTXID
func (s *SmartContract) queryFXTradeCollateral(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	queryString := fmt.Sprintf("{\"selector\": {\"docType\":\"CollateralDetail\",\"TXDATE\":\"%s\",\"FXTXID\":\"%s\"}}", args[0],args[1])

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

//peer chaincode invoke -n mycc -c '{"Args":["CreateCollateralDetail", "20181208","0001","0002","TWD","Cash","A03108","10000","0.98","980","00010002201812041256341"]}' -C myc 
//peer chaincode invoke -n mycc -c '{"Args":["CreateCollateralDetail", "20181208","0001","0002","USD","Cash","A03108","70000","0.98","980","00010002201812041256341"]}' -C myc 
func (s *SmartContract) CreateCollateralDetail(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)
	TimeNow2 := time.Now().Format(timelayout2)

	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}
	var newAmount, newDiscount, newDiscountAmount float64
	var TXID = args[1] + args[2] + TimeNow 

	newAmount, err := strconv.ParseFloat(args[6], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newDiscount, err = strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	newDiscountAmount, err = strconv.ParseFloat(args[8], 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	var CollateralDetail = CollateralDetail{ObjectType: "CollateralDetail", TXID: TXID, TXDATE: args[0], OwnCptyID: args[1], CptyID: args[2], Curr: args[3], CollateralType: args[4], CollateralDetail: args[5], Amount:newAmount,Discount:newDiscount ,DiscountAmount:newDiscountAmount ,FXTXID:args[9] ,CreateTime:TimeNow2}
	CollateralDetailAsBytes, _ := json.Marshal(CollateralDetail)
	err1 := APIstub.PutState(CollateralDetail.TXID, CollateralDetailAsBytes)
	if err1 != nil {
		return shim.Error("Failed to create state")
		fmt.Println("CreateCollateralDetail.PutState\n") 
	}

	return shim.Success(nil)
}


//peer chaincode query -n mycc -c '{"Args":["queryCollateralTransactionStatus","20181026","0001"]}' -C myc
func (s *SmartContract) queryCollateralTransactionStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	var queryString string
	TXDATE := args[0]
	CptyID := args[1]

	if CptyID == "All" {		
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"Collateral\",\"TXDATE\":\"%s\"}}", TXDATE)
	} else {	
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"Collateral\",\"TXDATE\":\"%s\",\"OwnCptyID\":\"%s\"}}", TXDATE, CptyID)
	}
	 	
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)" + queryString + "\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("resultsIterator.Close")
 
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

//peer chaincode query -n mycc -c '{"Args":["queryCollateralDetailStatus","20181208","0001"]}' -C myc
func (s *SmartContract) queryCollateralDetailStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	var queryString string

	TXDATE := args[0]
	CptyID := args[1]

	if CptyID == "All" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CollateralDetail\",\"TXDATE\":\"%s\"}}", TXDATE)
	}  else {	
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CollateralDetail\",\"TXDATE\":\"%s\",\"OwnCptyID\":\"%s\"}}", TXDATE, CptyID)
	}
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)" + queryString + "\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("resultsIterator.Close")
 
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

//peer chaincode query -n mycc -c '{"Args":["queryTXIDCollateral", "0001000220181124020917"]}' -C myc
//用TXID去查詢Collateral，回傳一筆   
func (s *SmartContract) queryTXIDCollateral(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	NewTXAsBytes, _ := APIstub.GetState(args[0])
	NewTX := TransactionCollateral{}
	json.Unmarshal(NewTXAsBytes, &NewTX)

	NewTXAsBytes, err := json.Marshal(NewTX)
	if err != nil {
		return shim.Error("Failed to query NewTX state")
	}

	return shim.Success(NewTXAsBytes)
}


/*
peer chaincode invoke -n mycc -c '{"Args":["CollateralSettlment", "20181208","0001000220181208033623"]}' -C myc 
每次只能交割一筆，所以要傳入CollateralDetail的TXID
*/
func (s *SmartContract) CollateralSettlment(APIstub shim.ChaincodeStubInterface,args []string) peer.Response {
	
	//TimeNow := time.Now().Format(timelayout)

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	queryString := fmt.Sprintf("{\"selector\": {\"docType\":\"CollateralDetail\",\"TXDATE\":\"%s\",\"TXID\":\"%s\"}}", args[0],args[1])

	fmt.Println("queryString= " + queryString + "\n") 
	
	resultsIterator, err := APIstub.GetQueryResult(queryString)
    defer resultsIterator.Close()
    if err != nil {
        return shim.Error("Failed to GetQueryResult")
	}
	
	transactionArr := []CollateralDetail{}
	var recint int64= 0
	var CptyAssetID, CashFlowDetail string

	for resultsIterator.HasNext() {

        queryResponse,err := resultsIterator.Next()
        if err != nil {
			return shim.Error("Failed to Next")
		}
		fmt.Println("queryResponse.Key= " + queryResponse.Key + "\n") 
	
		jsonByteObj := queryResponse.Value
		transaction := CollateralDetail{}
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
		//fmt.Println("assetAsBytes.USD= " + strconv.FormatFloat(assetTx.USD,'f', 4, 64) + "\n") 
		//fmt.Println("assetAsBytes.OwnCptyID= " + assetTx.OwnCptyID + "\n") 
		
		fmt.Println("assetAsBytes.Curr= " + transactionArr[recint].Curr  + "\n") 
		fmt.Println("assetAsBytes.CollateralType= " + transactionArr[recint].CollateralType + "\n") 
		fmt.Println("assetAsBytes.Amount= " + strconv.FormatFloat(transactionArr[recint].Amount,'f', 4, 64) + "\n") 
		
		if transactionArr[recint].Curr ==  "USD" && transactionArr[recint].CollateralType ==  "Cash" {
			assetTx.USD = assetTx.USD - transactionArr[recint].Amount
			CashFlowDetail = "USD"
		}
		if transactionArr[recint].Curr ==  "TWD" && transactionArr[recint].CollateralType ==  "Cash" {
			assetTx.TWD = assetTx.TWD - transactionArr[recint].Amount
			CashFlowDetail = "TWD"
		}
		if transactionArr[recint].Curr ==  "USD" && transactionArr[recint].CollateralType ==  "Bond" {
			assetTx.USDBond = assetTx.USDBond - transactionArr[recint].Amount
			CashFlowDetail = "USDBond"
		}
		if transactionArr[recint].Curr ==  "TWD" && transactionArr[recint].CollateralType ==  "Bond" {
			assetTx.TWDBond = assetTx.TWDBond - transactionArr[recint].Amount
			CashFlowDetail = "TWDBond"
		}
		fmt.Println("assetAsBytes.TWD= " + strconv.FormatFloat(assetTx.TWD,'f', 4, 64) + "\n") 
		assetAsBytes, err = json.Marshal(assetTx)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = APIstub.PutState(CptyAssetID, assetAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		
		response := s.CreateCashFlow(APIstub, []string{args[0], transactionArr[recint].OwnCptyID, transactionArr[recint].CptyID, "Collateral", CashFlowDetail, strconv.FormatFloat(transactionArr[recint].Amount,'f', 4, 64), args[1]})
		   // if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			   return shim.Error("Transfer failed: " + response.Message)
		}
		if response.Status == shim.OK {
			   fmt.Println("response.Status\n")
		}
		recint += 1 

	}	
	

	return shim.Success(nil)
}	

func (s *SmartContract) CreateCashFlow(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	TimeNow := time.Now().Format(timelayout)
	TimeNow2 := time.Now().Format(timelayout2)
		
	if len(args) < 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	TXDATE :=  args[0] 
	OwnCptyID := args[1]
	CptyID := args[2]
	CashFlowType := args[3]
	CashFlowDetail := args[4]
	FXTXID := args[6]
	CashFlowAmount, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		fmt.Println("Amount1 must be a numeric string.")
	} else if CashFlowAmount < 0 {
		fmt.Println("Amount1 must be a positive value.")
	}
   
	TXID := args[1] + args[2] + TimeNow

	var CashFlow = CashFlow{ObjectType: "CashFlow", TXID: TXID, TXDATE: TXDATE, OwnCptyID: OwnCptyID, CptyID: CptyID, CashFlowType: CashFlowType, CashFlowDetail:CashFlowDetail, CashFlowAmount: CashFlowAmount, FXTXID: FXTXID, CreateTime:TimeNow2}
	CashFlowAsBytes, _ := json.Marshal(CashFlow)
	err1 := APIstub.PutState(CashFlow.TXID, CashFlowAsBytes)
	if err1 != nil {
		fmt.Println("PutState.CashFlowAsBytes= " + err1.Error() + "\n")
		return shim.Error(err1.Error())
	}

	return shim.Success(nil)
}

//peer chaincode query -n mycc -c '{"Args":["queryCashFlowStatus","20181208","0001"]}' -C myc
func (s *SmartContract) queryCashFlowStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	var queryString string

	TXDATE := args[0]
	CptyID := args[1]

	if CptyID == "All" {
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CashFlow\",\"TXDATE\":\"%s\"}}", TXDATE)
	}  else {	
		queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"CashFlow\",\"TXDATE\":\"%s\",\"OwnCptyID\":\"%s\"}}", TXDATE, CptyID)
	}
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Printf("APIstub.GetQueryResult(queryString)" + queryString + "\n")
    if err != nil {
        return shim.Error(err.Error())
    }
	defer resultsIterator.Close()
	fmt.Printf("resultsIterator.Close")
 
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