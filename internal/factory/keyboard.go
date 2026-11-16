package factory

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// keyboard switches the terminal to read single key presses, without
// Enter and without echo. It returns nil keys when stdin is not a
// terminal. restore puts the terminal back as it was.
func keyboard() (keys <-chan byte, restore func()) {
	saved, err := stty("-g")
	if err != nil {
		return nil, func() {}
	}
	if _, err := stty("-icanon", "-echo", "min", "1"); err != nil {
		return nil, func() {}
	}
	restore = func() { stty(strings.TrimSpace(saved)) }

	ch := make(chan byte, 8)
	go func() {
		b := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(b)
			if err != nil {
				return
			}
			if n == 1 {
				ch <- b[0]
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		restore()
		fmt.Println()
		os.Exit(130)
	}()
	return ch, restore
}

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	return string(out), err
}
