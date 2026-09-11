package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Constants / config

const (
	MemoryStatFile = "mem.stat"
	DefaultTimeout = 2000 * time.Millisecond
	MaxTimeout     = 15000 * time.Millisecond
	MaxOutputBytes = 2 * 1024 * 1024 // 2MB cap on captured stdout/stderr
	WorkDir        = "/workspace"
)

const (
	RequestCompile = "compile"
	RequestRun     = "run"
	RequestPing    = "ping"
	RequestStop    = "stop"
)

// allowedTypes is the whitelist of request types this runner will ever act on.
// Anything else (including stray shell commands like "ls" sent by mistake into
// the JSON stream) is rejected up front rather than reaching exec.Command.
var allowedTypes = map[string]bool{
	RequestCompile: true,
	RequestRun:     true,
	RequestPing:    true,
	RequestStop:    true,
}

// DTOs

type Request struct {
	Type        string `json:"type"`
	Language    string `json:"language,omitempty"`
	Stdin       string `json:"stdin,omitempty"`
	TimeLimitMS int    `json:"time_limit,omitempty"`
}

type Response struct {
	Success  bool     `json:"success"`
	Output   string   `json:"output,omitempty"`
	Error    string   `json:"error,omitempty"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	RuntimeNS int64 `json:"runtime_ns"`
	MemoryKB  int64 `json:"memory_kb"`
	ExitCode  int   `json:"exit_code"`
	TimedOut  bool  `json:"timed_out"`
}

type Language struct {
	Compile []string
	Run     []string
}

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

// main: persistent stdin/stdout JSON loop, hardened against garbage input

func main() {
	if err := os.MkdirAll(WorkDir, 0755); err != nil {
		log.Fatalf("cannot create work dir: %v", err)
	}
	if err := os.Chdir(WorkDir); err != nil {
		log.Fatalf("cannot chdir to work dir: %v", err)
	}

	reader := bufio.NewReaderSize(os.Stdin, 1<<20)
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(os.Stdout)

	// One request handled at a time; a mutex guards the shared work dir /
	// memory-stat file so overlapping "run" calls (shouldn't happen with a
	// single decode loop, but be defensive) never race on disk state.
	var mu sync.Mutex

	for {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			// Malformed JSON (e.g. someone piped "ls\n" straight into the
			// process). Report it and keep the loop alive instead of dying.
			safeEncode(encoder, Response{
				Success: false,
				Error:   fmt.Sprintf("invalid request: could not parse JSON: %v", err),
			})
			// The decoder's internal buffer may be in a bad state after a
			// non-JSON token; resync by dropping the rest of the current line.
			_, _ = reader.ReadString('\n')
			continue
		}

		req, parseErr := parseRequest(raw)
		if parseErr != nil {
			safeEncode(encoder, Response{Success: false, Error: parseErr.Error()})
			continue
		}

		mu.Lock()
		resp := dispatch(req, encoder)
		mu.Unlock()

		// dispatch returns a zero Response{} for the "stop" case, where it
		// already wrote the reply and exited — but exiting happens inside,
		// so in practice we won't get back here after RequestStop.
		if req.Type != RequestStop {
			safeEncode(encoder, resp)
		}
	}
}

// parseRequest validates the raw JSON structurally and semantically before
// anything is allowed to touch exec.Command. This is the main "error proof"
// gate: unknown types, wrong shapes, and bad values are all caught here.
func parseRequest(raw json.RawMessage) (Request, error) {
	var req Request

	dec := json.NewDecoder(strings_NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return req, fmt.Errorf("invalid request shape: %v", err)
	}

	if strings.TrimSpace(req.Type) == "" {
		return req, errors.New("missing required field: type")
	}

	if !allowedTypes[req.Type] {
		return req, fmt.Errorf("unknown request type %q: expected one of compile, run, ping, stop", req.Type)
	}

	if req.Type == RequestCompile || req.Type == RequestRun {
		if _, ok := Languages[req.Language]; !ok {
			return req, fmt.Errorf("unsupported language %q", req.Language)
		}
	}

	if req.TimeLimitMS < 0 {
		return req, errors.New("time_limit must be non-negative")
	}
	if req.TimeLimitMS > int(MaxTimeout/time.Millisecond) {
		return req, fmt.Errorf("time_limit exceeds max allowed (%d ms)", MaxTimeout/time.Millisecond)
	}

	return req, nil
}

// small helper to avoid importing strings.NewReader under a name clash above
func strings_NewReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

// dispatch runs the actual handler with panic recovery, so a crash in one
// request (e.g. nil map access, unexpected exec failure) can never take down
// the persistent process or leave a hung connection.
func dispatch(req Request, encoder *json.Encoder) (resp Response) {
	defer func() {
		if r := recover(); r != nil {
			resp = Response{
				Success: false,
				Error:   fmt.Sprintf("internal error handling %q: %v", req.Type, r),
			}
		}
	}()

	switch req.Type {
	case RequestPing:
		return Response{Success: true, Output: "pong"}

	case RequestCompile:
		return compile(req)

	case RequestRun:
		return run(req)

	case RequestStop:
		safeEncode(encoder, Response{Success: true, Output: "stopping"})
		os.Exit(0)
		return Response{} // unreachable
	}

	// Should be unreachable due to parseRequest's whitelist, but keep a
	// final fallback so dispatch never returns a zero-value ambiguous resp.
	return Response{Success: false, Error: "unhandled request type"}
}

func safeEncode(encoder *json.Encoder, resp Response) {
	if err := encoder.Encode(resp); err != nil {
		// stdout pipe broken (e.g. parent process died) — nothing more we
		// can do but avoid panicking the runner.
		log.Printf("failed to write response: %v", err)
	}
}

// Compile

func compile(req Request) Response {
	lang := Languages[req.Language]
	if len(lang.Compile) == 0 {
		return Response{Success: true}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, lang.Compile[0], lang.Compile[1:]...)
	cmd.Dir = WorkDir
	cmd.Env = minimalEnv()

	out, err := runWithOutputCap(cmd)

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return Response{
			Success: false,
			Output:  out,
			Error:   "compilation timed out",
		}
	}

	if err != nil {
		if isBinaryMissing(err) {
			return Response{
				Success: false,
				Output:  out,
				Error:   fmt.Sprintf("compiler not available for %q: %v", req.Language, err),
			}
		}
		return Response{Success: false, Output: out, Error: err.Error()}
	}

	return Response{Success: true, Output: out}
}

func run(req Request) Response {
	lang := Languages[req.Language]
	if len(lang.Run) == 0 {
		return Response{Success: false, Error: fmt.Sprintf("no run command configured for %q", req.Language)}
	}

	timeout := DefaultTimeout
	if req.TimeLimitMS > 0 {
		timeout = time.Duration(req.TimeLimitMS) * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	memFile := filepath.Join(WorkDir, MemoryStatFile)
	_ = os.Remove(memFile) // best-effort cleanup of any stale stat file

	// Use /usr/bin/time to capture peak RSS; fall back gracefully if the
	// binary is missing from the sandbox image (common if the Dockerfile
	// slimmed the base image down).
	timeBinAvailable := binExists("/usr/bin/time")

	var cmd *exec.Cmd
	if timeBinAvailable {
		args := append([]string{"-f", "%M", "-o", memFile}, lang.Run...)
		cmd = exec.CommandContext(ctx, "/usr/bin/time", args...)
	} else {
		cmd = exec.CommandContext(ctx, lang.Run[0], lang.Run[1:]...)
	}

	cmd.Dir = WorkDir
	cmd.Env = minimalEnv()
	cmd.Stdin = strings.NewReader(req.Stdin)

	// Put the child in its own process group so a timeout kill also reaps
	// any grandchildren the sandboxed program spawns, instead of leaking
	// orphaned processes inside the container.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	start := time.Now()
	out, runErr := runWithOutputCap(cmd)
	runtime := time.Since(start)

	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
	if timedOut {
		killProcessGroup(cmd)
	}

	memoryKB := int64(0)
	if timeBinAvailable {
		if data, readErr := os.ReadFile(memFile); readErr == nil {
			memoryKB, _ = strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		}
	}

	resp := Response{
		Success: runErr == nil && !timedOut,
		Output:  out,
		Error:   errorString(runErr),
		Metadata: Metadata{
			RuntimeNS: runtime.Nanoseconds(),
			ExitCode:  exitCode(runErr),
			TimedOut:  timedOut,
			MemoryKB:  memoryKB,
		},
	}

	if timedOut {
		resp.Error = "time limit exceeded"
	} else if isBinaryMissing(runErr) {
		resp.Error = fmt.Sprintf("runtime not available for %q: %v", req.Language, runErr)
	}

	return resp
}

// runWithOutputCap runs cmd and returns combined stdout+stderr, truncated to
// MaxOutputBytes so a runaway program (e.g. `while True: print(...)`) can't
// exhaust host memory buffering output.
func runWithOutputCap(cmd *exec.Cmd) (string, error) {
	var buf capWriter
	buf.limit = MaxOutputBytes
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	out := buf.String()
	if buf.truncated {
		out += "\n[output truncated: exceeded limit]"
	}
	return out, err
}

type capWriter struct {
	data      []byte
	limit     int
	truncated bool
}

func (w *capWriter) Write(p []byte) (int, error) {
	if w.truncated {
		return len(p), nil // discard further writes but tell caller it "succeeded"
	}
	remaining := w.limit - len(w.data)
	if remaining <= 0 {
		w.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		w.data = append(w.data, p[:remaining]...)
		w.truncated = true
		return len(p), nil
	}
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *capWriter) String() string { return string(w.data) }

// killProcessGroup ensures the whole process tree started for a timed-out
// run is terminated, not just the direct child.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		return
	}
	_ = cmd.Process.Kill()
}

// minimalEnv strips the sandbox's environment down to the essentials so
// compiled/interpreted user code can't read host secrets or unrelated env
// vars that might have been set on the container.
func minimalEnv() []string {
	return []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=" + WorkDir,
		"LANG=C.UTF-8",
	}
}

func binExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isBinaryMissing(err error) bool {
	if err == nil {
		return false
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, os.ErrNotExist) || errors.Is(pathErr.Err, syscall.ENOENT)
	}
	return errors.Is(err, exec.ErrNotFound)
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
