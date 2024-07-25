package main

import (
	"fmt"
	"github.com/patstar123/go-base/spr"
	"github.com/patstar123/go-base/spr/test/comm"
)

type Simple struct {
}

func (s *Simple) CustomTypeValues() []any {
	return []any{
		comm.Args{},
	}
}

func (s *Simple) Multiply(args *comm.Args, reply *int) error {
	*reply = args.A * args.B
	err, v := (&Simple2Callback{}).Plus(args.A, args.B)
	if err != nil {
		return nil
	} else {
		*reply = *reply + v
	}
	return nil
}

type Complex struct{}

func (c *Complex) CustomTypeValues() []any {
	return []any{
		base.SUCCESS,
		comm.Args2{},
		comm.Flags{},
	}
}

func (c *Complex) Init(args *comm.Args2, reply *base.Result) error {
	fmt.Println("Complex: Init:", args, args.Flags1)
	*reply = base.REMOTE_SYSTEM_ERROR.AppendMsg("test error").SetData(*args)
	return nil
}

type Simple2Callback struct {
}

func (p *Simple2Callback) CustomTypeValues() []any {
	return []any{
		comm.Args{},
	}
}

func (p *Simple2Callback) Plus(a, b int) (error, int) {
	args := &comm.Args{A: a, B: b}
	var reply int
	err := runner.Callback("simple2.Plus", args, &reply)
	return err, reply
}

var runner *spr.SubProcRunner

func main() {
	base.InitSimpleLogger("test", "debug")

	rpcObjs := map[string]spr.RpcSvr{
		"simple":  &Simple{},
		"complex": &Complex{},
	}

	runner = &spr.SubProcRunner{}
	defer runner.StopLoop()
	runner.Run(rpcObjs, []spr.RunnerTyper{&Simple2Callback{}})
}
