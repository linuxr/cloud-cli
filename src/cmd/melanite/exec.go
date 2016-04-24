package main

import (
	"errors"
	"runner"
	"runner/sshrunner"

	"fmt"

	"github.com/codegangsta/cli"
)

var (
	// ErrGroupORNodeRequired require group or node option
	ErrGroupORNodeRequired = errors.New("option -g/--group or -n/--node is required")
	// ErrCmdRequired require cmd option
	ErrCmdRequired = errors.New("option -c/--cmd is required")
)

type execParams struct {
	GroupName string
	NodeName  string
	User      string
	Cmd       string
}

func initExecSubCmd(app *cli.App) {
	execSubCmd := cli.Command{
		Name:        "exec",
		Usage:       "exec <options>",
		Description: "exec command on groups or nodes",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "g,group",
				Value: "",
				Usage: "exec command on group",
			},
			cli.StringFlag{
				Name:  "n,node",
				Value: "",
				Usage: "exec command on node",
			},
			cli.StringFlag{
				Name:  "u,user",
				Value: "root",
				Usage: "user who exec the command",
			},
			cli.StringFlag{
				Name:  "c,cmd",
				Value: "",
				Usage: "command for exec",
			},
		},
		Action: func(c *cli.Context) {
			var ep, err = checkExecParams(c)
			if err != nil {
				fmt.Println(err)
				cli.ShowCommandHelp(c, "exec")
				return
			}
			if err = execCmd(ep); err != nil {
				fmt.Println(err)
			}
		},
	}

	if app.Commands == nil {
		app.Commands = cli.Commands{execSubCmd}
	} else {
		app.Commands = append(app.Commands, execSubCmd)
	}
}

func checkExecParams(c *cli.Context) (execParams, error) {
	var ep = execParams{
		GroupName: c.String("group"),
		NodeName:  c.String("node"),
		User:      c.String("user"),
		Cmd:       c.String("cmd"),
	}

	if ep.GroupName == "" && ep.NodeName == "" {
		return ep, ErrGroupORNodeRequired
	}

	if ep.Cmd == "" {
		return ep, ErrCmdRequired
	}

	if ep.User == "" {
		ep.User = "root"
	}

	return ep, nil
}

func execCmd(ep execParams) error {
	// TODO should use sshrunner from config

	// get node info for exec
	repo := GetRepo()
	var groups, err = repo.FilterNodeGroupsAndNodes(ep.GroupName, ep.NodeName)
	if err != nil {
		return err
	}

	// exec cmd on node
	for _, g := range groups {
		for _, n := range g.Nodes {
			fmt.Printf("start exec cmd(%s) on Group(%s)->Node(%s): >>>\n", ep.Cmd, g.Name, n.Name)
			var runCmd = sshrunner.New(n.User, n.Password, n.KeyPath, n.Host, n.Port)
			var input = runner.Input{
				ExecHost: n.Host,
				ExecUser: ep.User,
				Command:  ep.Cmd,
			}

			// display result
			if output, err := runCmd.SyncExec(input); err != nil {
				fmt.Println(err)
			} else {
				displayExecResult(output)
			}
		}
	}
	return nil
}

func displayExecResult(output *runner.Output) {
	fmt.Printf("start time: %s\n", output.ExecStart.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("end time:   %s\n", output.ExecEnd.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("stdout >>>\n%s\n", output.StdOutput)
	fmt.Printf("stderr >>>\n%s\n", output.StdError)
}