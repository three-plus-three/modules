package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/runner-mei/schd_job"
)

// Job is an interface for submitted cron jobs.
type Command struct {
	Execute   string
	Arguments []string
	Dir       string

	Logfile      string
	Environments []string
}

const maxBytes = 5 * 1024 * 1024
const maxNum = 5

func (job *Command) rotateLogFile() error {
	st, err := os.Stat(job.Logfile)
	if nil != err { // file exists
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if st.Size() < maxBytes {
		return nil
	}

	fname2 := job.Logfile + fmt.Sprintf(".%04d", maxNum)
	_, err = os.Stat(fname2)
	if nil == err {
		err = os.Remove(fname2)
		if err != nil {
			return err
		}
	}

	fname1 := fname2
	for num := maxNum - 1; num > 0; num-- {
		fname2 = fname1
		fname1 = job.Logfile + fmt.Sprintf(".%04d", num)

		_, err = os.Stat(fname1)
		if nil != err {
			continue
		}
		err = os.Rename(fname1, fname2)
		if err != nil {
			return err
		}
	}

	err = os.Rename(job.Logfile, fname1)
	if err != nil {
		return err
	}
	return nil
}

func (job *Command) Run(ctx context.Context) error {
	if e := job.rotateLogFile(); e != nil {
		return errors.New("rotate log file(" + job.Logfile + ") failed, " + e.Error())
	}
	out, e := os.OpenFile(job.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if nil != e {
		job.Logfile = strings.Replace(job.Logfile, "/", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, "\\", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, "*", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, ":", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, "\"", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, "|", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, "?", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, ">", "_", -1)
		job.Logfile = strings.Replace(job.Logfile, "<", "_", -1)

		out, e = os.OpenFile(job.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if nil != e {
			return errors.New("open log file(" + job.Logfile + ") failed, " + e.Error())
		}
	}
	defer out.Close()

	io.WriteString(out, "=============== begin ===============\r\n")
	defer io.WriteString(out, "===============  end  ===============\r\n")

	execPath := job.Execute
	if s := schd_job.Commands[job.Execute]; s != "" {
		s = execPath
	}

	executePath, found := schd_job.LookPath(schd_job.ExecutableFolder, execPath)
	if !found {
		executePath = job.Execute
	}

	cmd := exec.CommandContext(ctx, executePath, job.Arguments...)
	cmd.Stderr = out
	cmd.Stdout = out
	cmd.Dir = job.Dir

	if len(job.Environments) > 0 {
		osEnv := os.Environ()
		var environments []string
		environments = make([]string, 0, len(job.Environments)+len(osEnv))
		environments = append(environments, osEnv...)
		environments = append(environments, job.Environments...)
		cmd.Env = environments
	}

	io.WriteString(out, cmd.Path)
	for idx, s := range cmd.Args {
		if 0 == idx {
			continue
		}
		io.WriteString(out, "\r\n \t\t")
		io.WriteString(out, s)
	}
	io.WriteString(out, "\r\n===============  out  ===============\r\n")
	if e = cmd.Start(); nil != e {
		io.WriteString(out, "start failed, "+e.Error()+"\r\n")
		return nil
	}

	if e = cmd.Wait(); nil != e {
		io.WriteString(out, "run failed, "+e.Error()+"\r\n")
	} else if nil != cmd.ProcessState {
		io.WriteString(out, "run ok, exit with "+cmd.ProcessState.String()+".\r\n")
	}
	return nil
}
