package cmd

import (
	"bytes"
	"github.com/freshautomations/stemplate/defaults"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

const testresult = `Hi guest!

Welcome to this test template demonstration.

You should see a few examples of
* List item: first
* Map item: testmap
* Golang specific stuff
`

type stdoutRedirect struct {
	r, w, old *os.File
}

// How to redirect stdout: https://stackoverflow.com/questions/10473800/in-go-how-do-i-capture-stdout-of-a-function-into-a-string
func redirectStdOut() (out stdoutRedirect) {
	out.old = os.Stdout // keep backup of the real stdout
	out.r, out.w, _ = os.Pipe()
	os.Stdout = out.w
	return
}

func readAndResetStdOut(in stdoutRedirect) string {
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, in.r)
		outC <- buf.String()
	}()

	// back to normal state
	in.w.Close()
	os.Stdout = in.old
	return <-outC
}

func TestCheckArgs(t *testing.T) {
	cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version: defaults.Version,
	}

	assert.NotNil(t, CheckArgs(cmd, []string{"../test.json"}), "enough parameters")
	assert.NotNil(t, CheckArgs(cmd, []string{"notexist.json", "../test.template"}), "file found")
	assert.NotNil(t, CheckArgs(cmd, []string{"../test.json", "../notexist.template"}), "file found")
	assert.Nil(t, CheckArgs(cmd, []string{"../test.json", "../test.template"}), "parameter check")
}

func TestRunRoot(t *testing.T) {
	cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version: defaults.Version,
	}

	var stdO stdoutRedirect
	var err error

	stdO = redirectStdOut()
	_, err = RunRoot(cmd,[]string{"../test.json", "../test.template"})
	assert.Equal(t, testresult, readAndResetStdOut(stdO), "unexpected result")
	assert.Nil(t, err, "unexpected error")

	stdO = redirectStdOut()
	_, err = RunRoot(cmd,[]string{"../test.toml", "../test.template"})
	assert.Equal(t, testresult, readAndResetStdOut(stdO), "unexpected result")
	assert.Nil(t, err, "unexpected error")

	stdO = redirectStdOut()
	_, err = RunRoot(cmd,[]string{"../test.yaml", "../test.template"})
	assert.Equal(t, testresult, readAndResetStdOut(stdO), "unexpected result")
	assert.Nil(t, err, "unexpected error")

}
