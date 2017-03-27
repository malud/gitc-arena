package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"sync"
)

var botCount = 0

type AI struct {
	id           int
	name         string
	processState *os.ProcessState
	inPipe       io.WriteCloser
	outPipe      io.Reader
	m            sync.Mutex
}

func NewAI() *AI {
	cmd := exec.Command(*aiFlag)
	pr, err := cmd.StdinPipe()
	if err != nil {
		println(err)
	}
	pw, err := cmd.StdoutPipe()
	if err != nil {
		println(err)
	}
	cmd.Stderr = os.Stderr
	botCount++

	if err := cmd.Start(); err != nil {
		println(err)
		os.Exit(1)
	}

	ai := &AI{
		id:      botCount,
		name:    *aiFlag,
		inPipe:  pr,
		outPipe: pw,
	}

	go func() {
		state, err := cmd.Process.Wait()
		if err != nil {
			println(err)
		}
		ai.m.Lock()
		ai.processState = state
		ai.m.Unlock()
	}()

	return ai
}

func (a *AI) Feed_Inputs(b *bytes.Buffer) {
	_, err := io.Copy(a.inPipe, b)
	if err != nil {
		println("Err:", err.Error())
	}
}

func (a *AI) Alive() bool {
	a.m.Lock()
	defer a.m.Unlock()
	if a.processState != nil {
		return !a.processState.Exited()
	}
	return true
}
