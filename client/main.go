package main

import (
	"context"
	"flag"
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
			log.Println(msg)
		}
	}
	return
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

	singleRequest(c)
	manyRequests(c)

	// // Get all messages
	// streamCtx, streamCancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer streamCancel()
	// stream, err := c.StreamMessages(streamCtx, &pb.Empty{})
	// if err != nil {
	// 	log.Fatalf("%v.StreamMessages(_) = _, %v", c, err)
	// }

	// for {
	// 	msg, err := stream.Recv()
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Printf("%v.StreamMessages(_) = _, %v", c, err)
	// 		break
	// 	}
	// 	log.Println(msg)
	// }

	// // Send messages
	// // Create a random number of random points
	// sendCtx, sendCancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer sendCancel()
	// msgStream, err := c.SendMessages(sendCtx)
	// if err != nil {
	// 	log.Fatalf("%v.RecordRoute(_) = _, %v", c, err)
	// }
	// for i := 0; i < 5; i++ {
	// 	if err := msgStream.Send(&pb.Message{From: "Client", Msg: fmt.Sprintf("Message number %v", i)}); err != nil {
	// 		log.Fatalf("%v.Send() = %v", stream, err)
	// 	}
	// 	time.Sleep(time.Second)
	// }
	// _, err = msgStream.CloseAndRecv()
	// if err != nil {
	// 	log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	// }
	// log.Printf("Recieved closure from server")

	// // Chat
	// chatCtx, chatCancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer chatCancel()
	// chatStream, err := c.Chat(chatCtx)
	// if err != nil {
	// 	log.Fatalf("%v.Chat(_) = _, %v", c, err)
	// }
	// for i := 0; i < 5; i++ {
	// 	if err := chatStream.Send(&pb.Message{From: "Client", Msg: fmt.Sprintf("Chat message number %v", i)}); err != nil {
	// 		log.Fatalf("%v.Send() = %v", stream, err)
	// 	}
	// 	time.Sleep(time.Second)
	// }
	// chatCancel()
}
