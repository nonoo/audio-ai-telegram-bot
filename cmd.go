package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/danielgatis/go-vte/vtparser"
)

// https://stackoverflow.com/questions/71714228/go-exec-commandcontext-is-not-being-terminated-after-context-timeout

type Cmd struct {
	ctx context.Context
	*exec.Cmd
}

// NewCommand is like exec.CommandContext but ensures that subprocesses
// are killed when the context times out, not just the top level process.
func NewCommand(ctx context.Context, command string, args ...string) *Cmd {
	return &Cmd{ctx, exec.Command(command, args...)}
}

func (c *Cmd) Start() error {
	// Force-enable setpgid bit so that we can kill child processes when the
	// context times out or is canceled.
	if c.Cmd.SysProcAttr == nil {
		c.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.Cmd.SysProcAttr.Setpgid = true
	err := c.Cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		<-c.ctx.Done()
		p := c.Cmd.Process
		if p == nil {
			return
		}
		// Kill by negative PID to kill the process group, which includes
		// the top-level process we spawned as well as any subprocesses
		// it spawned.
		_ = syscall.Kill(-p.Pid, syscall.SIGKILL)
	}()
	return nil
}

func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

type lineParser struct {
	line     string
	lineChan chan string
}

func (p *lineParser) Print(r rune) {
	p.line += string(r)
}

func (p *lineParser) Execute(b byte) {
	if b == 0x0d || b == 0x0a {
		p.lineChan <- p.line
		p.line = ""
	}
}

func (c *Cmd) ParseOutput(doneChan chan bool, lineChan chan string, errChan chan error) error {
	cmdReader, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	c.Stderr = c.Stdout
	go func() {
		lineParser := &lineParser{lineChan: lineChan}

		// Dummy callbacks.
		putCallback := func(b byte) {}
		unhookCallback := func() {}
		hookCallback := func(params []int64, intermediates []byte, ignore bool, r rune) {}
		oscCallback := func(params [][]byte, bellTerminated bool) {}
		csiCallback := func(params []int64, intermediates []byte, ignore bool, r rune) {}
		escCallback := func(intermediates []byte, ignore bool, b byte) {}

		parser := vtparser.New(lineParser.Print, lineParser.Execute, putCallback, unhookCallback, hookCallback, oscCallback, csiCallback, escCallback)
		buf := make([]byte, 1)
		var err error
	cycle:
		for {
			select {
			case <-c.ctx.Done():
				break cycle
			default:
				_, err = cmdReader.Read(buf)
				if err != nil {
					if err == io.EOF { // We consider EOF as not an error.
						err = nil
					}
					break cycle
				}
				parser.Advance(buf[0])
			}
		}
		errChan <- err
		doneChan <- true
	}()
	return nil
}

func (c *Cmd) RunAndProcessOutput(processLineCallback func(string)) (canceled bool, err error) {
	doneChan := make(chan bool)
	defer close(doneChan)
	lineChan := make(chan string)
	defer close(lineChan)
	errChan := make(chan error)
	defer close(errChan)
	err = c.ParseOutput(doneChan, lineChan, errChan)
	if err != nil {
		return false, fmt.Errorf("parse error: %w", err)
	}
	err = c.Start()
	if err != nil {
		return false, fmt.Errorf("start error: %w", err)
	}

selectForLoop:
	for {
		select {
		case <-c.ctx.Done():
			<-errChan
			<-doneChan
			return true, nil
		case line := <-lineChan:
			processLineCallback(line)
		case err = <-errChan:
			break selectForLoop
		}
	}
	reqQueue.currentEntry.entry.cancelProcessUpdate()
	<-doneChan
	return false, err
}
