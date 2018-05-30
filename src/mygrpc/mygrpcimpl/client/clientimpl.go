// Implementations of the client of mygrpc

package client

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "io"
    "time"
    
    pb "mygrpc/mygrpc"
    
    "golang.org/x/net/context"
    "google.golang.org/grpc"
)

var myGrpcLogger = log.New(os.Stderr, "mygrpc_client_", log.LstdFlags|log.Lshortfile)

type myGrpcClientSet struct {
    useTestFile     bool   // whether use the testing data from a json file
    testChainInfo   []*pb.ServiceChain  // list of service chains read from the testing file
    serverAddr      string  // address of the GRPC server, 'ip:port'
    rpcType         string  // type of the RPC will be called
    callInterval    time.Duration   // interval between each call in seconds
    callTimeout     time.Duration   // maximal time in seconds waiting for one RPC completing
    callNum         int   // number of time calling the RPC per goroutine
    concurNum       int   // number of goroutine concurrently calling the RPC per client
    clientNum       int   // number of client instance running. Each client instance owns an independent tcp connection
}

func NewMyGrpcClientSet(utf bool, tci []*pb.ServiceChain, sa, rt string, ci, ct time.Duration, cn, ccn, clin int) *myGrpcClientSet {
    return &myGrpcClientSet{
               useTestFile: utf,
               testChainInfo: tci,
               serverAddr: sa,
               rpcType: rt,
               callInterval: ci,
               callTimeout:  ct,
               callNum: cn,
               concurNum: ccn,
               clientNum: clin,
           }
}

type routineChannel struct {
    clientId         int  // id of the related client istance
    routineId        int  // id of the related goroutine
    endChannel        chan struct{}  // a channel used to notify the completion of a goroutine
    timeoutChannel   chan bool  // a channel used to end the goroutine when timeout
}

func (r *routineChannel) Wait(ci, ct time.Duration, cn int) error {
    go func() {
        time.Sleep(time.Duration((int(ci) + int(ct)) * cn + int(ct)))
        r.timeoutChannel <- true
    }()
    
    select {
        case <-r.endChannel:
            myGrpcLogger.Printf("Processing of goroutine %d in client %d completes", r.routineId, r.clientId)
        case <-r.timeoutChannel:
            myGrpcLogger.Printf("Processing of goroutine %d in client %d timeouts", r.routineId, r.clientId)
    }
    
    return nil
}

// error type used to raise exceptions when dealing with service chain info
type ChainError struct {
    ChainId int32
    Msg     string
    Err     error
}

func (e *ChainError) Error() string {
    if e.Err == nil {
        return fmt.Sprintf("Error for service chain %d: %s", e.ChainId, e.Msg)
    }
    
    return fmt.Sprintf("Error for service chain %d: %s", e.ChainId, e.Err.Error())
}

// running the all intances of client
func (c *myGrpcClientSet) Run() error {
    for i := 0; i < c.clientNum; i++ {
        conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure())
        if err != nil {
            myGrpcLogger.Fatalf("fail to dial: %v", err)
        }
        defer conn.Close()
        client := pb.NewMyGrpcClient(conn)
        
        for j := 0; j < c.concurNum; j++ {
            cr := &routineChannel{
                      clientId:       i,
                      routineId:      j,
                      endChannel:     make(chan struct{}),
                      timeoutChannel: make(chan bool),
                  }
            // Call rpc
            switch c.rpcType {
                case "simple":
                    go c.CallSimpleRPC(client, cr.endChannel, cr.routineId, cr.clientId)
                case "server_stream":
                    go c.CallServerStreamRPC(client, cr.endChannel, cr.routineId, cr.clientId)
                case "client_stream":
                    go c.CallClientStreamRPC(client, cr.endChannel, cr.routineId, cr.clientId)
                case "bi_stream":
                    go c.CallBiStreamRPC(client, cr.endChannel, cr.routineId, cr.clientId)
            }
            cr.Wait(c.callInterval, c.callTimeout, c.callNum)
        }
    }
    
    return nil
}

