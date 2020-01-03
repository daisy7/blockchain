package main

func main() {
	bc := NewBlockChain()
	defer bc.DB.Close()
	cli := CLI{bc}
	cli.Run()
}
