package main

import (
	"github.com/patstar123/meeting-base/spr"
	"github.com/patstar123/meeting-base/spr/test/comm"
	"time"

	"github.com/livekit/protocol/logger"
)

type SimpleProxy struct {
	caller *spr.SubProcCaller
}

func (p *SimpleProxy) CustomTypeValues() []any {
	return []any{
		comm.Args{},
	}
}

func (p *SimpleProxy) Multiply(a, b int) (error, int) {
	args := &comm.Args{A: a, B: b}
	var reply int
	err := p.caller.Call("simple.Multiply", args, &reply)
	return err, reply
}

type ComplexProxy struct {
	caller *spr.SubProcCaller
}

func (p *ComplexProxy) CustomTypeValues() []any {
	return []any{
		base.SUCCESS,
		comm.Args2{},
		comm.Flags{},
	}
}

func (p *ComplexProxy) Init(value1, value2 string, flags *comm.Flags) base.Result {
	args := &comm.Args2{
		StringValue1: value1,
		StringValue2: value2,
		Flags1:       flags,
	}
	var res base.Result
	err := p.caller.Call("complex.Init", args, &res)
	if err != nil {
		logger.Warnw("lost runner", err)
		p.caller.TerminateRunnerFastly()
		return base.REMOTE_SYSTEM_ERROR.AppendErr("Init", err)
	}
	return res
}

type Simple2 struct{}

func (s *Simple2) CustomTypeValues() []any {
	return []any{
		comm.Args{},
	}
}

func (s *Simple2) Plus(args *comm.Args, reply *int) error {
	*reply = args.A + args.B
	return nil
}

func main() {
	base.InitSimpleLogger("test", "debug")

	cbObjs := map[string]spr.RpcSvr{
		"simple2": &Simple2{},
	}

	cbListener := &spr.SubProcCBListener{}
	cbListener.StartLoop("cbListener", cbObjs)

	caller := spr.SubProcCaller{}
	res := caller.CreateAndConnectRunner("spr_test_runner", "./spr_test_runner", []spr.RunnerTyper{
		&SimpleProxy{},
		&ComplexProxy{},
	}, cbListener.GetPort())
	if !res.IsOk() {
		return
	}

	proxy1 := &SimpleProxy{&caller}
	err, value := proxy1.Multiply(3, 4)
	logger.Infow("Multiply:", "err", err, "value", value)

	proxy2 := &ComplexProxy{&caller}
	res = proxy2.Init("v1", "v2", &comm.Flags{
		BoolValue:   true,
		StringValue: "StringValue",
		U32Value:    100,
		U16Value:    200,
	})
	if res.IsOk() {
		logger.Infow("Init:", "data", res.Data(),
			"flags", *res.Data().(comm.Args2).Flags1)
	} else {
		logger.Infow("Init:", "res", res)
	}

	time.Sleep(5 * time.Second)

	res = proxy2.Init("v21", "v22", &comm.Flags{
		BoolValue:   false,
		StringValue: "StringValue2",
		U32Value:    102,
		U16Value:    202,
	})
	if res.IsOk() {
		logger.Infow("Init2:", "data", res.Data(),
			"flags", *res.Data().(comm.Args2).Flags1)
	} else {
		logger.Infow("Init2:", "res", res)
	}

	time.Sleep(5 * time.Second)

	caller.TerminateRunnerSafely()
}
