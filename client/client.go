package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	gRPC "github.com/A-Guldborg/handin-03/proto"
	"google.golang.org/grpc"
)

var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", ":5400", "Tcp server")
var LamportTimeStamp int64 = 1

var server gRPC.ChittyChatClient //the server
var ServerConn *grpc.ClientConn  //the server connection

func main() {

	fmt.Println("--- CLIENT APP ---")
	fmt.Println("Please insert name:")
	fmt.Scanf("%s", clientsName)

	flag.Parse()

	fmt.Println("--- join Server ---")

	ConnectToServer()

	fmt.Println("Connected to server")
	defer ServerConn.Close()

	go JoinServer()

	parseInput()
}

func JoinServer() {
	log.Println("Hello world")
	pkt := gRPC.BasePacket{
		LamportTimeStamp: LamportTimeStamp,
		ClientName:       *clientsName,
	}
	client, err := server.Join(context.Background(), &pkt)
	if err != nil {
		log.Println("Error: " + err.Error())
	}

	incuming := make(chan struct{})

	go func() {
		for {
			// in, err := client.Recv()

			in := gRPC.Content{}

			err := client.RecvMsg(&in)
			if err == io.EOF {
				close(incuming)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nErr: %v", err)
			}

			if *clientsName != in.BasePacket.ClientName {
				fmt.Printf("MESSAGE: (%v) -> %v \n", in.BasePacket.ClientName, in.MessageBody)
			}
		}
	}()

	<-incuming

	// confirmation, _ := client.Recv()
	// log.Println(confirmation.MessageBody)
}

func sendMessage(message string) {
	stream, err := server.Message(context.Background())

	if err != nil {
		log.Printf("Cannot send message: error: %v", err)
	}
	msg := gRPC.Content{
		BasePacket: &gRPC.BasePacket{
			ClientName:       *clientsName,
			LamportTimeStamp: 1 + 1,
		},
		MessageBody: message,
	}
	stream.Send(&msg)

	ack, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("Error while sending: %s \n", err)
	} else {
		fmt.Printf("Message sent: %v \n", ack)
	}
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	defer ServerConn.Close()
	for {

		fmt.Print(*clientsName + " -> ")
		input, err := reader.ReadString('\n')

		if input == "leave" {
			return
		}

		go sendMessage(input)

		if err != nil {
			log.Fatal(err)
		}
		// Send to server
		// server.Message(context.Background(), &gRPC.Content{BasePacket: &gRPC.BasePacket{ClientName: *clientsName, LamportTimeStamp: LamportTimeStamp}, MessageBody: input})
		LamportTimeStamp++
	}
}

// func ConnectToServer() {
// 	var opts []grpc.DialOption
// 	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())

// 	//use context for timeout on the connection
// 	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel() //cancel the connection when we are done

// 	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
// 	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", *serverPort), opts...)
// 	if err != nil {
// 		log.Printf("Fail to Dial : %v", err)
// 		return
// 	}

// 	server = gRPC.NewChittyChatClient(conn)
// 	ServerConn = conn
// 	log.Println("the connection is: ", conn.GetState().String())
// }

func ConnectToServer() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	fmt.Printf("Dialing on %s \n", *serverPort)
	conn, err := grpc.Dial(*serverPort, opts...)
	fmt.Println("Dialed")
	if err != nil {
		log.Fatalf("Fail to dail: %v", err)
	}
	fmt.Println("Didn't fail to dial")

	fmt.Println(conn.GetState().String())

	server = gRPC.NewChittyChatClient(conn)
	if err != nil {
		log.Fatalf("Fail to dail: %v", err)
	}
}
