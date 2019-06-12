package subcmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/tukejonny/raftlock/pb"
	"github.com/tukejonny/raftlock/raftlock"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var (
	nodeID   string
	grpcAddr string

	raftAddr string
	raftDir  string

	joinAddr      string
	advertiseAddr string

	maxpool           int
	retainSnapshotCnt int

	timeout int
)

func startService(isLeader bool) (<-chan error, error) {
	grpcServer, svc := raftlock.NewService()
	svc.Store.RaftBind = raftAddr
	svc.Store.RaftDir = raftDir

	log.Printf("nodeID=%s\n", nodeID)
	log.Printf("grpcAddr=%s\n", grpcAddr)
	log.Printf("raftAddr=%s\n", raftAddr)
	log.Printf("raftDir=%s\n", raftDir)
	log.Printf("joinAddr=%s\n", joinAddr)

	if err := svc.Store.Open(isLeader, nodeID, maxpool, retainSnapshotCnt, time.Duration(timeout)*time.Second); err != nil {
		log.Println(err.Error())
		return nil, err
	}

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	grpcErrCh := make(chan error, 1)
	go func() {
		defer close(grpcErrCh)
		grpcErrCh <- grpcServer.Serve(lis)
	}()

	return grpcErrCh, nil
}

var Serve = cli.Command{
	Name:      "serve",
	Aliases:   []string{"s"},
	Usage:     "server control",
	ArgsUsage: " ",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "id, i",
			Usage:       "node id",
			Destination: &nodeID,
			EnvVar:      "RAFTLOCK_NODE_ID",
		},
		cli.StringFlag{
			Name:        "grpc, g",
			Usage:       "grpc address such as :8080",
			Destination: &grpcAddr,
			EnvVar:      "RAFTLOCK_GRPC_ADDRESS",
		},
		cli.StringFlag{
			Name:        "raft, r",
			Usage:       "raft address such as :8080",
			Destination: &raftAddr,
			EnvVar:      "RAFTLOCK_RAFT_ADDRESS",
		},
		cli.StringFlag{
			Name:        "dir, d",
			Usage:       "raft storage directory",
			Destination: &raftDir,
			EnvVar:      "RAFTLOCK_RAFT_DIRECTORY",
		},
		cli.IntFlag{
			Name:        "pool, p",
			Usage:       "max pool size",
			Value:       3,
			Destination: &maxpool,
			EnvVar:      "RAFTLOCK_MAX_POOLSIZE",
		},
		cli.IntFlag{
			Name:        "snapshots, s",
			Usage:       "retain snapshots count",
			Value:       3,
			Destination: &retainSnapshotCnt,
			EnvVar:      "RAFTLOCK_RETAIN_SNAPSHOT_COUNT",
		},
		cli.IntFlag{
			Name:        "timeout, t",
			Usage:       "timeout",
			Value:       10,
			Destination: &timeout,
			EnvVar:      "RAFTLOCK_TIMEOUT",
		},
	},
	Subcommands: []cli.Command{
		cli.Command{
			Name:      "run",
			Aliases:   []string{"r"},
			Usage:     "start server",
			ArgsUsage: " ",
			Action: func(cliCtx *cli.Context) error {
				grpcErrCh, err := startService(true)
				if err != nil {
					log.Println(err.Error())
					return err
				}

				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, os.Interrupt)

				select {
				case err := <-grpcErrCh:
					log.Println(err.Error())
					return err
				case <-sigCh:
				}

				return nil
			},
		},
		cli.Command{
			Name:      "join",
			Aliases:   []string{"j"},
			Usage:     "join raft cluster",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "join, j",
					Usage:       "join address such as :8080",
					Destination: &joinAddr,
					EnvVar:      "RAFTLOCK_JOIN_ADDRESS",
				},
				cli.StringFlag{
					Name:        "advertise, a",
					Usage:       "advertise address",
					Destination: &advertiseAddr,
					EnvVar:      "RAFTLOCK_ADVERTISE_ADDRESS",
				},
			},
			Action: func(cliCtx *cli.Context) error {
				grpcErrCh, err := startService(false)
				if err != nil {
					log.Println(err.Error())
					return err
				}

				conn, err := grpc.Dial(joinAddr, grpc.WithInsecure())
				if err != nil {
					log.Println(err.Error())
					return err
				}
				defer conn.Close()

				cli := pb.NewRaftLockClient(conn)
				_, err = cli.JoinCluster(context.TODO(), &pb.JoinClusterRequest{NodeId: nodeID, RemoteAddress: advertiseAddr})
				if err != nil {
					log.Println(err.Error())
					return err
				}

				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, os.Interrupt)

				select {
				case err := <-grpcErrCh:
					fmt.Println(err.Error())
					return err
				case <-sigCh:
				}

				return nil
			},
		},
	},
}
