package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	gRPC "github.com/A-Guldborg/handin-03/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", ":5400", "Tcp server")
var LamportTimeStamp int64 = 1

var server gRPC.ChittyChatClient //the server
var ServerConn *grpc.ClientConn  //the server connection
var id int64
var leftChat bool

func main() {

	fmt.Println("--- CLIENT APP ---")

	flag.Parse()

	ConnectToServer()
	joined := false

	fmt.Println("Connected to server")

	for !joined {
		fmt.Println("Please insert name:")
		fmt.Scanf("%s", clientsName)

		clientId, _ := server.Join(context.Background(), &gRPC.JoinPacket{ClientName: *clientsName})

		if clientId.Ack.StatusCode == 200 {
			joined = true
			id = clientId.Id
		}
		if clientId.Ack.StatusCode == 409 {
			fmt.Printf("Name is already taken by Client with Id: %d", clientId.Id)
		}
	}
	fmt.Println("Successfully connected, use /leave to leave the chat")

	go GetContentStream()

	parseInput()
}

func OnLeave() {
	pkt := gRPC.BasePacket{
		Id:               id,
		LamportTimeStamp: LamportTimeStamp,
	}
	leftChat = true
	server.Leave(context.Background(), &pkt)
}

func GetContentStream() {
	pkt := gRPC.BasePacket{
		Id:               id,
		LamportTimeStamp: LamportTimeStamp,
	}

	client, err := server.GetContentStream(context.Background(), &pkt)
	if err != nil {
		log.Println("Error: " + err.Error())
	}

	go func() {
		for !leftChat {
			in := gRPC.Content{}

			err := client.RecvMsg(&in)
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message from channel. \nErr: %v", err)
			}

			if id != in.BasePacket.Id {
				LamportTimeStamp = Max(in.BasePacket.LamportTimeStamp, LamportTimeStamp) + 1
				log.Printf("(%v, L:%d) -> %v \n", in.SenderName, in.BasePacket.LamportTimeStamp, in.MessageBody)
			}
		}
	}()

	for !leftChat {
		time.Sleep(2 * time.Minute)
	}
}

func sendMessage(message string) {
	msg := &gRPC.Content{
		BasePacket: &gRPC.BasePacket{
			Id:               id,
			LamportTimeStamp: LamportTimeStamp,
		},
		SenderName:  *clientsName,
		MessageBody: message,
	}
	ack, err := server.Message(context.Background(), msg)

	if err != nil {
		log.Printf("Cannot send message: error: %v", err)
	} else {
		LamportTimeStamp++
	}

	if ack.StatusCode >= 400 {
		fmt.Printf("Something went wrong sending a message")
	}
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	for !leftChat {
		validMessage := false
		var input string
		var err error
		for !validMessage {

			input, err = reader.ReadString('\n')
			if input == "/leave\n" {
				OnLeave()
				return
			}

			if len(input) > 128 {
				fmt.Println("Message is too long! the size of the message should be at most 128 characters.")
			} else if err != nil {
        fmt.Println("Error reading input")
        log.Fatal(err)
      } else {
        validMessage = true
      }
		}
		// Send to server
		go sendMessage(input)
	}
}

func ConnectToServer() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	fmt.Printf("Dialing on %s \n", *serverPort)
	conn, err := grpc.Dial(*serverPort, opts...)
	if err != nil {
		log.Fatalf("Fail to dial: %v", err)
	}

	server = gRPC.NewChittyChatClient(conn)
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
