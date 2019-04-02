package cmd

import (
	"github.com/freshautomations/stemplate/defaults"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var testresult = `Hi guest!

Welcome to this test template demonstration.

You should see a few examples of
* List item: first
* Map item: testmap
* Golang specific stuff
`

func TestCheckArgs(t *testing.T) {
	cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version: defaults.Version,
	}

	assert.NotNil(t, CheckArgs(cmd, []string{"../test.template", "../test.json"}), "enough parameters")
	assert.NotNil(t, CheckArgs(cmd, []string{"notexist.json"}), "file found")
	assert.NotNil(t, CheckArgs(cmd, []string{"../test.template"}), "file found")
	flags.File = "../test.json"
	assert.Nil(t, CheckArgs(cmd, []string{"../test.template"}), "parameter check")
}

func TestRunRoot(t *testing.T) {
	cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version: defaults.Version,
	}

	var resultfile []byte
	var err error

	flags.Output = "jsonresult.tmp"
	flags.File = "../test.json"
	_, err = RunRoot(cmd,[]string{"../test.template"})
	resultfile, err = ioutil.ReadFile(flags.Output)
	assert.Equal(t, string(resultfile), testresult, "unexpected result")
	assert.Nil(t, err, "unexpected error")
	_ = os.Remove(flags.Output)

	flags.Output = "tomlresult.tmp"
	flags.File = "../test.toml"
	_, err = RunRoot(cmd,[]string{"../test.template"})
	resultfile, err = ioutil.ReadFile(flags.Output)
	assert.Equal(t, string(resultfile), testresult, "unexpected result")
	assert.Nil(t, err, "unexpected error")
	_ = os.Remove(flags.Output)

	flags.Output = "yamlresult.tmp"
	flags.File = "../test.yaml"
	_, err = RunRoot(cmd,[]string{"../test.template"})
	assert.Equal(t, string(resultfile), testresult, "unexpected result")
	assert.Nil(t, err, "unexpected error")
	_ = os.Remove(flags.Output)

	flags.Output = "emvresult.tmp"
	flags.File = ""
	flags.String = "user,filename"
	flags.List = "list,gospecific"
	flags.Map = "map"
	_ = os.Setenv("user","guest")
	_ = os.Setenv("filename","test")
	_ = os.Setenv("list","first,second,third")
	_ = os.Setenv("gospecific","Go,lang")
	_ = os.Setenv("map","test=testmap,nottest='not a testmap'")
	_, err = RunRoot(cmd,[]string{"../test.template"})
	assert.Equal(t, string(resultfile), testresult, "unexpected result")
	assert.Nil(t, err, "unexpected error")
	_ = os.Remove(flags.Output)

}
