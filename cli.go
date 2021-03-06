package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

//CLI obj
type CLI struct {
	blockchain *BlockChain `json:"blockchain,omitempty"`
}

//CreateBlockChain 创建区块链
func (cli *CLI) CreateBlockChain(address string) {
	bc := CreateBlockChain(address)
	bc.DB.Close()
	fmt.Printf("创建成功 -> %s", address)
}

//GetBalance 查询钱包
func (cli *CLI) GetBalance(address string) {
	bc := NewBlockChain(address)
	defer bc.DB.Close()
	balance := 0
	UTXOs := bc.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("查询到地址[%s]的金额为[%d]\n", address, balance)
}

//PrintUseage 指导页
func (cli *CLI) PrintUseage() {
	fmt.Println("用法如下:")
	fmt.Println("querybalance -address 输入地址余额")
	fmt.Println("createblockchain -address 创建区块链")
	fmt.Println("send -from From -to To -amount Amount 转账")
	fmt.Println("showblockchain 显示区块链")
}

//VaiidateArgs 检验参数
func (cli *CLI) VaiidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUseage()
		os.Exit(1)
	}
}

//ShowBlockChain 显示区块链
func (cli *CLI) ShowBlockChain() {
	bc := NewBlockChain("")
	defer bc.DB.Close()
	bci := bc.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("上一块哈希:%x,当前块哈希:%x,", block.PrevBlockHash, block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow:%s\n", strconv.FormatBool(pow.Validate()))
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

//Send 进行交易
func (cli *CLI) Send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.DB.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("交易成功")
}

//Run 挖矿
func (cli *CLI) Run() {
	cli.VaiidateArgs()
	queryBalanceCmd := flag.NewFlagSet("querybalance", flag.ExitOnError)
	queryBalanceCmdAddress := queryBalanceCmd.String("address", "", "查询地址")
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createBlockChainCmdAddress := createBlockchainCmd.String("address", "", "根据地址创建")
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendCmdFrom := sendCmd.String("from", "", "发送者地址")
	sendCmdTo := sendCmd.String("to", "", "接收者者地址")
	sendCmdAmount := sendCmd.Int("amount", 0, "发送金额")
	showBlockChainCmd := flag.NewFlagSet("showblockchain", flag.ExitOnError)
	switch os.Args[1] {
	case "querybalance":
		err := queryBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "showblockchain":
		err := showBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.PrintUseage()
		os.Exit(1)
	}
	if queryBalanceCmd.Parsed() {
		if *queryBalanceCmdAddress == "" {
			queryBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.GetBalance(*queryBalanceCmdAddress)
	}
	if createBlockchainCmd.Parsed() {
		if *createBlockChainCmdAddress == "" {
			queryBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.CreateBlockChain(*createBlockChainCmdAddress)
	}
	if sendCmd.Parsed() {
		if *sendCmdFrom == "" || *sendCmdTo == "" || *sendCmdAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.Send(*sendCmdFrom, *sendCmdTo, *sendCmdAmount)
	}
	if showBlockChainCmd.Parsed() {
		cli.ShowBlockChain()
	}
}
