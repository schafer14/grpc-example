package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/schafer14/grpc-knative/messages"
)

type messageService struct{}

func newServer() messageService {
	return messageService{}
}

func (messageService) GetMessage(ctx context.Context, req *pb.Empty) (*pb.Message, error) {
	return &pb.Message{From: "Person A", Msg: "I want to send a message"}, nil
}

func (messageService) StreamMessages(req *pb.Empty, stream pb.MessageService_StreamMessagesServer) error {
	for i := 1; i <= 100; i++ {
		msg := &pb.Message{From: "Person A", Msg: fmt.Sprintf("Message %v", i)}
		stream.SendMsg(msg)
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (messageService) SendMessages(stream pb.MessageService_SendMessagesServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-stream.Context().Done():
			log.Println("Closed by client")
			break Loop
		default:
			message, err := stream.Recv()
			if err == io.EOF {
				break Loop
			}
			if err != nil {
				log.Println("An error occured recieving messages", err)
				return err
			}
			log.Printf("Recieved message from %v: %v", message.From, message.Msg)

		}
	}

	log.Println("Server no longer cares.")
	return stream.SendAndClose(&pb.Empty{})
}

func (messageService) Chat(stream pb.MessageService_ChatServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inChan := make(chan pb.Message)

	go func(input chan pb.Message) {
		for {
			in, err := stream.Recv()
			if err != nil {
				return
			}
			input <- *in
		}
	}(inChan)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Closed by server")
			return nil
		case <-stream.Context().Done():
			fmt.Println("Closed by client")
			stream.Send(&pb.Message{From: "Server", Msg: "Goodbye ol' friend"})
			return nil
		case msg := <-inChan:
			fmt.Printf("Recieved message %v\n", msg)
		}
	}
}

func main() {
	host := flag.String("host", ":8080", "The server host")

	flag.Parse()

	lis, err := net.Listen("tcp", *host)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMessageServiceServer(grpcServer, newServer())

	log.Printf("Listening on %v\n", *host)
	log.Fatal(grpcServer.Serve(lis))
}
