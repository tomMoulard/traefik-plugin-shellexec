package pluginshellexec_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pluginshellexec "github.com/tommoulard/traefik-plugin-shellexec"
)

func TestShellexec(t *testing.T) {
	cfg := pluginshellexec.CreateConfig()

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := pluginshellexec.New(ctx, next, cfg, "shellexec-plugin")
	require.NoError(t, err)
	testCases := []struct {
		command    string
		name       string
		returnCode string
		stderr     string
		stdin      string
		stdout     string
	}{
		{
			name:   "empty",
			stderr: "fork/exec : no such file or directory",
		},
		{
			name:       "simple command",
			command:    "true",
			returnCode: "0",
		},
		{
			name:       "command with return code != 0",
			command:    "/bin/bash -c 'exit 1'",
			returnCode: "1",
			stderr:     "exit status 1",
		},
		{
			name:       "command stdout",
			command:    "echo etc",
			returnCode: "0",
			stdout:     "etc",
		},
		{
			name:       "command stderr",
			command:    "ls /not/an/actual/file",
			returnCode: "2",
			stderr:     "exit status 2",
		},
		{
			name:       "command stdin",
			command:    "tr 'a-z' 'A-Z'",
			returnCode: "0",
			stdin:      "qwerty",
			stdout:     "QWERTY",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()

			values := map[string]string{}
			if test.stdin != "" {
				values["stdin"] = test.stdin
			}

			if test.command != "" {
				values["command"] = test.command
			}

			jsonData, err := json.Marshal(values)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", bytes.NewBuffer(jsonData))
			require.NoError(t, err)

			handler.ServeHTTP(recorder, req)

			var res map[string]interface{}
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)

			t.Logf("test: %+v\nres: %+v", test, res)

			assert.Regexp(t, regexp.MustCompile(test.returnCode), res["return_code"])
			assertstd(t, test.stdout, res["stdout"])
			assertstd(t, test.stderr, res["stderr"])
		})
	}
}

func assertstd(t *testing.T, expect string, got interface{}) {
	t.Helper()

	if expect == "" {
		assert.Equal(t, "", got)
	}
	assert.Regexp(t, regexp.MustCompile(expect), got)
}
