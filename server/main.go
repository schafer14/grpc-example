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
		select {
		case <-stream.Context().Done():
			log.Printf("Connection closed by client\n")
			return nil
		default:
			msg := &pb.Request{From: "Streaming Server", Body: fmt.Sprintf("Message %v", i)}
			stream.SendMsg(msg)
			fmt.Printf("Sending %v\n", i)
			time.Sleep(250 * time.Millisecond)
		}
	}
	return nil
}

// ClientStreamRequests listens to messages from a client and then at the end of the messages sends
// a response to notify of receipt.
func (requestService) ClientStreamRequests(stream pb.RequestService_ClientStreamRequestsServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inChan := make(chan pb.Request)
	go func(ch chan pb.Request) {
		for {
			message, err := stream.Recv()
			if err != nil {
				cancel()
				return
			}
			ch <- *message
		}
	}(inChan)

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-stream.Context().Done():
			log.Println("Closed by client")
			break Loop
		case message := <-inChan:
			log.Println(&message)
		}
	}

	log.Println("Final response")
	return stream.SendAndClose(&pb.Request{From: "Listening Server", Body: "Done"})
}

func (requestService) BidirectionalRequests(stream pb.RequestService_BidirectionalRequestsServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inChan := make(chan pb.Request)
	go func(input chan pb.Request) {
		for {
			in, err := stream.Recv()
			if err != nil {
				cancel()
				return
			}
			input <- *in
		}
	}(inChan)

	outChan := make(chan pb.Request)
	go func(out chan pb.Request) {
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				out <- pb.Request{From: "Bidirectional Server", Body: fmt.Sprintf("Message %v", i)}
				time.Sleep(time.Millisecond * 150)
			}
		}
	}(outChan)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Closed by server")
			return nil
		case <-stream.Context().Done():
			fmt.Println("Closed by client")
			return nil
		case message := <-outChan:
			stream.Send(&message)
		case msg := <-inChan:
			fmt.Println(&msg)
			stream.Send(&pb.Request{From: "Server", Body: "I got your message!"})
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

	pb.RegisterRequestServiceServer(grpcServer, newService())

	log.Printf("Listening on %v\n", *host)
	log.Fatal(grpcServer.Serve(lis))
}
