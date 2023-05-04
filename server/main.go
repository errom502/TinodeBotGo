package main

import (
	"csToGo25042023/CSbotToGo"
	"fmt"
	pbx "github.com/tinode/chat/pbx"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
)

var c CSbotToGo.ChatBot

func main() {
	c.Server = grpc.NewServer()
	chatBotPlugin := &c.ChatBotPlugin

	pbx.RegisterPluginServer(c.Server, chatBotPlugin.PluginServer)
	lise := "0.0.0.0:40051"
	listenHost := strings.Split(lise, ":")[0]
	listenPort, err := strconv.Atoi(strings.Split(lise, ":")[1])
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenHost, listenPort))
	if err != nil {
		panic(err)
	}
	//wg := new(sync.WaitGroup)
	//wg.Add(1)
	//go func() {
	//	err = c.Server.Serve(lis)
	//	wg.Done()
	//}()
	//go func() {
	//	if err := c.Server.Serve(lis); err != nil {
	//		log.Fatalf("failed to serve: %v", err)
	//	}
	//	log.Println("Server connect c")
	//}()
	fmt.Println("Server started")

	err = c.Server.Serve(lis)
	if err != nil {
		panic(err)
	}
}
