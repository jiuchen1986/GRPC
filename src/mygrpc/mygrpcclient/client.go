// Entry of the client of mygrpc

package main

import (
    "encoding/json"
    "flag"
    "log"
    "os"
    "io/ioutil"
    "time"    
    
    pb "mygrpc/mygrpc"
    impl "mygrpc/mygrpcimpl/client"
    val  "mygrpc/util/validate"
)

var (
    // tls              = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
    useTestFile         = flag.Bool("test_file", true, "Uses the json file containing service chain info as the data source")
    chainInfoFile       = flag.String("chain_info_file", "/usr/src/grpc/src/mygrpc/testdata/test_data_client.json", "A json file containing service chain info for testing")
    serverAddr          = flag.String("server", "localhost:8082", "The address of the mygrpc server")
    rpcType             = flag.String("rpc", "simple", "The type of RPC will be called, including 'simple', 'server_stream', 'client_stream', 'bi_stream'")
    callInterval        = flag.Int("inv", 1, "The interval between each call in seconds")
    callTimeout         = flag.Int("time", 1, "The maximal time in seconds waiting for one RPC completing")
    callNum             = flag.Int("num", 10, "The number of time calling the RPC per goroutine")
    concurNum           = flag.Int("con", 1, "The number of goroutine concurrently calling the RPC per client")
    clientNum           = flag.Int("cli", 1, "The number of client instance running. Each client instance owns an independent tcp connection")
    
    myGrpcLogger        = log.New(os.Stderr, "mygrpc_client_", log.LstdFlags|log.Lshortfile)
)

func main() {
    
    flag.Parse()
    if !val.ValRpcType(*rpcType) {
        myGrpcLogger.Fatalf("Invalid rpc type: %s", *rpcType)
    }
    if *useTestFile {
    
        myGrpcLogger.Printf("Use service chain info in the json file at %s", *chainInfoFile)
        file_data, err := ioutil.ReadFile(*chainInfoFile)
        if err != nil {
            myGrpcLogger.Fatalf("Failed to load service chain info from %s: %v", *chainInfoFile, err)
        }
        var chain_info []*pb.ServiceChain
        if err := json.Unmarshal(file_data, &chain_info); err != nil {
            myGrpcLogger.Fatalf("Failed to unmarshal the service chain info: %v", err)
        }
        myGrpcLogger.Printf("Successfully load service chain info from the json file at %s", *chainInfoFile)
        
        
        clientSet := impl.NewMyGrpcClientSet(*useTestFile,
                                             chain_info,
                                             *serverAddr,
                                             *rpcType,
                                             time.Duration(*callInterval * int(time.Second)),
                                             time.Duration(*callTimeout * int(time.Second)),
                                             *callNum,
                                             *concurNum,
                                             *clientNum)

        myGrpcLogger.Printf("Running MyGrpc grpc client")
        if err := clientSet.Run(); err != nil {
            myGrpcLogger.Fatalf("Failed running MyGrpc grpc client: %v", err)
        }
        
        return
    
    }
    
    myGrpcLogger.Fatal("Current implementation only support loading service chain info from json file!")
    
}