package utils

import (
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"lx/meeting/base/logger"
)

var gOnAbort func()

func SetAbortCallback(onAbort func()) {
	if gOnAbort != nil {
		panic("logical error: SetAbortCallback")
	}
	gOnAbort = onAbort
}

func ListenAbortSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGHUP,
		syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT, syscall.SIGBUS, syscall.SIGFPE,
		syscall.SIGALRM)

	go func() {
		running := true
		for running {
			select {
			case sig := <-sigChan:
				handleSignal(sig)
				running = false
			}
		}
	}()
}

func ListenPanic(abort bool) {
	if r := recover(); r != nil {
		logger.Warnw("got panic:", nil, "desc", r, "program", os.Args[0])
		logger.Infow(string(debug.Stack()))

		if gOnAbort != nil {
			gOnAbort()
		}

		if abort {
			os.Exit(0)
		}
	}
}

func handleSignal(sig os.Signal) {
	logger.Warnw("got signal:", nil, "os.Signal", sig, "program", os.Args[0])

	if gOnAbort != nil {
		gOnAbort()
	}

	if sig == syscall.SIGINT {
		os.Exit(0)
	} else {
		panic(sig)
	}
}
