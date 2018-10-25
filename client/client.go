package main

import (
	"flag"
	"fmt"
	"net/rpc"

	"github.com/Matrix86/sentrygram/core"
)

var (
	addr     = "127.0.0.1"
	port     = "8082"
	text     = ""
	username = ""
)

func printBanner() {
	fmt.Println(core.Name + " " + core.Version)
	fmt.Println("Author: " + core.Author)
	fmt.Println("Usage:")
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&addr, "a", "127.0.0.1", "Address of the server.")
	flag.StringVar(&port, "p", "8082", "Port of the server.")
	flag.StringVar(&text, "t", "", "Text to send (required).")
	flag.StringVar(&username, "u", "", "Username of the receiver (required).")

	flag.Parse()

	if username == "" || text == "" {
		printBanner()
		return
	}

	client, err := rpc.Dial("tcp", addr+":"+port)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer client.Close()

	request := &core.RpcRequest{Text: text, UserName: username}
	response := new(core.RpcResponse)

	err = client.Call("Handler.SendMessage", request, response)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.Message)
	}
}
