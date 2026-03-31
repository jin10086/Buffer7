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

	// 1. Detect command and select target
	cmdName := os.Args[1]
	targetRegistry := "https://registry.npmjs.org"
	envVar := "NPM_CONFIG_REGISTRY"
	
	if cmdName == "pip" || cmdName == "pip3" || cmdName == "poetry" || cmdName == "pdm" {
		targetRegistry = "https://pypi.org"
		envVar = "PIP_INDEX_URL"
	}

	// 2. Find a free port
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	// 3. Start Proxy in background
	p, _ := proxy.NewProxy(targetRegistry)
	server := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: p}
	go server.ListenAndServe()

	// 4. Prepare Subprocess
	cmdArgs := os.Args[2:]
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 5. Inject Env
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if envVar == "PIP_INDEX_URL" {
		proxyURL += "/simple" // pip 要求指向 /pypi 或 /simple
	}
	env := os.Environ()
	env = append(env, envVar+"="+proxyURL)
	cmd.Env = env

	// 5. Run
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
