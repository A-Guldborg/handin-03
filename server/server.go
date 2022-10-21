package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	gRPC "github.com/A-Guldborg/handin-03/proto"
	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChittyChatServer

	channels []chan *gRPC.Content
	port     string
}

var port = flag.String("port", "5400", "Server port")

func newServer() *Server {
	s := &Server{
		port:     "5400",
		channels: make([]chan *gRPC.Content, 0),
	}
	fmt.Println(s)
	return s
}

func main() {
	// Start server up to be listened to
	listener, err := net.Listen("tcp", "localhost:"+*port)
	log.Printf("Listening on port %s \n", *port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	gRPC.RegisterChittyChatServer(grpcServer, newServer())
	grpcServer.Serve(listener)
}

func (s *Server) Join(basepacket *gRPC.BasePacket, stream gRPC.ChittyChat_JoinServer) error {
	// Client joins
	newChannel := make(chan *gRPC.Content)
	s.channels = append(s.channels, newChannel)

	log.Println("Server served")

	// doing this never closes the stream
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg := <-newChannel:
			log.Printf("GO ROUTINE (got message): %v \n", msg)
			stream.Send(msg)
		}
	}
}

func (s *Server) Leave(ctx context.Context, basepacket *gRPC.BasePacket) (*gRPC.BasePacket, error) {
	var clientName = basepacket.ClientName
	var LamportTimeStamp = basepacket.LamportTimeStamp

	// Broadcast stuff

	log.Printf("%s left the chat at L(%d)", clientName, LamportTimeStamp)

	return basepacket, nil
}

// func (s *Server) Message(ctx context.Context, message *gRPC.Content) (*gRPC.BasePacket, error) {
// 	var clientName = message.BasePacket.ClientName
// 	var LamportTimeStamp = message.BasePacket.LamportTimeStamp
// 	var body = message.MessageBody

// 	log.Printf("%s joined the chat at L(%d)", clientName, LamportTimeStamp)

// 	// Broadcast stuff

// 	//

// 	log.Printf("%s L(%d): %s", clientName, LamportTimeStamp, body)

// 	var basepacket = gRPC.BasePacket{
// 		ClientName:       clientName,
// 		LamportTimeStamp: LamportTimeStamp,
// 	}

// 	return &basepacket, nil
// }

func (s *Server) Message(stream gRPC.ChittyChat_MessageServer) error {
	// Server connection starts
	log.Println("Server connection started")

	log.Println("Trying to receive")
	in, err := stream.Recv()
	log.Println("message recieved?")
	if err == io.EOF {
		return nil // Client stops connection
	} else if err != nil {
		return err // Error happened
	}

	log.Printf("%d - %s -> %s ", in.BasePacket.LamportTimeStamp, in.BasePacket.ClientName, in.MessageBody) // log message

	ack := gRPC.Ack{StatusCode: 200}
	stream.SendAndClose(&ack)

	go func() {
		streams := s.channels
		for _, msgChan := range streams {
			msgChan <- in
		}
	}()

	return nil
}
