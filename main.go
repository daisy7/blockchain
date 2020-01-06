package main

import "fmt"

func main() {
	fmt.Printf("%x\n", Encode([]byte("abcdefg")))
	fmt.Printf("%s\n", Decode(Encode([]byte("abcdefg"))))
	cli := CLI{}
	cli.Run()
}
