package cgo

//#include <stdio.h>
import "C"
import (
	"bufio"
	"os"
	"sync"
	"syscall"
)

var redirectMu sync.Mutex

type RedirectedOutput struct {
	Err  error
	Read []byte
}

func CaptureCGOOutput(call func() error) (outC chan *RedirectedOutput) {
	outC = make(chan *RedirectedOutput)

	stdout, err := syscall.Dup(syscall.Stdout)
	if err != nil {
		outC <- &RedirectedOutput{Err: err}
		return
	}

	stderr, err := syscall.Dup(syscall.Stderr)
	if err != nil {
		outC <- &RedirectedOutput{Err: err}
		return
	}

	r, w, err := os.Pipe()
	if err != nil {
		outC <- &RedirectedOutput{Err: err}
		return
	}

	go func() {
		redirectMu.Lock()
		defer redirectMu.Unlock()
		defer func() {
			C.fflush(nil)
			if err := syscall.Dup2(stdout, syscall.Stdout); err != nil {
				outC <- &RedirectedOutput{Err: err}
				return
			}
			if err := syscall.Close(stdout); err != nil {
				outC <- &RedirectedOutput{Err: err}
				return
			}
			if err := syscall.Dup2(stderr, syscall.Stderr); err != nil {
				outC <- &RedirectedOutput{Err: err}
				return
			}
			if err := syscall.Close(stderr); err != nil {
				outC <- &RedirectedOutput{Err: err}
				return
			}
			if err := w.Close(); err != nil {
				outC <- &RedirectedOutput{Err: err}
			}
			if err := r.Close(); err != nil {
				outC <- &RedirectedOutput{Err: err}
			}
		}()

		if err := syscall.Dup2(int(w.Fd()), syscall.Stdout); err != nil {
			outC <- &RedirectedOutput{Err: err}
			return
		}

		if err := syscall.Dup2(int(w.Fd()), syscall.Stderr); err != nil {
			outC <- &RedirectedOutput{Err: err}
			return
		}

		if err := call(); err != nil {
			outC <- &RedirectedOutput{Err: err}
		}

	}()

	go func() {
		br := bufio.NewScanner(r)
		br.Split(bufio.ScanLines)
		for br.Scan() {
			outC <- &RedirectedOutput{Read: br.Bytes()}
		}
		close(outC)
	}()

	return
}
