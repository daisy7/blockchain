package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const SUBSIDY = 10

type TXInput struct {
	Txid      []byte //交易id
	Vout      int    //交易的output索引
	ScriptSig string //任意用户定义的钱包地址
}

//CanUnlockOutputWith 判断是否可以花费
func (input *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

type TXOutput struct {
	Value        int    //币的数量
	ScriptPubkey string //输出给的地址
}

func (output *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return output.ScriptPubkey == unlockingData
}

type Transaction struct {
	ID   []byte //交易id
	Vin  []TXInput
	Vout []TXOutput
}

//IsCoinBase 检查交易是否为coinbase
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//SetID 设置交易id
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

//NewCoinBaseTX 挖矿奖励
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("挖矿奖励给%s", to)
	}
	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{SUBSIDY, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	return &tx
}
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var input []TXInput
	var output []TXOutput
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("交易金额不足")
	}
	for txid, outs := range validOutputs {
		txid, err := hex.Decode(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TXInput{txid, out, from}
			inputs = append(inputs, input)
		}
	}
	//交易叠加
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		// 记录以后的交易
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
