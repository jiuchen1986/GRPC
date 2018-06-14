# Study Notes
Records knowledge from studying gRPC

## Look At The Generated Codes From Protobuf
The gRPC uses the protobuf to define messages, RPCs and services establishing the end-to-end RPC, which helps to generate codes for both the client and the server serializing/deserializing messages.

### Message Definitions
Assuming a message called `Req` is defined like

    message Req {
      int32 req_id = 1;
      repeated ReqBody req_body = 2;
    }

with a message called `ReqBody` is defined like

    message ReqBody {
      string body = 1;
    }

in the protobuf file, two structs will be generated in the `.pb.go` file as

    type Req struct {
    	ReqId int32 `protobuf:"varint,1,opt,name=req_id,json=reqId" json:"req_id,omitempty"`
    	ReqBody []*ReqBody `protobuf:"bytes,2,rep,name=req_body,json=reqBody" json:"req_body,omitempty"`
    }

    type ReqBody struct {
        Body string `protobuf:"bytes,1,opt,name=body,json=body" json:"body,omitempty"
    }

For each struct, several functions are generated like

    func (m *Req) Reset(){ *m = Req{} }
    func (m *Req) String() string{ return proto.CompactTextString(m) }
    func (*Req) ProtoMessage()   {}
    func (*Req) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

with getter functions for each element in the message like

    func (m *Req) GetReqId() int32 {...}
    func (m *Req) GetReqBody() []*ReqBody {...}


### Service Definitions
Assuming a service with 4 RPCs is defined in the protobuf file like

    service MyGrpc {
      
      // Simple RPC
      rpc SimpleCall(Req) returns (Resp) {}
    
      // Server-side streaming RPC
      rpc ServerStreamCall(Req) returns (stream Resp) {}
    
      // Client-side streaming RPC 
      rpc ClientStreamCall(stream Req) returns (Resp) {}
    
      // Bidirectional streaming RPC
      rpc BiStreamCall(stream Req) returns (stream Resp) {}
    
    }

An interface, containing RPC-related functions, with an according struct will be generated for the client as follows
    
    // Client API for MyGrpc service
    
    type MyGrpcClient interface {
    	// Simple RPC
    	SimpleCall(ctx context.Context, in *Req, opts ...grpc.CallOption) (*Resp, error)
    	// Server-side streaming RPC
    	ServerStreamCall(ctx context.Context, in *Req, opts ...grpc.CallOption) (MyGrpc_ServerStreamCallClient, error)
    	// Client-side streaming RPC
    	ClientStreamCall(ctx context.Context, opts ...grpc.CallOption) (MyGrpc_ClientStreamCallClient, error)
    	// Bidirectional streaming RPC
    	BiStreamCall(ctx context.Context, opts ...grpc.CallOption) (MyGrpc_BiStreamCallClient, error)
    }
    
    type myGrpcClient struct {
    	cc *grpc.ClientConn
    }

Meanwhile, for the server, only an interface without the implementing struct will be generated as follows

    
    // Server API for MyGrpc service

    type MyGrpcServer interface {
	    // Simple RPC
	    SimpleCall(context.Context, *Req) (*Resp, error)
	    // Server-side streaming RPC
	    ServerStreamCall(*Req, MyGrpc_ServerStreamCallServer) error
	    // Client-side streaming RPC
	    ClientStreamCall(MyGrpc_ClientStreamCallServer) error
	    // Bidirectional streaming RPC
	    BiStreamCall(MyGrpc_BiStreamCallServer) error
    }



### RPC Definitions
Basically 4 types of RPC could be defined using gRPC, generating different functions in codes.

#### Simple RPC
In a simple RPC, the client sends a request to the server using the stub and waits for a response to come back, just like a normal function call. Usually it is defined in the protobuf file like:

    rpc SimpleCall(Req) returns (Resp) {}

where the Req and the Resp are message objects defined in the protobuf file for the same gRPC processing.

With the simple RPC, a function for the client side is generated like

    SimpleCall(ctx context.Context, in *Req, opts ...grpc.CallOption) (*Resp, error)

which returns an output and terminates for each time called with a single input in a block way.

And for the server side, a similar function is generated like

    SimpleCall(ctx context.Context, in *Req) (*Resp, error)

Which is called and returns once receiving a request from a client a time.

#### Server-Side Streaming RPC
In a server-side streaming RPC, the client sends a request to the server and gets a stream to read a sequence of messages back. The client reads from the returned stream until there are no more messages. Usually it is defined like:

    ServerStreamCall(Req) returns (stream Resp)

This will generate a function for the client side like

    ServerStreamCall(ctx context.Context, in *Req, opts ...grpc.CallOption) (MyGrpc_ServerStreamCallClient, error)

which is called with a single input and returns an object of a specific client interface `MyGrpc_ServerStreamCallClient` for this RPC. A `grpc.ClientStream` object and a `Recv() (*Resp, error)` function is defined in the client interface like

    type MyGrpc_ServerStreamCallClient interface {
    	Recv() (*Resp, error)
    	grpc.ClientStream
    }

Hence the client could use the `Recv() (*Resp, error)` function to continuously read messages out from the stream until no more messages are received.

For the server side, a function is generated like

    ServerStreamCall(*Req, MyGrpc_ServerStreamCallServer) error

where an object of a specific server interface `MyGrpc_ServerStreamCallServer` is created like

    type MyGrpc_ServerStreamCallServer interface {
    	Send(*Resp) error
    	grpc.ServerStream
    }

with a `grpc.ServerStream` object and a `Send(*Resp)` function. Then the server can use the `Send(*Resp)` function to send back responses to the client.


#### Client-Side Streaming RPC
In a client-side streaming RPC, the client writes a sequence of messages and sends them to the server, again using a provided stream. Once the client has finished writing the messages, it waits for the server to read them all and return its response. Usually it is like

    rpc ClientStreamCall(stream Req) returns (Resp) {}

This generate a function for the client like

    ClientStreamCall(ctx context.Context, opts ...grpc.CallOption) (MyGrpc_ClientStreamCallClient, error)

which returns an object of a specific client interface `MyGrpc_ClientStreamCallClient`. Two functions called `Send(*Req)` and `CloseAndRecv() (*Resp, error)`, with an object of the `grpc.ClientStream` are defined in the client interface as follows,

    type MyGrpc_ClientStreamCallClient interface {
    	Send(*Req) error
    	CloseAndRecv() (*Resp, error)
    	grpc.ClientStream
    }

Hence the client can use the `Send(*Req)` function to continuously send messages, and use `CloseAndRecv() (*Resp, error)` function to get the response message from the server and close the stream.

For the server side, a function is generated like

    ClientStreamCall(MyGrpc_ClientStreamCallServer) error

with an object of a specific server interface `MyGrpc_ClientStreamCallServer` as a input, which is defined like

    type MyGrpc_ClientStreamCallServer interface {
    	SendAndClose(*Resp) error
    	Recv() (*Req, error)
    	grpc.ServerStream
    }

with two functions called `SendAndClose(*Resp) error` and `Recv() (*Req, error)`, and an object of `grpc.ServerStream`. Hence, the server can use the function `Recv() (*Req, error)` to continuously receive messages from the client, as well as send back message and close the stream with the function `SendAndClose(*Resp) error`.

#### Bidirectional Streaming RPC
In a bidirectional streaming RPC, both sides send a sequence of messages using a read-write stream. The two streams operate independently, so clients and servers can read and write in whatever order they like. Usually it is like

    rpc BiStreamCall(stream Req) returns (stream Resp) {}

This generate a function for the client side like

    BiStreamCall(ctx context.Context, opts ...grpc.CallOption) (MyGrpc_BiStreamCallClient, error)

which returns an object of a specific client interface `MyGrpc_BiStreamCallClient` defined like

    type MyGrpc_BiStreamCallClient interface {
    	Send(*Req) error
    	Recv() (*Resp, error)
    	grpc.ClientStream
    }

with two functions called `Send(*Req) error` and `Recv() (*Resp, error)`, as well as an object of the `grpc.ClientStream`. Then, the client can use the `Send(*Req) error` function and the `Recv() (*Resp, error)` function to continuously send and receive messages to and from the server.

For the server side, a function is generated like

    BiStreamCall(MyGrpc_BiStreamCallServer) error

which receives an object of a specific server interface `MyGrpc_BiStreamCallServer` defined like
    
    type MyGrpc_BiStreamCallServer interface {
    	Send(*Resp) error
    	Recv() (*Req, error)
    	grpc.ServerStream
    }

with two functions called `Send(*Resp) error` and `Recv() (*Req, error)`, as well as an object of the `grpc.ServerStream`. Then, the server can use the `Send(*Resp) error` function and the `Recv() (*Req, error)` function to continuously send and receive messages to and from the client.


## Server Implementations
With the generated codes for the server side shown above, codes of the server providing the defined services can be implemented. 

### Common Steps
The implementations of the server usually include several common steps as follows

- Implementing the generated top-level server interface, e.g. `MyGrpcServer`, by filling the processing logic into the RPC-related functions.
- Using the generated server interface specific to each RPC, e.g. `MyGrpc_ServerStreamCallServer` and `MyGrpc_ClientStreamCallServer`, in each related function to receive and send messages from and to the client.
- Creating a generic gRPC server with options by executing like `grpcServer := grpc.NewServer(opts...)`.
- Registering the implemented server object to the generic gRPC server via the generated function in the `.pb.go` file, e.g. `func RegisterMyGrpcServer(s *grpc.Server, srv MyGrpcServer) {...}`.
- Running a tcp server listening on a port
- Running the generic gRPC server with the opened tcp server.


### More Tips

#### How To End Client-Side Streaming RPC
For the client-side streaming RPC, a simple way to end the RPC and close the stream is like

    for {
        req, err := srv.Recv()
        if err == io.EOF {
            ...
            return srv.SendAndClose(resp)
        }    
        ...
    }

where the type of variables above are like

    req    *Req
    srv    MyGrpc_ClientStreamCallServer
    resp   *Resp

As the outer `for{...}` loop is for continuously receiving messages from the client-side stream, the main idea for ending the RPC is calling the `SendAndClose(*Resp) error` function if a `io.EOF` error occurs when try to receive more messages from the client-side stream. 


#### How To End Bidirectional Streaming RPC
For the bidirectional streaming RPC, a simple way to end the RPC and close the stream is like

    for {
        req, err := srv.Recv()
        if err == io.EOF {
            ...
            return nil
        }    
        ...
    }

where the type of variables are same to the ones in the last section.

As the outer `for{...}` loop is for continuously receiving and sending messages from and to the client, the main idea for ending the RPC is just returning non error if a `io.EOF` error occurs when try to receive more messages from the client. More messages can be send to the client before returning.


## Client Implementations
With the generated codes for the client side shown above, codes of the client consuming the defined services can be implemented.

### Common Steps
The implementations of the client usually include several common steps as follows

- Creating a connection object by calling `func Dial(target string, opts ...DialOption) (*ClientConn, error)` of the `google.golang.org/grpc` package, where the `target` represents an address, with both the host and the port, of a running grpc server.
- Creating a client object for the defined grpc service using generated function in the `.pb.go` file, e.g. `func NewMyGrpcClient(cc *grpc.ClientConn) MyGrpcClient`, with the created connection object in the last step.
- For the **Simple RPC**, directly called the corresponding generated function in the `.pb.go` file, e.g. `func SimpleCall(ctx context.Context, in *Req, opts ...grpc.CallOption) (*Resp, error)`, via the created client object to send a message to the server and get a message containing the response.
- For the **Server-side Streaming RPC**, called the corresponding generated function in the `.pb.go` file, e.g. `func ServerStreamCall(ctx context.Context, in *Req, opts ...grpc.CallOption) (MyGrpc_ServerStreamCallClient, error)`, via the created client object to send a message to the server and get a rpc-type-specific client object, e.g. `MyGrpc_ServerStreamCallClient`, which is used to continuously receive streaming messages from the server side.
- For the **Client-side Streaming RPC**, called the corresponding generated function in the `.pb.go` file, e.g. `func ClientStreamCall(ctx context.Context, opts ...grpc.CallOption) (MyGrpc_ClientStreamCallClient, error)`, via the created client object to get a rpc-type-specific client object, e.g. `MyGrpc_ClientStreamCallClient`, which is used to continuously send streaming messages to the server side and get a message containing the response at the end.
-  For the **Bidirectional Streaming RPC**, called the corresponding generated function in the `.pb.go` file, e.g. `func BiStreamCall(ctx context.Context, opts ...grpc.CallOption) (MyGrpc_BiStreamCallClient, error)`, via the created client object to get a rpc-type-specific client object, e.g. `MyGrpc_BiStreamCallClient`, which is used to continuously send and receive streaming messages to and from the server side at the same time.


### More Tips

#### L4 vs. L7
- Each time creating a client object, e.g. `MyGrpcClient`, a new L4 TCP connection will be established to the server.
- Multiple callings of the different types of rpc via a single client object will lead to transportation with multiple streams on a single L4 TCP connection.


#### Non TLS
When use `grpc.Dial(...)` to establish a connection bearing grpc, securing communication with TLS is enabled by default. To call grpc without TLS, use `grpc.WithInsecure()` as a options in the `grpc.Dial(...)`, e.g. `conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())`.

#### How To Complete Server-Side Streaming RPC
First, call the rpc-specific function ones to send a single message to the server like:

    cli, err:= client.ServerStreamCall(ctx, req)

where the type of variables above are like

    req    *Req
    client MyGrpcClient
    cli    MyGrpc_ServerStreamCallClient

Then, similar to the ways ending client-side streaming RPC at the server side, complete server-side streaming RPC at the client side by

    for {
        resp, err := cli.Recv()
        if err == io.EOF {
            ...
            return
        }    
        ...
    }

where the type of variables above are like

    resp   *Resp
    cli    MyGrpc_ServerStreamCallClient

As the outer `for{...}` loop is for continuously receiving messages from the server-side stream, the main idea for ending the RPC is returning if a `io.EOF` error occurs when try to receive more messages from the server-side stream.


#### Tips for Bidirectional Streaming RPC
The bidirectional streaming RPC is often used in the full-duplex cases, where sending and receiving are taken place in parallel. A simple way to do this at the client side is setting up a goroutine executing receiving/sending or both, and processing receiving/sending in the main process. For example,

    waitch := make(chan struct {})
    go func() {
        for {
            resp, er := cli.Recv()
            if er == io.EOF {
                close(waitch)
                return
            }
            ...
        }
    }()
    for _, req := range reqs {
        er := cli.Send(req)
        ...
    }
    cli.CloseSend()
    <-waitch        
        

where the type of variables above are like

    resp   *Resp
    cli    MyGrpc_BiStreamCallClient
    reqs   []*Req
    req    *Req


The above example realizes concurrently receiving and sending messages at client side, and completes when both the receiving and the sending end. Note that the function `CloseSend()` in `cli` comes from the `grpc.ClientStream` interface as a part of the interface combination forming the interface `MyGrpc_BiStreamCallClient`.


## Wire Format Mapping between GRPC and HTTP/2

See [https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md).


## Other Materials

### HTTP/2
- [https://http2.github.io/](https://http2.github.io/)
- [https://http2.akamai.com/](https://http2.akamai.com/)
- [https://kinsta.com/learn/what-is-http2/](https://kinsta.com/learn/what-is-http2/)

### Protocol Buffers
- [https://developers.google.com/protocol-buffers/docs/overview](https://developers.google.com/protocol-buffers/docs/overview)

### GRPC Gateway
- [https://github.com/grpc-ecosystem/grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)

### GRPC Resource
- [https://github.com/grpc-ecosystem](https://github.com/grpc-ecosystem)

