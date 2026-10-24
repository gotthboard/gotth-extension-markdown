// Command sanitized-run is the only accepted entry point for release validation
// and construction. It is compiled statically so hostile dynamic-loader state
// cannot execute before the environment is cleared.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: sanitized-run <command> [args...]")
		return 2
	}
	dirty := os.Getenv("GOTTH_EXTENSION_EVIDENCE_DIRTY")
	if dirty != "1" {
		dirty = "0"
	}
	runRoot, err := os.MkdirTemp("/tmp", "gm.")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create sanitized run root:", err)
		return 1
	}
	defer os.RemoveAll(runRoot)
	for _, dir := range []string{"home", "tmp", "gocache"} {
		if err := os.Mkdir(filepath.Join(runRoot, dir), 0o700); err != nil {
			fmt.Fprintln(os.Stderr, "initialize sanitized run root:", err)
			return 1
		}
	}
	env := []string{
		"AR=/usr/bin/ar",
		"CC=/usr/bin/gcc",
		"CGO_CFLAGS=-O2 -g",
		"CGO_CPPFLAGS=",
		"CGO_CXXFLAGS=-O2 -g",
		"CGO_ENABLED=0",
		"CGO_FFLAGS=-O2 -g",
		"CGO_LDFLAGS=-O2 -g",
		"CXX=/usr/bin/g++",
		"GIT_ATTR_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_NO_REPLACE_OBJECTS=1",
		"GO111MODULE=on",
		"GOAMD64=v1",
		"GOARCH=amd64",
		"GOCACHE=" + filepath.Join(runRoot, "gocache"),
		"GOCACHEPROG=",
		"GODEBUG=",
		"GOENV=off",
		"GOEXPERIMENT=",
		"GOFIPS140=off",
		"GOFLAGS=",
		"GOMAXPROCS=2",
		"GOMODCACHE=/home/linus/go/pkg/mod",
		"GONOPROXY=",
		"GONOSUMDB=",
		"GOOS=linux",
		"GOPATH=/home/linus/go",
		"GOPRIVATE=",
		"GOPROXY=off",
		"GOROOT=/usr/lib/go",
		"GOSUMDB=off",
		"GOTELEMETRY=off",
		"GOTOOLCHAIN=local",
		"GOTMPDIR=" + filepath.Join(runRoot, "tmp"),
		"GOVCS=off",
		"GOWORK=off",
		"GOTTH_EXTENSION_EVIDENCE_DIRTY=" + dirty,
		"GOTTH_SANITIZED_ENV=1",
		"HOME=" + filepath.Join(runRoot, "home"),
		"LANG=C",
		"LC_ALL=C",
		"PATH=/usr/lib/go/bin:/usr/bin:/bin",
		"PKG_CONFIG=/usr/bin/pkg-config",
		"SOURCE_DATE_EPOCH=0",
		"TMPDIR=" + filepath.Join(runRoot, "tmp"),
		"TZ=UTC",
	}
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	forward := make(chan os.Signal, 8)
	signal.Notify(forward, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(forward)
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "start sanitized command:", err)
		return 1
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for sig := range forward {
			if s, ok := sig.(syscall.Signal); ok {
				_ = syscall.Kill(-cmd.Process.Pid, s)
			}
		}
	}()
	err = cmd.Wait()
	signal.Stop(forward)
	close(forward)
	<-done
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	fmt.Fprintln(os.Stderr, "wait for sanitized command:", err)
	return 1
}
