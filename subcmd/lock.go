package subcmd

import (
	"context"

	"github.com/tukejonny/raftlock/pb"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var (
	id string
)

var Lock = cli.Command{
	Name:      "lock",
	Aliases:   []string{"l"},
	Usage:     "raftlock's lock control",
	ArgsUsage: " ",
	Subcommands: []cli.Command{
		cli.Command{
			Name:      "acquire",
			Aliases:   []string{"a"},
			Usage:     "acquire lock",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "grpc, g",
					Usage:       "grpc address such as :8080",
					Destination: &grpcAddr,
					EnvVar:      "RAFTLOCK_GRPC_ADDRESS",
				},
				cli.StringFlag{
					Name:        "id, i",
					Usage:       "lock id",
					Destination: &id,
				},
			},
			Action: func(cliCtx *cli.Context) error {
				conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
				if err != nil {
					panic(err)
				}
				defer conn.Close()

				cli := pb.NewRaftLockClient(conn)
				_, err = cli.AcquireLock(context.TODO(), &pb.AcquireLockRequest{Id: id})
				if err != nil {
					panic(err)
				}

				return nil
			},
		},
		cli.Command{
			Name:      "release",
			Aliases:   []string{"r"},
			Usage:     "release lock",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "grpc, g",
					Usage:       "grpc address such as :8080",
					Destination: &grpcAddr,
					EnvVar:      "RAFTLOCK_GRPC_ADDRESS",
				},
				cli.StringFlag{
					Name:        "id, i",
					Usage:       "lock id",
					Destination: &id,
				},
			},
			Action: func(cliCtx *cli.Context) error {
				conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
				if err != nil {
					panic(err)
				}
				defer conn.Close()

				cli := pb.NewRaftLockClient(conn)
				_, err = cli.ReleaseLock(context.TODO(), &pb.ReleaseLockRequest{Id: id})
				if err != nil {
					panic(err)
				}

				return nil
			},
		},
	},
}
