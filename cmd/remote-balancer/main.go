package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	lbpb "google.golang.org/grpc/grpclb/grpc_lb_v1"

	"github.com/reinventer/grpc-lb-test/proto"
	"github.com/reinventer/grpc-lb-test/remote"
	"github.com/reinventer/grpc-lb-test/server"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclb"
	"google.golang.org/grpc/naming"
)

const (
	SERVERS_NUM = 10
	LB_TOKEN    = `lbtoken`
)

type watcher struct {
	balancers   []*naming.Update
	done, close chan struct{}
}

func (w *watcher) Next() ([]*naming.Update, error) {
	if w.balancers != nil {
		defer func() {
			w.balancers = nil
		}()
		return w.balancers, nil
	}
	<-w.close
	close(w.done)
	return nil, nil
}

func (w *watcher) Close() {
	close(w.close)
	<-w.done
}

type remoteResolver struct {
	w    *watcher
	addr string
}

func (r *remoteResolver) Resolve(target string) (naming.Watcher, error) {
	r.w.close = make(chan struct{})
	r.w.done = make(chan struct{})
	return r.w, nil
}

func main() {
	var (
		startWg, stopWg sync.WaitGroup
		servers         []*grpc.Server
		mtx             sync.Mutex
		lbstart         = make(chan struct{})
	)
	addrBal := fmt.Sprintf(`localhost:%d`, 15070)
	balancer, err := remote.New(LB_TOKEN, `/v2/keys/balancer`, `http://127.0.0.1:2379`)
	if err != nil {
		log.Fatal(err)
	}
	lbsrv := grpc.NewServer()
	stopWg.Add(1)
	go func() {
		lisBal, err := net.Listen(`tcp`, addrBal)
		if err != nil {
			panic(err)
		}
		defer lisBal.Close()

		lbpb.RegisterLoadBalancerServer(lbsrv, balancer)

		close(lbstart)

		lbsrv.Serve(lisBal)
		stopWg.Done()
	}()
	<-lbstart

	servers = append(servers, lbsrv)

	w := &watcher{
		balancers: []*naming.Update{
			{
				Op:   naming.Add,
				Addr: addrBal,
				Metadata: &grpclb.Metadata{
					AddrType: grpclb.GRPCLB,
				},
			},
		},
	}

	res := &remoteResolver{
		w: w,
	}

	b := grpclb.Balancer(res)

	startWg.Add(SERVERS_NUM)
	stopWg.Add(SERVERS_NUM)
	for i := 0; i < SERVERS_NUM; i++ {
		go func(i int) {
			addr := fmt.Sprintf(`localhost:%d`, 15080+i)
			lis, err := net.Listen(`tcp`, addr)
			if err != nil {
				panic(err)
			}
			defer lis.Close()

			grpcSrv := grpc.NewServer()
			proto.RegisterLBTestServer(grpcSrv, server.New(fmt.Sprintf(`test_%d`, i)))
			mtx.Lock()
			servers = append(servers, grpcSrv)
			mtx.Unlock()
			startWg.Done()
			// TODO: set in etcd /v2/keys/balancer/service/[uuid] => localhost:[addr]

			grpcSrv.Serve(lis)
			stopWg.Done()
			// TODO: remove from etcd /v2/keys/balancer/service/[uuid]
		}(i)
	}

	startWg.Wait()
	log.Printf(`started %d servers`, SERVERS_NUM)

	conn, err := grpc.Dial(`service`, grpc.WithBalancer(b), grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	cli := proto.NewLBTestClient(conn)

	for {
		id, err := cli.Who(context.Background(), &proto.Empty{})
		if err != nil {
			panic(err)
		}
		log.Printf(`who: %s`, id.ID)
		time.Sleep(time.Second)
	}

	balancer.Close()
	conn.Close()

	for _, s := range servers {
		s.GracefulStop()
	}

	stopWg.Wait()
	log.Print(`servers stopped`)
}
