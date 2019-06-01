package raftlock

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	"github.com/tukejonny/raftlock/pb"
	"github.com/tukejonny/raftlock/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Service struct {
	Store *store.Store
}

func NewService() (grpcServer *grpc.Server, svc *Service) {
	grpcServer = grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    1 * time.Second,
			Timeout: 1 * time.Second,
		}),
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
		),
	)

	svc = &Service{
		Store: store.NewStore(),
	}

	pb.RegisterRaftLockServer(grpcServer, svc)
	reflection.Register(grpcServer)

	return
}

func (s *Service) AcquireLock(ctx context.Context, req *pb.AcquireLockRequest) (*pb.AcquireLockResponse, error) {
	var id = req.GetId()

	_, err := s.Store.Acquire(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.AcquireLockResponse{}, nil
}

func (s *Service) ReleaseLock(ctx context.Context, req *pb.ReleaseLockRequest) (*pb.ReleaseLockResponse, error) {
	var id = req.GetId()

	if err := s.Store.Release(id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.ReleaseLockResponse{}, nil
}

func (s *Service) JoinCluster(ctx context.Context, req *pb.JoinClusterRequest) (*pb.JoinClusterResponse, error) {
	var (
		nodeID     = req.GetNodeId()
		remoteAddr = req.GetRemoteAddress()
	)
	if err := s.Store.Join(nodeID, remoteAddr); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.JoinClusterResponse{}, nil
}

func (s *Service) Nodes(ctx context.Context, req *pb.NodesRequest) (*pb.NodesResponse, error) {
	nodes, err := s.Store.Nodes()
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	nodeList := make([]*pb.Node, len(nodes))
	for i := 0; i < len(nodes); i++ {
		nodeList[i] = &pb.Node{
			NodeId:   string(nodes[i].ID),
			Addr:     string(nodes[i].Address),
			Suffrage: string(nodes[i].Suffrage),
		}
	}

	return &pb.NodesResponse{Nodes: nodeList}, nil
}

func (s *Service) Stats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	return &pb.StatsResponse{Stats: s.Store.Stats()}, nil
}
