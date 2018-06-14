// Validation for mygrpc

package validate

func ValRpcType(t string) bool {
    rpc_type := [4]string{"simple", "server_stream", "client_stream", "bi_stream"}
    for _, st := range rpc_type {
        if t == st {
            return true
        }
    }
    return false
}