package resolver

import (
	"errors"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
	"sync"
)

type res struct {
	api client.KeysAPI
}

func New(prefix string, etcdAddrs ...string) (*res, error) {
	cfg := client.Config{
		Endpoints: etcdAddrs,
		Transport: client.DefaultTransport,
	}

	cli, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	api := client.NewKeysAPIWithPrefix(cli, prefix)

	return &res{api: api}, nil
}

func (r *res) Resolve(target string) (naming.Watcher, error) {
	return newWatcher(target, r.api)
}

type watch struct {
	sync.Mutex
	updates []*naming.Update
	stop    context.CancelFunc
	done    chan struct{}
	updated chan struct{}
}

func newWatcher(target string, api client.KeysAPI) (*watch, error) {
	var (
		ctx context.Context
		w   = &watch{}
	)
	watcher := api.Watcher(`/`+target, &client.WatcherOptions{Recursive: true})
	ctx, w.stop = context.WithCancel(context.Background())

	res, err := api.Get(ctx, `/`+target, &client.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	if !res.Node.Dir {
		return nil, errors.New(`wrong node`)
	}
	w.updated = make(chan struct{})
	w.addUpdates(naming.Add, res.Node.Nodes...)

	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		for {
			res, err = watcher.Next(ctx)
			if err != nil {
				panic(err)
			}

			switch res.Action {
			case `set`, `update`, `create`:
				w.addUpdates(naming.Add, res.Node)
			case `delete`:
				w.addUpdates(naming.Delete, res.PrevNode)
			}
		}
	}()

	return w, nil
}

func (w *watch) Next() ([]*naming.Update, error) {
	<-w.updated
	w.Lock()
	defer func() {
		w.updates = nil
		w.updated = make(chan struct{})
		w.Unlock()
	}()
	return w.updates, nil
}

func (w *watch) Close() {
	w.stop()
	<-w.done
}

func (w *watch) addUpdates(op naming.Operation, nodes ...*client.Node) {
	w.Lock()
	for _, n := range nodes {
		w.updates = append(w.updates, &naming.Update{
			Op:   op,
			Addr: n.Value,
		})

	}
	close(w.updated)
	w.Unlock()
}
