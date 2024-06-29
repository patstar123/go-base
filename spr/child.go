package spr

import (
	"errors"
	"lx/meeting/base"
	"lx/meeting/base/logger"
	"lx/meeting/base/utils"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

// SubProcRunner 子进程执行器
type SubProcRunner struct {
	listener net.Listener
	callback *SubProcCallback
}

func (c *SubProcRunner) Run(rpcObjs map[string] /*rpcObjName*/ RpcSvr, cbTypes []CallbackTyper) base.Result {
	utils.SetAbortCallback(func() { c.StopLoop() })
	utils.ListenAbortSignal()
	defer utils.ListenPanic(true)

	if c.listener != nil {
		res := base.LOGICAL_ERROR.AppendMsg("has been running")
		logger.Errorw("[child]Run failed", res)
		return res
	}

	if len(os.Args) < 2 {
		res := base.INVALID_PARAM.AppendMsg("lost command params: child name")
		logger.Errorw("[child]Run failed", res)
		return res
	}
	if len(os.Args) < 3 {
		res := base.INVALID_PARAM.AppendMsg("lost command params: log level")
		logger.Errorw("[child]Run failed", res)
		return res
	}
	if len(os.Args) < 4 {
		res := base.INVALID_PARAM.AppendMsg("lost command params: rpc port")
		logger.Errorw("[child]Run failed", res)
		return res
	}

	cbPort := ""
	if len(os.Args) >= 5 {
		cbPort = os.Args[4]
	}

	name := os.Args[1]
	logger.SetLogLevel(os.Args[2])
	rpcPort := os.Args[3]
	_, err := strconv.Atoi(rpcPort)
	if err != nil {
		res := base.INVALID_PARAM.AppendErr("invalid rpc port: "+rpcPort, err)
		logger.Errorw("[child]Run failed", res)
		return res
	}
	if cbPort != "" {
		_, err = strconv.Atoi(cbPort)
		if err != nil {
			res := base.INVALID_PARAM.AppendErr("invalid callback port: "+cbPort, err)
			logger.Errorw("[child]Run failed", res)
			return res
		}
	}

	rpc.RegisterName(BaseChildName, &BaseChild{c})
	if rpcObjs != nil {
		for rpcName, rpcSvr := range rpcObjs {
			LoadRpcTypes(rpcSvr)
			rpc.RegisterName(rpcName, rpcSvr)
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+rpcPort)
	if err != nil {
		res := base.INVALID_PARAM.AppendErr("start rpc server failed", err)
		logger.Errorw("[child]Run failed", res)
		return res
	}
	c.listener = listener

	if cbPort != "" {
		c.callback = &SubProcCallback{}
		go func() {
			res := c.callback.ConnectListener(cbPort, cbTypes)
			if res.IsOk() {
				logger.Infow("[child]" + name + " connected to callback listener")
			} else {
				logger.Warnw("[child]"+name+" connect to callback listener failed", res)
			}
		}()
	}

	logger.Infow("[child]" + name + " listening on 127.0.0.1:" + rpcPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			logger.Warnw("[child]rpc accepting error:", err)
			continue
		}

		go rpc.ServeConn(conn)
	}

	logger.Infow("[child]" + name + " exited")
	return base.SUCCESS
}

func (c *SubProcRunner) StopLoop() {
	if c.listener != nil {
		c.listener.Close()
		c.listener = nil
	}
	if c.callback != nil {
		c.callback.Disconnect()
		c.callback = nil
	}
}

func (c *SubProcRunner) Callback(cbMethod string, args any, reply any) error {
	if c.callback == nil {
		res := base.LOGICAL_ERROR.AppendMsg("parent not set callback port")
		logger.Errorw("[child]Callback failed", res)
		return res
	}

	return c.callback.Call(cbMethod, args, reply)
}

type BaseChild struct {
	handler *SubProcRunner
}

func (c *BaseChild) RpcStopChild(args int, reply *int) error {
	if c.handler != nil {
		handler := c.handler
		c.handler = nil
		go func() {
			time.Sleep(10 * time.Millisecond)
			handler.StopLoop()
		}()
	}
	*reply = 0
	return nil
}

func (c *BaseChild) RpcPing(args int, reply *bool) error {
	*reply = c.handler != nil
	return nil
}
