package remote

import (
	"errors"
	"sync"

	"bytes"
	"fmt"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	lbpb "google.golang.org/grpc/grpclb/grpc_lb_v1"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"net"
	"strconv"
	"strings"
)

type balancer struct {
	sync.Mutex
	token   string
	ctx     context.Context
	cancel  context.CancelFunc
	api     client.KeysAPI
	wg      sync.WaitGroup
	stopped bool
}

func New(token, prefix string, etcdAddrs ...string) (*balancer, error) {
	cfg := client.Config{
		Endpoints: etcdAddrs,
		Transport: client.DefaultTransport,
	}

	cli, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	b := &balancer{
		token: token,
		api:   client.NewKeysAPIWithPrefix(cli, prefix),
	}

	b.ctx, b.cancel = context.WithCancel(context.Background())

	return b, nil
}

func (b *balancer) BalanceLoad(stream lbpb.LoadBalancer_BalanceLoadServer) error {
	b.Lock()
	if b.stopped {
		b.Unlock()
		return status.Error(codes.Aborted, `balancer stopped`)
	}
	b.Unlock()
	b.wg.Add(1)

	defer func() {
		b.wg.Done()
	}()

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
	updates, err := b.updates(initReq.Name)
	if err != nil {
		return err
	}

	for a := range updates {
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
	b.Lock()
	b.stopped = true
	b.Unlock()

	b.cancel()
	b.wg.Wait()
}

func (b *balancer) updates(target string) (<-chan *lbpb.ServerList, error) {
	var (
		watcher = b.api.Watcher(`/`+target, &client.WatcherOptions{Recursive: true})
		ch      = make(chan *lbpb.ServerList)
		list    = &lbpb.ServerList{}
	)

	res, err := b.api.Get(b.ctx, `/`+target, &client.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	if !res.Node.Dir {
		return nil, errors.New(`wrong node`)
	}

	list.Servers = []*lbpb.Server{}
	for _, n := range res.Node.Nodes {
		ip, port, err := getAddress(n.Value)
		if err != nil {
			grpclog.Printf(`wrong address, key: %s, value: %s, error: %s`, n.Key, n.Value, err)
			continue
		}
		list.Servers = append(list.Servers, &lbpb.Server{
			IpAddress:        ip,
			Port:             port,
			LoadBalanceToken: b.token,
		})
	}

	go func() {
		defer close(ch)
		ch <- list

	WATCH:
		for {
			res, err = watcher.Next(b.ctx)
			if err != nil {
				grpclog.Print(err)
				break
			}

			switch res.Action {
			case `set`, `update`, `create`:
				ip, port, err := getAddress(res.Node.Value)
				if err != nil {
					grpclog.Printf(`wrong address, key: %s, value: %s, error: %s`, res.Node.Key, res.Node.Value, err)
				}
				for _, s := range list.Servers {
					if bytes.Compare(s.IpAddress, ip) == 0 && s.Port == port {
						//already added
						continue WATCH
					}
				}
				list.Servers = append(list.Servers, &lbpb.Server{
					IpAddress:        ip,
					Port:             port,
					LoadBalanceToken: b.token,
				})
			case `delete`:
				ip, port, err := getAddress(res.PrevNode.Value)
				if err != nil {
					grpclog.Printf(`wrong address, key: %s, value: %s, error: %s`, res.PrevNode.Key, res.PrevNode.Value, err)
				}
				for i, s := range list.Servers {
					if bytes.Compare(s.IpAddress, ip) == 0 && s.Port == port {
						list.Servers = append(list.Servers[:i], list.Servers[i+1:]...)
						break
					}
				}
			}
			ch <- list
		}
	}()

	return ch, nil
}

func getAddress(addr string) ([]byte, int32, error) {
	parts := strings.Split(addr, `:`)
	if len(parts) != 2 {
		return nil, 0, errors.New(`wrong ip:port`)
	}
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return nil, 0, errors.New(`wrong ip address`)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf(`wrong port: %s`, err)
	}
	if port < 0 {
		return nil, 0, fmt.Errorf(`port should be positive`, err)
	}

	return []byte(ip), int32(port), nil
}
