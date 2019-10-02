package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/schafer14/grpc-example/requests"
)

type requestService struct {
	// this is a good place for dependencies such as databases and loggers
}

func newService() requestService {
	return requestService{}
}

// GetRequest returns a single request object
func (requestService) GetRequest(ctx context.Context, req *pb.Empty) (*pb.Request, error) {
	return &pb.Request{From: "Server", Body: "I have a single message for you."}, nil
}

// ServerStreamRequests streams 100 request objects to the client
func (requestService) ServerStreamRequests(req *pb.Empty, stream pb.RequestService_ServerStreamRequestsServer) error {
	for i := 1; i <= 100; i++ {
		msg := &pb.Request{From: "Streaming Server", Body: fmt.Sprintf("Message %v", i)}
		stream.SendMsg(msg)
		time.Sleep(250 * time.Millisecond)
	}
	return nil
}

// func (messageService) SendMessages(stream pb.MessageService_SendMessagesServer) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// Loop:
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			break Loop
// 		case <-stream.Context().Done():
// 			log.Println("Closed by client")
// 			break Loop
// 		default:
// 			message, err := stream.Recv()
// 			if err == io.EOF {
// 				break Loop
// 			}
// 			if err != nil {
// 				log.Println("An error occured recieving messages", err)
// 				return err
// 			}
// 			log.Printf("Recieved message from %v: %v", message.From, message.Msg)

// 		}
// 	}

// 	log.Println("Server no longer cares.")
// 	return stream.SendAndClose(&pb.Empty{})
// }

// func (messageService) Chat(stream pb.MessageService_ChatServer) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	inChan := make(chan pb.Message)

// 	go func(input chan pb.Message) {
// 		for {
// 			in, err := stream.Recv()
// 			if err != nil {
// 				return
// 			}
// 			input <- *in
// 		}
// 	}(inChan)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			fmt.Println("Closed by server")
// 			return nil
// 		case <-stream.Context().Done():
// 			fmt.Println("Closed by client")
// 			stream.Send(&pb.Message{From: "Server", Msg: "Goodbye ol' friend"})
// 			return nil
// 		case msg := <-inChan:
// 			fmt.Printf("Recieved message %v\n", msg)
// 		}
// 	}
// }

func main() {
	host := flag.String("host", ":8080", "The server host")

	flag.Parse()

	lis, err := net.Listen("tcp", *host)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterRequestServiceServer(grpcServer, newService())

	log.Printf("Listening on %v\n", *host)
	log.Fatal(grpcServer.Serve(lis))
}
