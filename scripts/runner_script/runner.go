package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const MEMORY_FILE string = "/tmp/juri-memory"

const (
	// req types accepted by this runner process.
	RequestCompile = "compile"
	RequestRun     = "run"
	RequestPing    = "ping"
	RequestStop    = "stop"
)

// dto structs
type Request struct {
	Type string `json:"type"`

	Language string `json:"language,omitempty"`
	Stdin    string `json:"stdin,omitempty"`

	// time limit in milliseconds per run request.
	TimeLimitMS int `json:"time_limit,omitempty"`
}

type Response struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`

	Metadata *Metadata `json:"metadata,omitempty"`
}

type Metadata struct {
	RuntimeNS int64 `json:"runtime_ns"`
	MemoryKB  int64 `json:"memory_kb"`
	ExitCode  int   `json:"exit_code"`
	TimedOut  bool  `json:"timed_out"`
}

// language (for the cmd mapping)
type Language struct {
	Compile []string
	Run     []string
}

// contains all the commands necessary for running and compiling
var Languages = map[string]Language{
	"c": {
		Compile: []string{"gcc", "main.c", "-O2", "-o", "program"},
		Run:     []string{"./program"},
	},
	"cpp": {
		Compile: []string{"g++", "main.cpp", "-O2", "-o", "program"},
		Run:     []string{"./program"},
	},
	"java": {
		Compile: []string{"javac", "Main.java"},
		Run:     []string{"java", "Main"},
	},
	"python": {
		Run: []string{"python3", "main.py"},
	},
}

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req Request
		//reads a single json req
		if err := decoder.Decode(&req); err != nil {
			return
		}
		handleRequest(encoder, req)
	}
}

// handler
func handleRequest(encoder *json.Encoder, req Request) {
	defer func() {
		if r := recover(); r != nil {
			_ = encoder.Encode(Response{
				Success: false,
				Error:   "runner panic: " + fmt.Sprint(r),
			})
		}
	}()

	var resp Response

	switch req.Type {
	case RequestPing:
		resp = Response{
			Success: true,
			Output:  "pong",
		}

	case RequestCompile:
		resp = compile(req)

	case RequestRun:
		resp = run(req)

	case RequestStop:
		_ = encoder.Encode(Response{Success: true})
		os.Exit(0)

	default:
		resp = Response{
			Success: false,
			Error:   "unknown request",
		}
	}

	_ = encoder.Encode(resp)
}

func compile(req Request) Response {
	lang := Languages[req.Language]
	// if a language has no compile command(like python for now, js soon)
	if len(lang.Compile) == 0 {
		return Response{
			Success: true,
		}
	}
	//preparing cmd
	cmd := exec.Command(lang.Compile[0], lang.Compile[1:]...)
	out, err := cmd.CombinedOutput() //run cmd along with capturing output

	return Response{
		Success: err == nil,
		Output:  string(out),
		Error:   errorString(err),
	}
}

func run(req Request) Response {
	// refetch the requested language
	lang := Languages[req.Language]
	// defaults to 2 seconds if no limit was provided
	if req.TimeLimitMS <= 0 {
		req.TimeLimitMS = 2000
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(req.TimeLimitMS)*time.Millisecond,
	)
	defer cancel()

	// make the command to run using /usr/bin/time so we can capture peak memory usage in a file
	args := []string{
		"-f", "%M",
		"-o", MEMORY_FILE,
		lang.Run[0],
	}
	args = append(args, lang.Run[1:]...)

	cmd := exec.CommandContext(ctx, "/usr/bin/time", args...)
	// pass input
	cmd.Stdin = strings.NewReader(req.Stdin)
	//running
	start := time.Now() //calculates rt
	out, err := cmd.CombinedOutput()
	runtime := time.Since(start)

	// read the memory file
	memoryKB := int64(0)
	if data, readErr := os.ReadFile(MEMORY_FILE); readErr == nil {
		memoryKB, _ = strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	}

	resp := Response{
		Success: err == nil,
		Output:  string(out),
		Error:   errorString(err),
		Metadata: &Metadata{
			RuntimeNS: runtime.Nanoseconds(),
			ExitCode:  exitCode(err),
			TimedOut:  errors.Is(ctx.Err(), context.DeadlineExceeded),
			MemoryKB:  memoryKB,
		},
	}

	if resp.Metadata.TimedOut {
		resp.Success = false
		resp.Error = "time limit exceeded"
	}

	return resp
}

// fetches exit code
func exitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1 //default
}

// fetches error message
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
