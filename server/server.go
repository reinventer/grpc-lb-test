package server

import (
	"github.com/reinventer/grpc-lb-test/proto"
	"golang.org/x/net/context"
)

type serv struct {
	name string
}

func New(name string) *serv {
	return &serv{name}
}

func (s *serv) Who(ctx context.Context, _ *proto.Empty) (*proto.ID, error) {
	return &proto.ID{s.name}, nil
}
