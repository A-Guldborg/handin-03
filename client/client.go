package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gRPC "github.com/A-Guldborg/handin-03/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")
var LamportTimeStamp int64 = 1

var server gRPC.ChittyChatClient //the server
var ServerConn *grpc.ClientConn  //the server connection

func main() {

	fmt.Println("Please insert name:")
	fmt.Scanf("%s", clientsName)

	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	fmt.Println("--- join Server ---")

	ConnectToServer()
	defer ServerConn.Close()

	JoinServer()

	parseInput()
}

func JoinServer() {
	server.Join(context.Background(), &gRPC.BasePacket{ClientName: *clientsName, LamportTimeStamp: LamportTimeStamp})
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)

		if !conReady(server) {
			log.Printf("Client %s: Something was wrong with connection :(", *clientsName)
			continue
		}

		// Send to server
		server.Message(context.Background(), &gRPC.Content{ClientName: *clientsName, LamportTimeStamp: LamportTimeStamp, MessageBody: input})
	}
}

func ConnectToServer() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//use context for timeout on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel() //cancel the connection when we are done

	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	server = gRPC.NewChittyChatClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func conReady(s gRPC.ChittyChatClient) bool {
	return ServerConn.GetState().String() == "READY"
}
