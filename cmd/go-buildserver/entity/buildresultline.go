package entity

import (
	"time"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/process"
)

type BuildResultLine struct {
	Line string
	pipe process.PipeType
	Pipe string
	Time time.Time
}

func NewBuildResultLine(line string, pipe process.PipeType, moment time.Time) BuildResultLine {
	pipeLineTypeMapping := map[process.PipeType]string{
		process.STDERR: "STDERR",
		process.STDOUT: "STDOUT",
	}
	return BuildResultLine{
		Line: line,
		pipe: pipe,
		Pipe: pipeLineTypeMapping[pipe],
		Time: moment,
	}
}

func NewBuildResultLinePipeString(line string, pipe string, moment time.Time) BuildResultLine {
	pipeLineTypeMapping := map[string]process.PipeType{
		"STDERR": process.STDERR,
		"STDOUT": process.STDOUT,
	}
	return BuildResultLine{
		Line: line,
		pipe: pipeLineTypeMapping[pipe],
		Pipe: pipe,
		Time: moment,
	}
}
