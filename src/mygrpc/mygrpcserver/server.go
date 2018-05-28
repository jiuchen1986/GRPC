// Entry of the server of mygrpc

package main

import (
    "encoding/json"
    "flag"
    "log"
    "os"
    "fmt"
    "io/ioutil"
    "net"    
    
    pb "mygrpc/mygrpc"
    impl "mygrpc/mygrpcimpl/server"
    
    "google.golang.org/grpc"
)

var (
    // tls              = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
    useTestFile      = flag.Bool("test_file", true, "Uses the json file containing service info as the data source")
    svcInfoFile      = flag.String("svc_info_file", "/usr/src/grpc/src/mygrpc/testdata/test_data_server.json", "A json file containing service info for testing")
    svcName          = flag.String("name", "svcA", "The name of the service providing by this server")
    port             = flag.Int("port", 8082, "The server port")
    
    myGrpcLogger     = log.New(os.Stderr, "mygrpc_server_", log.LstdFlags|log.Lshortfile)
)

func main() {
    
    flag.Parse()
    if *useTestFile {
    
        myGrpcLogger.Printf("Use service info in the json file at %s", *svcInfoFile)
        file_data, err := ioutil.ReadFile(*svcInfoFile)
        if err != nil {
            myGrpcLogger.Fatalf("Failed to load service info from %s: %v", *svcInfoFile, err) 
        }
        var svc_info []*pb.ServiceDescriptor
        if err := json.Unmarshal(file_data, &svc_info); err != nil {
            myGrpcLogger.Fatalf("Failed to unmarshal the service info: %v", err)
        }
        myGrpcLogger.Printf("Successfully load service info from the json file at %s", *svcInfoFile)
        
        grpcServer := grpc.NewServer()
        pb.RegisterMyGrpcServer(grpcServer, impl.NewMyGrpcServer(*useTestFile, svc_info, *svcName))
        
        lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
        if err != nil {
            myGrpcLogger.Fatalf("Failed to listen: %v", err)
        }
        
        myGrpcLogger.Printf("Starting MyGrpc grpc server at port %d", *port)
        grpcServer.Serve(lis)
    
    }
    
    myGrpcLogger.Fatal("Current implementation only support loading service info from json file!")
    
}