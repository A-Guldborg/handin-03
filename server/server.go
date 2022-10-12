package main

import (
	"context"
	"flag"
	"log"
	"net"

	gRPC "github.com/A-Guldborg/handin-03/proto"
	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChittyChatServer

	port string
}

var port = flag.String("port", "5400", "Server port")

func main() {
	listener, _ := net.Listen("tcp", "localhost:5400")
	grpcServer := grpc.NewServer()
	server := &Server{
		port: *port,
	}

	gRPC.RegisterChittyChatServer(grpcServer, server)
	grpcServer.Serve(listener)
}

func (s *Server) Join(ctx context.Context, basepacket *gRPC.BasePacket) (*gRPC.BasePacket, error) {
	var clientName = basepacket.ClientName
	var LamportTimeStamp = basepacket.LamportTimeStamp

	// Broadcast stuff

	log.Printf("%s joined the chat at L(%d)", clientName, LamportTimeStamp)

	return basepacket, nil
}

func (s *Server) Leave(ctx context.Context, basepacket *gRPC.BasePacket) (*gRPC.BasePacket, error) {
	var clientName = basepacket.ClientName
	var LamportTimeStamp = basepacket.LamportTimeStamp

	// Broadcast stuff

	log.Printf("%s left the chat at L(%d)", clientName, LamportTimeStamp)

	return basepacket, nil
}

func (s *Server) Message(ctx context.Context, message *gRPC.Content) (*gRPC.BasePacket, error) {
	var clientName = message.ClientName
	var LamportTimeStamp = message.LamportTimeStamp
	var body = message.MessageBody

	// Broadcast stuff

	log.Printf("%s L(%d): %s", clientName, LamportTimeStamp, body)

	var basepacket = gRPC.BasePacket{
		ClientName: clientName,
		LamportTimeStamp: LamportTimeStamp,
	}
	
	return &basepacket, nil
}
