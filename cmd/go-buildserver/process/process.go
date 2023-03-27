package process

import (
	"bufio"
	"io"
	"os/exec"
	"time"
)

type PipeType uint8

const (
	STDOUT PipeType = iota
	STDERR          = iota
)

func pipeYielder(pipe io.ReadCloser, callback func(time.Time, string)) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		m := scanner.Text()
		callback(time.Now(), m)
	}
}

func StartProcessWithStdErrStdOutCallback(cmd string, args []string, callback func(PipeType, time.Time, string)) {
	process := exec.Command(cmd, args...)

	stdout, _ := process.StdoutPipe()
	stderr, _ := process.StderrPipe()
	process.Start()
	go pipeYielder(stdout, func(t time.Time, s string) {
		callback(STDOUT, t, s)
	})
	go pipeYielder(stderr, func(t time.Time, s string) {
		callback(STDERR, t, s)
	})

	process.Wait()

}
