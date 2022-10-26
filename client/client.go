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

	fmt.Println("--- join Server ---")

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
			// in, err := client.Recv()

			in := gRPC.Content{}

			err := client.RecvMsg(&in)
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nErr: %v", err)
			}

			if id != in.BasePacket.Id {
				LamportTimeStamp = Max(in.BasePacket.LamportTimeStamp, LamportTimeStamp) + 1

				fmt.Printf("(%v, L:%d) -> %v \n", in.SenderName, in.BasePacket.LamportTimeStamp, in.MessageBody)
			}
		}
	}()

	for !leftChat {
		time.Sleep(2 * time.Minute)
	}
}

func sendMessage(message string) {
	stream, err := server.Message(context.Background())

	if err != nil {
		log.Printf("Cannot send message: error: %v", err)
	}
	msg := gRPC.Content{
		BasePacket: &gRPC.BasePacket{
			Id:               id,
			LamportTimeStamp: LamportTimeStamp,
		},
		SenderName:  *clientsName,
		MessageBody: message,
	}
	stream.Send(&msg)
	LamportTimeStamp++

	ack, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("Error while sending: %s \n", err)
	}

	if ack.StatusCode >= 400 {
		fmt.Printf("Something went wrong sending a message")
	}
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	for !leftChat {

		// fmt.Print(*clientsName + " -> ")
		input, err := reader.ReadString('\n')
		if input == "/leave\n" {
			OnLeave()
			return
		}
		// Send to server
		go sendMessage(input)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func ConnectToServer() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	fmt.Printf("Dialing on %s \n", *serverPort)
	conn, err := grpc.Dial(*serverPort, opts...)
	fmt.Println("Dialed")
	if err != nil {
		log.Fatalf("Fail to dial: %v", err)
	}
	fmt.Println("Didn't fail to dial")

	fmt.Println(conn.GetState().String())

	server = gRPC.NewChittyChatClient(conn)
	if err != nil {
		log.Fatalf("Fail to dial: %v", err)
	}
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
