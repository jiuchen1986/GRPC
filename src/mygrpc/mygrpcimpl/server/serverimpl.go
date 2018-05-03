// Implementations of the server of mygrpc

package server

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "os"
    
    pb "mygrpc/mygrpc"
    
    "golang.org/x/net/context"
)

var MyGrpcLogger = log.New(os.Stderr, "mygrpc_server_", log.LstdFlags|log.Lshortfile)

type myGrpcServer struct {
    useTestFile bool   // whether use the testing data from a json file
    testSvcInfo map[string]*pb.ServiceDescriptor  // map of service descriptors read from the testing file
    svcName     string   // name of the service providing by the server
}

// error type used to raise exceptions when dealing with service info
type ServiceError struct {
    SvcName string
    ChainId int32
    Msg     string
    Err     error    
}

func (e *ServiceError) Error() string {
    if e.Err == nil {
        return fmt.Sprintf("Error for service %s in chain %d: %s", e.SvcName, e.ChainId, e.Msg)
    }
    
    return fmt.Sprintf("Error for service %s in chain %d: %s", e.SvcName, e.ChainId, e.Err.Error())
}

func NewMyGrpcServer(useTest bool, testData []*pb.ServiceDescriptor, name string) *myGrpcServer {
    if useTest {
        jsonStr, _ := json.Marshal(testData)
        MyGrpcLogger.Printf("Generate mygrpc server with testing svc info: \n%s", string(jsonStr))
        svcMap := make(map[string]*pb.ServiceDescriptor)
        for _, svc := range testData {
            svcMap[svc.GetSvcName()] = svc
        }
        return &myGrpcServer{useTestFile: useTest, testSvcInfo: svcMap, svcName: name}
    }
    
    return nil
}

// return a descriptor of a service chain
func (s *myGrpcServer) GetServiceChainDescriptor(sc *pb.ServiceChain) (*pb.ServiceChainDescriptor, error) {
    if s.useTestFile {
        cd := make([]*pb.ServiceDescriptor, sc.GetChainLen())
        for _, svc := range sc.GetChain() {
            sd, prs := s.testSvcInfo[svc.GetSvcName()]
            if !prs {
                return nil, &ServiceError{
                                SvcName: svc.GetSvcName(),
                                ChainId: sc.GetChainId(),
                                Msg:     "No service found",
                                Err:     nil,
                            }
            }
            if svc.GetSvcPos() > sc.GetChainLen() || svc.GetSvcPos() < 1 {
                return nil, &ServiceError{
                                SvcName: svc.GetSvcName(),
                                ChainId: sc.GetChainId(),
                                Msg:     fmt.Sprintf("Wrong service position %d with chain len %d", svc.GetSvcPos(), sc.GetChainLen()),
                                Err:     nil,
                            }
            }
            sd.SvcPos = svc.GetSvcPos()
            cd[sd.SvcPos - 1] = sd
        }
        scd := &pb.ServiceChainDescriptor{
                   ChainId:    sc.GetChainId(),
                   ChainLen:   sc.GetChainLen(),
                   ChainDesc:  cd,
               }
        return scd, nil
    }
    
    return nil, nil
}

func (s *myGrpcServer) GetChainReqResp(ctx context.Context, sc *pb.ServiceChain) (*pb.ServiceChainDescriptor, error) {
    jsonStr, _ := json.Marshal(sc)
    MyGrpcLogger.Printf("Received service chain request: \n%s", string(jsonStr))
    return s.GetServiceChainDescriptor(sc)
}

func (s *myGrpcServer) GetChainsReqResps(scs *pb.ServiceChains, srv pb.MyGrpc_GetChainsReqRespsServer) error {
    jsonStr, _ := json.Marshal(scs)
    MyGrpcLogger.Printf("Received service chains request: \n%s", string(jsonStr))
    var scd *pb.ServiceChainDescriptor
    var err error
    for _, sc := range scs.GetChains() {
        scd, err = s.GetServiceChainDescriptor(sc)
        if err != nil {
            return err
        }
        
        err = srv.Send(scd)
        if err != nil {
            return err
        }
    }
    
    return nil
}

func (s *myGrpcServer) GetChainsReqsResp(srv pb.MyGrpc_GetChainsReqsRespServer) error {
    scds := &pb.ServiceChainDescriptors{ChainDescs: make([]*pb.ServiceChainDescriptor, 1)}
    
    // continuously receiving messages from clients
    var scd *pb.ServiceChainDescriptor
    var err error
    for {
        sc, e := srv.Recv()        
        if e == io.EOF {
            scds.ChainDescs = scds.ChainDescs[1:]  //drop the first empty element
            return srv.SendAndClose(scds)
        }
        
        if e != nil {
            return e
        }
        
        jsonStr, _ := json.Marshal(sc)
        MyGrpcLogger.Printf("Received service chain request: \n%s", string(jsonStr))
        
        scd, err = s.GetServiceChainDescriptor(sc);
        if err != nil {
            return err
        }
        
        scds.ChainDescs = append(scds.ChainDescs, scd)
    }
}

func (s *myGrpcServer) GetChainReqsResps(srv pb.MyGrpc_GetChainReqsRespsServer) error {
    var scd *pb.ServiceChainDescriptor
    var err error
    for {
        sc, e := srv.Recv()        
        if e == io.EOF {
            return nil
        }
        
        if e != nil {
            return err
        }
        
        jsonStr, _ := json.Marshal(sc)
        MyGrpcLogger.Printf("Received service chain request: \n%s", string(jsonStr))
        
        scd, err = s.GetServiceChainDescriptor(sc);
        if err != nil {
            return err
        }
        
        err = srv.Send(scd)
        if err != nil {
            return err
        }
    }
}