package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"

	"github.com/reinventer/grpc-lb-test/proto"
	"github.com/reinventer/grpc-lb-test/resolver"
	"github.com/reinventer/grpc-lb-test/server"
	"golang.org/x/net/context"
	"time"
)

const (
	SERVERS_NUM = 10
)

func main() {
	var (
		startWg, stopWg sync.WaitGroup
		servers         []*grpc.Server
		mtx             sync.Mutex
	)

	res, err := resolver.New(`/v2/keys/balancer`, `http://127.0.0.1:2379`)
	if err != nil {
		log.Fatal(err)
	}
	b := grpc.RoundRobin(res)

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

	conn, err := grpc.Dial(`service`, grpc.WithBalancer(b), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	cli := proto.NewLBTestClient(conn)

	time.Sleep(time.Second)
	for {
		id, err := cli.Who(context.Background(), &proto.Empty{})
		if err != nil {
			panic(err)
		}
		log.Printf(`who: %s`, id.ID)
		time.Sleep(time.Second)
	}

	for _, s := range servers {
		s.GracefulStop()
	}

	stopWg.Wait()
	log.Print(`servers stopped`)
}
