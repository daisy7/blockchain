package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

type BlockChain struct {
	Tip []byte
	DB  *bolt.DB
}
type BlockChainIterator struct {
	CurrentHash []byte
	DB          *bolt.DB
}

const dbFile = "blockchain.db"
const BUCKET_NAME = "blocks"
const GENESIS_BLOCK_DATA = "创世区块"
const DB_LAST_KEY = "1"

//GetDbLast 取出数据库最后一个数据
func (blockchain *BlockChain) GetDbLast() []byte {
	var lashHash []byte
	err := blockchain.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		if bucket == nil {
			fmt.Println("数据库不包含区块链,创建新的区块链")
			cbtx := NewCoinBaseTX(address, GENESIS_BLOCK_DATA)
			genesis := NewGenesisBlock(cbtx)
			bucket, err := tx.CreateBucket([]byte(BUCKET_NAME))
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put([]byte(DB_LAST_KEY), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			lashHash = genesis.Hash
		} else {
			lashHash = bucket.Get([]byte(DB_LAST_KEY))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lashHash
}

//DbExists 判断数据库存在
func DbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//NewBlockChain 创建区块链
func NewBlockChain(address string) *BlockChain {
	if DbExists() == false {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
	})

	bc := BlockChain{tip, db}
	bc.Tip = bc.GetDbLast()
	return &bc
}

// MineBlock 挖矿
func (blockchain *BlockChain) MineBlock(txs []*Transaction) {
	var lastHash []byte
	err := blockchain.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		lastHash = bucket.Get([]byte(DB_LAST_KEY))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(txs, lastHash)
	err = blockchain.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte(DB_LAST_KEY), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		blockchain.Tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//FindeUnspendTransactions 查找没有花销的交易
func (blockchain *BlockChain) FindeUnspendTransactions(address string) []Transaction {
	var unspentTXs []Transaction        //交易事务
	spentTXOS := make(map[string][]int) //开辟内存
	bci := blockchain.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outindex, out := range tx.Vout {
				if spentTXOS[txID] != nil {
					for _, spentOut := range spentTXOS[txID] {
						if spentOut == outindex {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOS[inTxID] = append(spentTXOS[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

//FindUTXO 获取未花费交易输出
func (blockchain *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := blockchain.FindeUnspendTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

//FindSpendableOutputs 查找进行转账的交易
func (blockchain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutput := make(map[string][]int)
	unspentTXs := blockchain.FindeUnspendTransactions(address)
	accmulated := 0
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outindex, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accmulated < amount {
				accmulated += out.Value
				unspentOutput[txID] = append(unspentOutput[txID], outindex)
				if accmulated >= amount {
					break Work
				}
			}
		}
	}
	return accmulated, unspentOutput
}

//AddBlock 添加区块
func (blocks *BlockChain) AddBlock(data string) {
	var prevHash []byte
	err := blocks.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		prevHash = bucket.Get([]byte("1"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(data, prevHash)
	blocks.DB.Update(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(BUCKET_NAME))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("1"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		blocks.Tip = newBlock.Hash
		return nil
	})
}

//Iterator 获取迭代器
func (blocks *BlockChain) Iterator() *BlockChainIterator {
	bcit := BlockChainIterator{blocks.Tip, blocks.DB}
	return &bcit
}

//Next 下一个区块
func (it *BlockChainIterator) Next() *Block {
	var block *Block
	if it.currentHash != nil {
		err := it.db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(BUCKET_NAME))
			encodeBlock := bucket.Get(it.currentHash)
			block = Deserialize(encodeBlock)
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		it.currentHash = block.PrevBlockHash
	}
	return block
}