func (c *myGrpcClientSet) CallSimpleRPC(client pb.MyGrpcClient, ech chan struct{}, rid, cid int) error {    
    if c.useTestFile {
        myGrpcLogger.Printf("Calling simple rpc by goroutine %d in client %d", rid, cid)
        cfs := make([]context.CancelFunc, c.callNum)  // cancel function list for the context in each calling
        ctxs := make([]context.Context, c.callNum)  // context list for each calling
        for k := 0; k < c.callNum; k++ {
            ctxs[k], cfs[k] = context.WithTimeout(context.Background(), c.callTimeout + time.Duration(k * int(c.callInterval)))
        }            
        defer release(ech, cfs)
        
        for k := 0; k < c.callNum; k++ {
            scd, err := client.GetChainReqResp(ctxs[k], c.testChainInfo[k%len(c.testChainInfo)])
            if err != nil {
                myGrpcLogger.Fatalf("Failed to call simple rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, err)
            }
            jsonStr, _ := json.Marshal(scd)
            myGrpcLogger.Printf("Get a response from simple rpc by goroutine %d in client %d for call number %d: %s", rid, cid, k, string(jsonStr))
            
            time.Sleep(c.callInterval)
        }
        
        return nil        
    }
    
    myGrpcLogger.Printf("Current implementations only support using test data, requests are pass")
    close(ech)
    return nil
}

func (c *myGrpcClientSet) CallServerStreamRPC(client pb.MyGrpcClient, ech chan struct{}, rid, cid int) error {    
    if c.useTestFile {
        myGrpcLogger.Printf("Calling server-streaming rpc by goroutine %d in client %d", rid, cid)
        scl := make([]*pb.ServiceChain, len(c.testChainInfo))
        for l, sc := range c.testChainInfo {
            scl[l] = sc
        }
        scs := &pb.ServiceChains {
                   Chains: scl,
               }
        
        cfs := make([]context.CancelFunc, c.callNum)  // cancel function list for the context in each calling
        ctxs := make([]context.Context, c.callNum)  // context list for each calling
        for k := 0; k < c.callNum; k++ {
            ctxs[k], cfs[k] = context.WithTimeout(context.Background(), c.callTimeout + time.Duration(k * int(c.callInterval)))
        }            
        defer release(ech, cfs)
        
        var stream pb.MyGrpc_GetChainsReqRespsClient
        var err error
        for k := 0; k < c.callNum; k++ {
            
            stream, err = client.GetChainsReqResps(ctxs[k], scs)
            if err != nil {
                myGrpcLogger.Fatalf("Failed to call server-streaming rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, err)
            }
            for {
                scd, er := stream.Recv()
                if er == io.EOF {
                    break
                }
                if er != nil {
                    myGrpcLogger.Fatalf("Failed to receive from server-streaming rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, er)
                }
                jsonStr, _ := json.Marshal(scd)
                myGrpcLogger.Printf("Get a response from server-streaming rpc by goroutine %d in client %d for call number %d: %s", rid, cid, k, string(jsonStr))
            }

            time.Sleep(c.callInterval)
        }
        
        return nil        
    }
    
    myGrpcLogger.Printf("Current implementations only support using test data, requests are pass")
    close(ech)
    return nil
}

func (c *myGrpcClientSet) CallClientStreamRPC(client pb.MyGrpcClient, ech chan struct{}, rid, cid int) error {    
    if c.useTestFile {
        myGrpcLogger.Printf("Calling client-streaming rpc by goroutine %d in client %d", rid, cid)
        cfs := make([]context.CancelFunc, c.callNum)  // cancel function list for the context in each calling
        ctxs := make([]context.Context, c.callNum)  // context list for each calling
        for k := 0; k < c.callNum; k++ {
            ctxs[k], cfs[k] = context.WithTimeout(context.Background(), c.callTimeout + time.Duration(k * int(c.callInterval)))
        }            
        defer release(ech, cfs)
        
        var stream pb.MyGrpc_GetChainsReqsRespClient
        var err error        
        for k := 0; k < c.callNum; k++ {
            
            stream, err = client.GetChainsReqsResp(ctxs[k])
            if err != nil {
                myGrpcLogger.Fatalf("Failed to call client-streaming rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, err)
            }
            for _, sc := range c.testChainInfo {
                er := stream.Send(sc)
                if er != nil {
                    myGrpcLogger.Fatalf("Failed to send to client-streaming rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, er)
                }
                jsonStr, _ := json.Marshal(sc)
                myGrpcLogger.Printf("Send to client-streaming rpc by goroutine %d in client %d for call number %d: %s", rid, cid, k, string(jsonStr))
            }
            
            scds, er := stream.CloseAndRecv()
            if er != nil {
                myGrpcLogger.Fatalf("Failed to receive from client-streaming rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, er)
            }
            jsonStr, _ := json.Marshal(scds)
            myGrpcLogger.Printf("Get a response from client-streaming rpc by goroutine %d in client %d for call number %d: %s", rid, cid, k, string(jsonStr))
            
            time.Sleep(c.callInterval)
        }
        
        return nil        
    }
    
    myGrpcLogger.Printf("Current implementations only support using test data, requests are pass")
    close(ech)
    return nil
}

func (c *myGrpcClientSet) CallBiStreamRPC(client pb.MyGrpcClient, ech chan struct{}, rid, cid int) error {    
    if c.useTestFile {
        myGrpcLogger.Printf("Calling bi-streaming rpc by goroutine %d in client %d", rid, cid)
        cfs := make([]context.CancelFunc, c.callNum)  // cancel function list for the context in each calling
        ctxs := make([]context.Context, c.callNum)  // context list for each calling
        for k := 0; k < c.callNum; k++ {
            ctxs[k], cfs[k] = context.WithTimeout(context.Background(), c.callTimeout + time.Duration(k * int(c.callInterval)))
        }            
        defer release(ech, cfs)
        
        var stream pb.MyGrpc_GetChainsReqsRespsClient
        var err error        
        for k := 0; k < c.callNum; k++ {
            
            stream, err = client.GetChainsReqsResps(ctxs[k])
            if err != nil {
                myGrpcLogger.Fatalf("Failed to call bi-streaming rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, err)
            }
            waitch := make(chan struct {})
            go func() {
                for {
                    scd, er := stream.Recv()
                    if er == io.EOF {
                        close(waitch)
                        return
                    }
                    if er != nil {
                        myGrpcLogger.Fatalf("Failed to receive from server-streaming rpc by co-goroutine of goroutine %d in client %d for call number %d: %v", rid, cid, k, er)
                    }
                    jsonStr, _ := json.Marshal(scd)
                    myGrpcLogger.Printf("Get a response from server-streaming rpc by co-goroutine of goroutine %d in client %d for call number %d: %s", rid, cid, k, string(jsonStr))
                }
            }()
            
            for _, sc := range c.testChainInfo {
                er := stream.Send(sc)
                if er != nil {
                    myGrpcLogger.Fatalf("Failed to send to bi-streaming rpc by goroutine %d in client %d for call number %d: %v", rid, cid, k, er)
                }
                jsonStr, _ := json.Marshal(sc)
                myGrpcLogger.Printf("Send to client-streaming rpc by goroutine %d in client %d for call number %d: %s", rid, cid, k, string(jsonStr))
                time.Sleep(1e6)
            }
            stream.CloseSend()  // function of the grpc.ClientStream interface, which is the part of the interface combination composing pb.MyGrpc_GetChainsReqsRespsClient interface
            <-waitch
            
            time.Sleep(c.callInterval)
        }
        
        return nil        
    }
    
    myGrpcLogger.Printf("Current implementations only support using test data, requests are pass")
    close(ech)
    return nil
}

// function used to ensure resource are released when calling rpcs in a goroutine
// by closing ending channel and canceling all contexts
func release(ech chan struct{}, cfs []context.CancelFunc) {
    close(ech)
    for _, cf := range cfs {
        cf()
    }
}