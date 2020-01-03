package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
)

func main1() {
	for index := 0; index < 1000000000; index++ {
		data := sha256.Sum256([]byte(strconv.Itoa(index)))
		fmt.Printf("%10d,%x\n", index, data)
		if string(data[len(data)-3:]) == "000" {
			fmt.Println("挖矿成功")
			break
		}
	}
}
