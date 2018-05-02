// Entry of the server of mygrpc

package mygrpcserver

import (
    "encoding/json"
    "flag"
    "log"
    "os"
    "fmt"
    "io/ioutil"
    "net"    
    
    pb "mygrpc/mygrpc"
    
    "google.golang.org/grpc"
)

var (
    // tls              = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
    useTestFile      = flag.Bool("test_file", true, "Uses the json file containing service info as the data source")
    svcInfoFile      = flag.String("svc_info_file", "/usr/src/grpc/src/mygrpc/test_data.json", "A json file containing service info for testing")
    svcName          = flag.String("name", "svcA", "The name of the service providing by this server")
    port             = flag.Int("port", 8082, "The server port")
    
    MyGrpcLogger     = log.New(os.Stderr, "mygrpc_", log.LstdFlags|log.Lshortfile)
)

func main() {
    
    flag.Parse()
    if *useTestFile {
    
        MyGrpcLogger.Printf("Use service info in the json file at %s", *svcInfoFile)
        file_data, err := ioutil.ReadFile(*svcInfoFile)
        if err != nil {
            MyGrpcLogger.Fatalf("Failed to load service info from %s: %v", *svcInfoFile, err) 
        }
        var svc_info []*pb.ServiceDescriptor
        if err := json.Unmarshal(file_data, &svc_info); err != nil {
            MyGrpcLogger.Fatalf("Failed to unmarshal the service info: %v", err)
        }
        MyGrpcLogger.Printf("Successfully load service info from the json file at %s", *svcInfoFile)
        
        gprcServer = grpc.NewServer()
        pb.RegisterMyGrpcServer(grpcServer, NewMyGrpcServer(*useTestFile, svc_info, *svc_name))
        
        lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
        if err != nil {
            MyGrpcLogger.Fatalf("Failed to listen: %v", err)
        }
        
        MyGrpcLogger.Printf("Starting MyGrpc grpc server at port %d", *port)
        grpcServer.Serve(lis)
    
    }
    
    MyGrpcLogger.Fatal("Current implementation only support loading service info from json file!")
    
}