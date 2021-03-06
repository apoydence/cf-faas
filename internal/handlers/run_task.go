package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	faas "github.com/poy/cf-faas"
)

type RunTask struct {
	command         string
	expectedHeaders []string
	rawOutput       bool
	r               TaskRunner
	log             *log.Logger
}

type TaskRunner interface {
	RunTask(command, name string) (string, error)
}

type TaskRunnerFunc func(command, name string) (string, error)

func (f TaskRunnerFunc) RunTask(command, name string) (string, error) {
	return f(command, name)
}

func NewRunTask(
	command string,
	expectedHeaders []string,
	rawOutput bool,
	r TaskRunner,
	log *log.Logger,
) faas.Handler {
	return &RunTask{
		command:         command,
		expectedHeaders: expectedHeaders,
		rawOutput:       rawOutput,
		r:               r,
		log:             log,
	}
}

func (r *RunTask) Handle(req faas.Request) (faas.Response, error) {
	name := r.encodeTaskName(req)
	result, err := r.r.RunTask(r.command, name)
	if err != nil {
		return faas.Response{}, err
	}

	output := []byte(result)
	if !r.rawOutput {
		output = []byte(fmt.Sprintf(`{"task_guid":%q,"task_name":%q}`, result, name))
	}

	return faas.Response{
		StatusCode: http.StatusOK,
		Body:       output,
	}, nil
}

func (r *RunTask) encodeTaskName(req faas.Request) string {
	header := http.Header{}
	for _, h := range r.expectedHeaders {
		h = strings.Title(h)
		header[h] = req.Header[h]
	}

	req.Header = header
	data, err := json.Marshal(req)
	if err != nil {
		r.log.Panic(err)
	}

	return base64.StdEncoding.EncodeToString(data)
}
