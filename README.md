# gRPC balancer example

Steps:
* Get and install this example:
```sh
$ go get github.com/reinventer/grpc-lb-test
```
* Install and run etcd
* Add some keys to etcd:
```sh
  $ etcdctl set balancer/service/server0 localhost:15080
  $ etcdctl set balancer/service/server1 localhost:15081
```
* Run example
```
$ grpc-lb-test
```
* Try to add/remove available servers to etcd:
```sh
  $ etcdctl set balancer/service/server2 localhost:15082
  $ etcdctl rm balancer/service/server1
```
* Watch logs
* ...
* PROFIT
