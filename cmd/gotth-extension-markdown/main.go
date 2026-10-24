package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gotthboard/gotth-extension-markdown/internal/markdown"
	rendererv1 "github.com/gotthboard/gotth-extension-markdown/proto/gotth/markdown/renderer/v1"
	extensionsv1 "github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(markdown.BuildVersion)
		return
	}
	if len(os.Args) != 1 || run() != nil {
		fmt.Fprintln(os.Stderr, "gotth-extension-markdown: startup failed")
		os.Exit(1)
	}
}

func run() error {
	runtimeConfig, err := markdown.LoadRuntimeFromEnv()
	if err != nil {
		return err
	}
	debug.SetMemoryLimit(markdown.ProcessMemoryLimitBytes)
	service, err := markdown.NewService(runtimeConfig)
	if err != nil {
		return err
	}
	listener, err := net.Listen("unix", runtimeConfig.SocketPath)
	if err != nil {
		return err
	}
	defer listener.Close()
	if err := os.Chmod(runtimeConfig.SocketPath, 0o600); err != nil {
		return err
	}
	server := grpc.NewServer(grpc.MaxRecvMsgSize(70<<10), grpc.MaxSendMsgSize(400<<10), grpc.MaxConcurrentStreams(uint32(runtimeConfig.Config.MaxConcurrent+4)), grpc.UnaryInterceptor(service.UnaryAuthInterceptor))
	extensionsv1.RegisterExtensionControlServer(server, service)
	rendererv1.RegisterMarkdownRendererServer(server, service)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	stopped := make(chan struct{})
	go func() {
		select {
		case <-signals:
			finished := make(chan struct{})
			go func() { server.GracefulStop(); close(finished) }()
			select {
			case <-finished:
			case <-time.After(2 * time.Second):
				server.Stop()
			}
		case <-stopped:
		}
	}()
	err = server.Serve(listener)
	close(stopped)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return nil
}
