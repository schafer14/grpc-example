# gRPC Study pt 2: Simple App

The previous post discussed the motivations and benefits of using gRPC. This post will focus on getting started building gRPC applications in [go](https://golang.org/). Most of this post will run parallel to the go documentation on the [gRPC examples website](https://github.com/grpc/grpc-go/tree/master/examples). This post will deviate from those docs particularly in handling bidirectional requests in a concurrent non-blocking fashion.

## Before you start

Make sure go is installed on your computer. You will also need to install gRPC, go protobuf generator. To get setup you can follow install documentation [here](https://grpc.io/docs/quickstart/go/).

---

## Project scaffold

This project will be much simpler then the application we'll build in the next post; as such, we will keep the directory structure simple. We will add most of the source code to just two files. While this is a great way to start building software it is worth pointing out that this structure will not scale to larger projects.

Every project needs a mod file so create that

```bash
# replace my git repository with your own
go mod init github.com/schafer14/grpc-example
```

## The proto file

Let us start by creating a module that will hold our proto file and all the generated code associated with it. Create a file `requests/service.proto`. gRPC supports unary, server streaming, client streaming, and bidirectional streaming requests. We will create a single endpoint for each request type.

```proto
syntax="proto3";

message Request {
    string body = 1;
    string from = 2;
}

message Empty{}

service RequestService {
    rpc GetRequest(Empty) returns (Request) {}  
    rpc ServerStreamRequests(Empty) returns (stream Request) {}  
    rpc ClientStreamRequests(stream Request) returns (Request) {}  
    rpc BidirectionalRequests(stream Request) returns (stream Request) {}  
}
```

First two types are defined for use in our service, a request type and an empty type. The request type has a couple fields and the empty type is empty. It is worth noting that even though all fields are strings in this example protobufs support a rich [list of types](https://developers.google.com/protocol-buffers/docs/proto3) and support utilities to make composite types.

Next a service is created. This contains a list of functions that our service will implement. We use one function for each of the four types of functions gRPC supports. Notice that streams are denoted by the `stream` keyword. All other requests are assumed to be single messages.

We are now ready to generate the supporting code for our service.

```bash
protoc --go_out=plugins=grpc:. requests/service.proto
```

This creates a second file in our requests module. This file defines an interface that our service must implement, once that service is implemented the generated code contains functions for both servers and clients to interact with each other.

## Building a server

Building a server that serves this service is pretty straight forward now. We need to implement the interface generated in the previous step, and then define some configuration to run the server.

A good starting point is implementing the generated interface. This interface has a single function for each function in our proto file. We will go through them one by one. This code will be in a new module so create a file `server/main.go` in the root directory.

Implement a struct that will implement the generated interface.

```go
// server/main.go
package main

import (
    // replace with the git repository you used in the go mod init command
    pb "github.com/schafer14/grpc-example/requests"
    // ...
)

type requestService struct {
    // this is a good place for dependencies such as databases and loggers
}

func newService() requestService {
    return requestService{}
}
```

This begins to describe the struct `requestService` that will implement all the functions we need to run our service.

### GetRequests (Unary)

```go
// GetRequest returns a single request object
func (requestService) GetRequest(ctx context.Context, req *pb.Empty) (*pb.Request, error) {
    return &pb.Request{From: "Server", Body: "I have a single message for you."}, nil
}
```

If you are not used to working with protobufs this snippet might look strange, but it is not complicated at all. Remember the file that `protoc` generated for us? It defined a new type for each of the messages in our service. All we are doing here is using those types. If you ever wonder what type you need referring to that file can be very helpful.

Since this function has a signature that takes a single empty object and returns a single request object without any streams the function is familiar. 

### ServerStreamRequests (Server Streaming)

In this type of request the server will stream requests to the client. Either the server or the client can end the connection at any time. 


