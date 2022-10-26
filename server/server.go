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
	names    []string
	nextId   int
	port     string
}

var port = flag.String("port", "5400", "Server port")

func newServer() *Server {
	s := &Server{
		port:     "5400",
		channels: make([]chan *gRPC.Content, 0),
		names:    make([]string, 0),
		nextId:   0,
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

func (s *Server) Join(ctx context.Context, joinPacket *gRPC.JoinPacket) (*gRPC.ClientId, error) {
	for i, name := range s.names {
		if name == joinPacket.ClientName {
			return &gRPC.ClientId{Id: int64(i), Ack: &gRPC.Ack{StatusCode: 409}}, nil
		}
	}
	availableId := &gRPC.ClientId{Id: int64(s.nextId), Ack: &gRPC.Ack{StatusCode: 200}}
	s.names = append(s.names, joinPacket.ClientName)
	log.Printf("Name is available, new client given id %d", s.nextId)
	in := &gRPC.Content{
		SenderName:  s.names[s.nextId],
		MessageBody: fmt.Sprintf("%s joined the chat at L(%d)", s.names[s.nextId], 0),
		BasePacket: &gRPC.BasePacket{
			Id:               int64(s.nextId),
			LamportTimeStamp: 0,
		},
	}

	go func() {
		streams := s.channels
		for _, msgChan := range streams {
			if msgChan != nil {
				msgChan <- in
			}
		}
	}()
	s.nextId = s.nextId + 1

	return availableId, nil
}

func (s *Server) GetContentStream(BasePacket *gRPC.BasePacket, stream gRPC.ChittyChat_GetContentStreamServer) error {
	// Client joins with id
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

func (s *Server) Leave(ctx context.Context, basepacket *gRPC.BasePacket) (*gRPC.Ack, error) {
	var id = basepacket.Id
	var LamportTimeStamp = basepacket.LamportTimeStamp

	// Broadcast stuff
	in := &gRPC.Content{
		SenderName:  s.names[id],
		MessageBody: fmt.Sprintf("%s left the chat at L(%d)", s.names[id], LamportTimeStamp),
		BasePacket: &gRPC.BasePacket{
			Id:               id,
			LamportTimeStamp: LamportTimeStamp,
		},
	}

	go func() {
		streams := s.channels
		for _, msgChan := range streams {
			if msgChan != nil {
				msgChan <- in
			}
		}
	}()

	log.Printf("%s left the chat at L(%d)", s.names[id], LamportTimeStamp)
	s.channels[id] = nil
	s.names[id] = ""

	// Close connection to client from serverside

	return &gRPC.Ack{StatusCode: 200}, nil
}

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

	log.Printf("%d - %s -> %s ", in.BasePacket.LamportTimeStamp, s.names[in.BasePacket.Id], in.MessageBody) // log message

	ack := gRPC.Ack{StatusCode: 200}
	stream.SendAndClose(&ack)

	go func() {
		streams := s.channels
		for _, msgChan := range streams {
			if msgChan != nil {
				msgChan <- in
			}
		}
	}()

	return nil
}
