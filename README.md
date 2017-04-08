# gRPC balancer example

Steps:
* Get and install this example:
```sh
$ go get github.com/reinventer/grpc-lb-test/cmd/...
```
* Install and run etcd
* Add some keys to etcd:
```sh
  $ etcdctl set balancer/service/server0 127.0.0.1:15080
  $ etcdctl set balancer/service/server1 127.0.0.1:15081
```
* Run examples
if you want to test local balancer
```
$ local-balancer
```
if you want to test remote balancer
```
$ remote-balancer
```

* Try to add/remove available servers to etcd:
```sh
  $ etcdctl set balancer/service/server2 127.0.0.1:15082
  $ etcdctl rm balancer/service/server1
```
* Watch logs
* ...
* PROFIT
