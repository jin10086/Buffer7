package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"github.com/buffer7/buffer7/pkg/proxy"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: buffer7 <command> [args...]")
		os.Exit(1)
	}

	// 1. Find a free port
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	// 2. Start Proxy in background
	p, _ := proxy.NewProxy("https://registry.npmjs.org")
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: p}
	go server.ListenAndServe()

	// 3. Prepare Subprocess
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 4. Inject Env
	proxyURL := fmt.Sprintf("http://localhost:%d", port)
	env := os.Environ()
	env = append(env, "NPM_CONFIG_REGISTRY="+proxyURL)
	cmd.Env = env

	// 5. Run
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
