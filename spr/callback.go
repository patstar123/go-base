package spr

import (
	"errors"
	"fmt"
	"lx/meeting/base"
	"lx/meeting/base/logger"
	"net"
	"net/rpc"
)

// SubProcCBListener 子进程回调监听器
type SubProcCBListener struct {
	name     string
	listener net.Listener
	port     int
}

func (l *SubProcCBListener) StartLoop(name string, rpcObjs map[string] /*rpcObjName*/ RpcSvr) base.Result {
	if l.listener != nil {
		res := base.LOGICAL_ERROR.AppendMsg("has been running")
		logger.Errorw("[cbSvr]StartLoop failed", res)
		return res
	}

	l.name = name
	l.port = -1

	// 查找可用callback端口
	port := findAvailablePort()
	if port <= 0 {
		res := base.INTERNAL_ERROR.AppendMsg("has no available callback port")
		logger.Warnw("[cbSvr]StartLoop failed for "+l.name, res)
		return res
	}
	cbPort := fmt.Sprintf("%v", port)

	if debugEnabled {
		cbPort = debugCallbackPort
	}

	// 注册回调方法和数据类型
	if rpcObjs != nil {
		for rpcName, rpcSvr := range rpcObjs {
			LoadRpcTypes(rpcSvr)
			rpc.RegisterName(rpcName, rpcSvr)
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+cbPort)
	if err != nil {
		res := base.INVALID_PARAM.AppendErr("start callback listener failed", err)
		logger.Errorw("[cbSvr]StartLoop failed", res)
		return res
	}

	l.listener = listener
	l.port = port

	go func() {
		logger.Infow("[cbSvr]" + l.name + " listening on 127.0.0.1:" + cbPort)
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break
				}

				logger.Warnw("[cbSvr]rpc accepting error:", err)
				continue
			}

			go rpc.ServeConn(conn)
		}
		logger.Infow("[cbSvr]" + l.name + " exited")
	}()

	return base.SUCCESS
}

func (l *SubProcCBListener) StopLoop() {
	if l.listener != nil {
		l.listener.Close()
		l.listener = nil
		l.name = ""
		l.port = -1
	}
}

func (l *SubProcCBListener) GetPort() int {
	return l.port
}

// SubProcCallback 子进程回调器
type SubProcCallback struct {
	cli *rpc.Client
}

type CallbackTyper = RunnerTyper

func (c *SubProcCallback) ConnectListener(cbPort string, typers []CallbackTyper) base.Result {
	if c.cli != nil {
		res := base.LOGICAL_ERROR.AppendMsg("has been connected")
		logger.Errorw("[cbCli]ConnectListener failed", res)
		return res
	}

	client, err := rpc.Dial("tcp", "localhost:"+cbPort)
	if err != nil {
		res := base.INTERNAL_ERROR.AppendErr("connect callback listener failed", err)
		logger.Warnw("[cbCli]ConnectListener failed", res)
		return res
	}

	if typers != nil {
		for _, typer := range typers {
			LoadRpcTypes(typer)
		}
	}

	c.cli = client
	return base.SUCCESS
}

func (c *SubProcCallback) Disconnect() {
	if c.cli != nil {
		c.cli.Close()
		c.cli = nil
	}
}

func (c *SubProcCallback) Call(serviceMethod string, args any, reply any) error {
	if c.cli == nil {
		res := base.LOGICAL_ERROR.AppendMsg("connect first")
		logger.Errorw("[cbCli]Call failed", res)
		return res
	}

	return c.cli.Call(serviceMethod, args, reply)
}
