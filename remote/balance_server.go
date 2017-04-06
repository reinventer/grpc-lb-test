package remote

import (
	//"time"

	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"

	lbpb "google.golang.org/grpc/grpclb/grpc_lb_v1"
)

type addressUpdates chan *lbpb.ServerList

type balancer struct {
	servers map[string]addressUpdates
}

func New(prefix string, etcdAddrs ...string) {
	
}

func (b *balancer) BalanceLoad(stream lbpb.LoadBalancer_BalanceLoadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	initReq := req.GetInitialRequest()
	resp := &lbpb.LoadBalanceResponse{
		LoadBalanceResponseType: &lbpb.LoadBalanceResponse_InitialResponse{
			InitialResponse: new(lbpb.InitialLoadBalanceResponse),
		},
	}
	if err := stream.Send(resp); err != nil {
		return err
	}
	for _, a := range b.servers[initReq.Name] {
		resp = &lbpb.LoadBalanceResponse{
			LoadBalanceResponseType: &lbpb.LoadBalanceResponse_ServerList{
				ServerList: a,
			},
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (b *balancer) Close() {
	for _, c := range b.servers {
		close(c)
	}
}
