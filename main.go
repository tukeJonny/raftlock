package main

import (
	"os"

	"github.com/tukejonny/raftlock/subcmd"
	"github.com/urfave/cli"
)

func main() {
	os.Exit(cliMain())
}

func cliMain() int {
	app := cli.NewApp()
	app.Name = "raftlock"
	app.Usage = "lockmgr with raft concensus example"
	app.Description = ""
	app.Authors = []cli.Author{
		{
			Name:  "tukeJonny",
			Email: "ne250143@yahoo.co.jp",
		},
	}
	app.HelpName = "raftlock"

	app.Commands = []cli.Command{
		subcmd.Lock,
		subcmd.Serve,
	}

	app.Action = func(ctx *cli.Context) error {
		cli.ShowAppHelp(ctx)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		return 1
	}

	return 0
}
