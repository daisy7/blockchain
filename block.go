package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

//Block 定义区块
type Block struct {
	Timestamp     int64          //时间
	Transactions  []*Transaction //交易
	PrevBlockHash []byte         //前驱哈希
	Hash          []byte         //哈希
	Nonce         int            //难度值
}

//HashTransactions 计算交易哈希
func (block *Block) HashTransactions() []byte {
	var txHashs [][]byte
	var txHash [32]byte
	for _, tx := range block.Transactions {
		txHashs = append(txHashs, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashs, []byte{}))
	return txHash[:]
}

//NewBlock 创建一个区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		time.Now().Unix(),
		transactions,
		prevBlockHash,
		[]byte{},
		0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

//NewGenesisBlock 创建创世区块
func NewGenesisBlock(transaction *Transaction) *Block {
	return NewBlock([]*Transaction{transaction}, []byte{})
}

//Serialize 序列化
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//Deserialize 反序列化
func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
