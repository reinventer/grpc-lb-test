package balancer

import (
	"sync"

	"errors"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type balAddr struct {
	key       string
	addr      grpc.Address
	connected bool
}

type bal struct {
	sync.Mutex

	etcdAddrs []string
	addresses []balAddr
	i         int

	notify   chan []grpc.Address
	stopNext context.CancelFunc
}

func New(etcdAddrs ...string) *bal {
	return &bal{
		etcdAddrs: etcdAddrs,
	}
}

func (b *bal) Start(target string, config grpc.BalancerConfig) error {
	cfg := client.Config{
		Endpoints: b.etcdAddrs,
		Transport: client.DefaultTransport,
	}

	c, err := client.New(cfg)
	if err != nil {
		return err
	}

	api := client.NewKeysAPI(c)
	watcher := api.Watcher(target, &client.WatcherOptions{Recursive: true})

	ctx, cancel := context.WithCancel(context.Background())
	b.stopNext = cancel

	res, err := api.Get(ctx, target, &client.GetOptions{Recursive: true})
	if err != nil {
		return err
	}

	if res.Node.Dir {
		b.setAddr(res.Node.Nodes...)
	}

	b.notify = make(chan []grpc.Address)

	go func() {
		select {
		case <-ctx.Done():
			return
		case b.notify <- b.addrs():
		}

		for {
			res, err = watcher.Next(ctx)
			if err != nil {
				panic(err)
			}

			switch res.Action {
			case `set`, `update`, `create`:
				b.setAddr(res.Node)
			case `delete`:
				b.delAddr(res.Node)
			}

			select {
			case <-ctx.Done():
				return
			case b.notify <- b.addrs():
			}
		}
	}()

	return nil
}

func (b *bal) Up(addr grpc.Address) (down func(error)) {
	b.Lock()
	defer b.Unlock()

	for _, a := range b.addresses {
		if a.addr == addr {
			a.connected = true
			break
		}
	}
	return func(err error) {
		b.Lock()
		for _, a := range b.addresses {
			if a.addr == addr {
				a.connected = false
				break
			}
		}
		b.Unlock()
	}
}

func (b *bal) Get(ctx context.Context, opts grpc.BalancerGetOptions) (grpc.Address, func(), error) {
	b.Lock()
	defer b.Unlock()

	if b.i >= len(b.addresses) {
		b.i = 0
	}

	var (
		addr balAddr
		i    = b.i
	)

	for {
		if i >= len(b.addresses) {
			i = 0
		}
		if addr = b.addresses[i]; !addr.connected {
			b.i = i + 1
			break
		}
		i += 1
		if i == b.i {
			// TODO: wait if opts.BlockingWait
			return addr.addr, nil, errors.New(`address not found`)
		}
	}

	return addr.addr, nil, nil
}

func (b *bal) Notify() <-chan []grpc.Address {
	return b.notify
}

func (b *bal) Close() error {
	b.stopNext()
	if b.notify != nil {
		close(b.notify)
	}
	return nil
}

func (b *bal) addrs() []grpc.Address {
	b.Lock()
	defer b.Unlock()
	addrs := []grpc.Address{}
	for _, a := range b.addresses {
		addrs = append(addrs, a.addr)
	}
	return addrs
}

func (b *bal) setAddr(nodes ...*client.Node) {
	b.Lock()
	for _, n := range nodes {
		var find bool
		for _, a := range b.addresses {
			if a.key == n.Key {
				a.addr = grpc.Address{Addr: n.Value}
				find = true
				break
			}
		}

		if !find {
			b.addresses = append(b.addresses, balAddr{
				key:  n.Key,
				addr: grpc.Address{Addr: n.Value},
			})
		}
	}
	b.Unlock()
}

func (b *bal) delAddr(nodes ...*client.Node) {
	b.Lock()
	for _, n := range nodes {
		for i, a := range b.addresses {
			if a.key == n.Key {
				b.addresses = append(b.addresses[:i], b.addresses[i+1:]...)
				break
			}
		}
	}
	b.Unlock()
}
