package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/eldarion-gondor/piper"
	"github.com/flynn/flynn/pkg/attempt"
	"github.com/tj/go-spin"
	"golang.org/x/crypto/ssh/terminal"
)

func remoteExec(endpoint string, enableTty bool) (int, error) {
	done := make(chan struct{}, 1)
	var showIndicator bool
	var outs io.Writer
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		outs = os.Stdout
		showIndicator = true
	} else if terminal.IsTerminal(int(os.Stderr.Fd())) {
		outs = os.Stderr
		showIndicator = true
	}
	if showIndicator {
		s := spin.New()
		s.Set(spin.Box1)
		go func() {
		loop:
			for {
				select {
				case <-done:
					break loop
				case <-time.After(100 * time.Millisecond):
					fmt.Fprintf(outs, "\r\033[36mAttaching...\033[m %s ", s.Next())
				}
			}
		}()
	}
	// wait for ok to report 200
	if err := (attempt.Strategy{
		Total: 2 * time.Minute,
		Delay: 1 * time.Second,
	}.Run(func() error {
		okURL := "https://" + endpoint + "/ok"
		resp, err := http.Get(okURL)
		if err != nil {
			return err
		}
		if resp.StatusCode == 200 {
			return nil
		}
		return errors.New("non-200 response")
	})); err != nil {
		done <- struct{}{}
		if showIndicator {
			fmt.Fprintf(outs, "\r\033[36mAttaching...\033[m failed\n")
		}
		return 1, err
	}
	return func() int {
		opts := piper.Opts{}
		if enableTty {
			if terminal.IsTerminal(int(os.Stdin.Fd())) {
				w, h, err := terminal.GetSize(int(os.Stdin.Fd()))
				if err != nil {
					fatal(err.Error())
				}
				state, err := terminal.MakeRaw(int(os.Stdin.Fd()))
				if err != nil {
					fatal(err.Error())
				}
				defer terminal.Restore(int(os.Stdin.Fd()), state)
				opts.Tty = true
				opts.Width = w
				opts.Height = h
			}
		}
		var pipe *piper.Pipe
		if err := (attempt.Strategy{
			Total: 10 * time.Second,
			Delay: 1 * time.Second,
		}.Run(func() error {
			var err error
			pipe, err = piper.NewClientPipe(endpoint, opts, nil)
			if err != nil {
				return err
			}
			return nil
		})); err != nil {
			if showIndicator {
				fmt.Fprintf(outs, "\r\033[36mAttaching...\033[m error\n")
			}
			failure(err.Error())
			return 1
		}
		done <- struct{}{}
		if showIndicator {
			fmt.Fprintf(outs, "\r\033[36mAttaching...\033[m ok\n")
		}
		exitCode, err := pipe.Interact()
		if err != nil {
			failure(err.Error())
			return 1
		}
		return exitCode
	}(), nil
}
