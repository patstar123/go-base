package spr

import (
	"avd/meeting/base"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"time"

	"avd/meeting/base/logger"
)

// SubProcCaller 子进程控制器
type SubProcCaller struct {
	name string
	cli  *rpc.Client
	cmd  *exec.Cmd
}

func (p *SubProcCaller) CreateAndConnectRunner(nameSrc, programSrc string,
	typers []RunnerTyper, callbackPort int) base.Result {
	if p.cli != nil || p.cmd != nil {
		res := base.LOGICAL_ERROR.AppendMsg("has been created")
		logger.Warnw("[parent]CreateAndConnectRunner failed for "+nameSrc, res)
		return res
	}

	var err error
	name, err := p.safeCheck(nameSrc)
	if err != nil {
		res := base.LOGICAL_ERROR.AppendMsg(err.Error())
		logger.Warnw("[parent]CreateAndConnectRunner safeCheck failed for "+nameSrc, res)
		return res
	}
	program, err := p.safeCheck(programSrc)
	if err != nil {
		res := base.LOGICAL_ERROR.AppendMsg(err.Error())
		logger.Warnw("[parent]CreateAndConnectRunner safeCheck failed for "+programSrc, res)
		return res
	}

	p.name = name

	// 查找可用rpc端口
	port := findAvailablePort()
	if port <= 0 {
		res := base.INTERNAL_ERROR.AppendMsg("has no available rpc port")
		logger.Warnw("[parent]CreateAndConnectRunner failed for "+name, res)
		return res
	}
	rpcPort := fmt.Sprintf("%v", port)

	// 回调端口
	cbPort := ""
	if callbackPort > 0 {
		cbPort = fmt.Sprintf("%v", callbackPort)
	}

	level := logger.GetLogLevel()
	if level == "" {
		level = "info"
	}

	var cmd *exec.Cmd
	if !debugEnabled {
		// 启动子进程
		logger.Infow("[parent]try to start child process", "name", name,
			"log", level, "rpcPort", rpcPort, "cbPort", cbPort)
		cmd = exec.Command(program, name, level, rpcPort, cbPort)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			res := base.INTERNAL_ERROR.AppendErr("start child process failed", err)
			logger.Warnw("[parent]CreateAndConnectRunner failed for "+name, res)
			return res
		}
	} else {
		rpcPort = debugRpcPort
		cbPort = debugCallbackPort
		logger.Warnw("!!!debug:"+program+" "+name+" "+level+" "+rpcPort+" "+cbPort, nil)
	}

	// 连接子进程RPC服务
	var client *rpc.Client
	for i := 0; i < 250; i++ {
		client, err = rpc.Dial("tcp", "localhost:"+rpcPort)
		if err == nil {
			break
		}

		if i < 3 { // 100ms
			logger.Debugw("connect child process rpc port failed, try later")
		} else if i < 20 { // 1000ms
			logger.Infow("connect child process rpc port failed, try later")
		} else {
			logger.Warnw("connect child process rpc port failed, try later", nil)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		res := base.INTERNAL_ERROR.AppendErr("connect child process rpc port failed after 5s", err)
		logger.Warnw("[parent]CreateAndConnectRunner failed for "+name, res)
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
			go cmd.Wait()
		}
		return res
	}

	// 注册远端的类型到Rpc服务
	if typers != nil {
		for _, typer := range typers {
			LoadRpcTypes(typer)
		}
	}

	p.cmd = cmd
	p.cli = client
	return base.SUCCESS
}

func (p *SubProcCaller) safeCheck(input string) (string, error) {
	return input, nil
}

func (p *SubProcCaller) TerminateRunnerSafely() {
	if p.cli == nil {
		return
	}

	logger.Infow("[parent]try to stop child: " + p.name)
	var replay int
	err := p.cli.Call(StopMethod, 0, &replay)
	if err != nil {
		logger.Warnw("[parent]stop child failed, force to kill it: "+p.name, err)
		if p.cmd != nil && p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
	} else {
		logger.Infow("[parent]child stopped: " + p.name)
	}

	if p.cmd != nil {
		cmd := p.cmd
		go cmd.Wait()
		p.cmd = nil
	}

	p.cli.Close()
	p.cli = nil
}

func (p *SubProcCaller) TerminateRunnerFastly() {
	if p.cmd != nil {
		logger.Infow("[parent]force to kill child: " + p.name)
		if p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
		cmd := p.cmd
		go cmd.Wait()
		p.cmd = nil
	}

	if p.cli != nil {
		p.cli.Close()
		p.cli = nil
	}
}

func (p *SubProcCaller) Call(serviceMethod string, args any, reply any) error {
	if (!debugEnabled && p.cmd == nil) || p.cli == nil {
		res := base.LOGICAL_ERROR.AppendMsg("create child first: " + p.name)
		logger.Warnw("[parent]Call failed", res)
		return res
	}

	return p.cli.Call(serviceMethod, args, reply)
}

func (p *SubProcCaller) Ping() (error, bool /*alive*/) {
	var replay bool
	err := p.Call(PingMethod, 0, &replay)
	return err, replay
}
