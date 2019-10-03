package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/schafer14/grpc-example/requests"
	"google.golang.org/grpc"
)

func singleRequest(client pb.RequestServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.GetRequest(ctx, &pb.Empty{})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Println(r)
}

func manyRequests(client pb.RequestServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stream, err := client.ServerStreamRequests(ctx, &pb.Empty{})
	defer stream.CloseSend()
	if err != nil {
		log.Fatalf("server stream requests failed: %v", err)
	}

	reqCh := make(chan pb.Request)

	go func(ch chan pb.Request) {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Println("Server closed the stream")
				return
			}
			if err != nil {
				log.Printf("server stream requests failed to get message: %v", err)
				return
			}
			ch <- *msg
		}
	}(reqCh)

Loop:
	for {
		select {
		case <-ctx.Done():
			log.Println("Client closed the stream")
			break Loop
		case <-stream.Context().Done():
			log.Println("connection closed")
			break Loop
		case msg := <-reqCh:
			log.Println(&msg)
		}
	}
	return
}

func sendMany(client pb.RequestServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	stream, err := client.ClientStreamRequests(context.Background())
	if err != nil {
		log.Fatalf("Failed to get client stream: %v\n", err)
	}

Loop:
	for i := 0; i < 100; i++ {
		select {
		case <-ctx.Done():
			break Loop
		case <-stream.Context().Done():
			break Loop
		default:
			err := stream.Send(&pb.Request{From: "Client", Body: fmt.Sprintf("Message %v", i)})
			if err != nil {
				break
			}
			time.Sleep(time.Millisecond * 300)
		}
	}

	result, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Could not recieve message from server: %v\n", err)
	}
	log.Println(result)
	return
}

func bidirectional(client pb.RequestServiceClient) {
	log.Println("ehere")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.BidirectionalRequests(context.Background())
	if err != nil {
		log.Fatalf("Failed to get bidirectional stream: %v\n", err)
	}

	inChan := make(chan pb.Request)
	go func(ch chan pb.Request) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := stream.Recv()
				if err != nil {
					cancel()
					return
				}
				ch <- *msg
			}
		}
	}(inChan)

	outChan := make(chan pb.Request)
	go func(ch chan pb.Request) {
	Loop:
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				break Loop
			default:
				if err := stream.Send(&pb.Request{From: "Bidirectional Client", Body: fmt.Sprintf("Message %v", i)}); err != nil {
					log.Fatalf("Could not send message %v\n", err)
				}
				time.Sleep(time.Millisecond * 400)
			}
		}
		cancel()
	}(outChan)

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case msg := <-inChan:
			log.Println(&msg)
		default:
		}
	}
}

func main() {
	host := flag.String("host", ":8080", "The server host")

	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*host, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRequestServiceClient(conn)

	// singleRequest(c)
	// manyRequests(c)
	// sendMany(c)
	bidirectional(c)
}
