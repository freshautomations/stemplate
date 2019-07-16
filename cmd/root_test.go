package cmd

import (
	"github.com/freshautomations/stemplate/defaults"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var testresult = `Hi guest!

Welcome to this test template demonstration.

You should see a few examples of
* List item: first
* Map item: testmap
* Golang specific stuff
`

var testcustomfunctionsresult = `Hi,

Welcome to custom functions demonstration.

* substitute variable: testmap
* count 0-4 using counter: 0 1 2 3 4
* addition 3 + 5: 8
* substraction 8 - 2: 6
* left "abcdefg" 3: abc
* right "abcdefg" 3: efg
* string cut the last char from "abcdefg": abcdef
* mid "abcdefg" 3 2: de
`

// rootDir has to be ".." for CircleCI to work correctly.
var rootDir = ".."

func TestCheckArgs(t *testing.T) {
	cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version: defaults.Version,
	}

	assert.NotNil(t, CheckArgs(cmd, []string{filepath.Join(rootDir, "test_templates", " test.template"), filepath.Join("..", "test.json")}), "enough parameters")
	assert.NotNil(t, CheckArgs(cmd, []string{"notexist.json"}), "file found")
	assert.NotNil(t, CheckArgs(cmd, []string{filepath.Join(rootDir, "test_templates", "test.template")}), "file found")
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries", "test.json")
	assert.Nil(t, CheckArgs(cmd, []string{filepath.Join(rootDir, "test_templates", "test.template")}), "parameter check")
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

	inputFlags.Extension = ".template"

	// JSON test
	inputFlags.Output = "jsonresult.tmp"
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.json")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates", "test.template")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(inputFlags.Output)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, testresult, string(resultfile), "unexpected result")
	_ = os.Remove(inputFlags.Output)

	// TOML test
	inputFlags.Output = "tomlresult.tmp"
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.toml")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates", "test.template")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(inputFlags.Output)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, testresult, string(resultfile), "unexpected result")
	_ = os.Remove(inputFlags.Output)

	// YAML test
	inputFlags.Output = "yamlresult.tmp"
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.yaml")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates","test.template")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(inputFlags.Output)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, testresult, string(resultfile), "unexpected result")
	_ = os.Remove(inputFlags.Output)

	// Environment variables test
	inputFlags.Output = "emvresult.tmp"
	inputFlags.File = ""
	inputFlags.String = "user,filename"
	inputFlags.List = "list,gospecific"
	inputFlags.Map = "map"
	_ = os.Setenv("user", "guest")
	_ = os.Setenv("filename", "test")
	_ = os.Setenv("list", "first,second,third")
	_ = os.Setenv("gospecific", "Go,lang")
	_ = os.Setenv("map", "test=testmap,nottest='not a testmap'")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates","test.template")})
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, testresult, string(resultfile), "unexpected result")
	_ = os.Remove(inputFlags.Output)
	inputFlags.String = ""
	inputFlags.List = ""
	inputFlags.Map = ""

	// Output is a directory test
	inputFlags.Output = filepath.Join(rootDir, "outputdir1")
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.json")
	err = os.Mkdir(inputFlags.Output, os.ModePerm)
	assert.Nil(t, err, "unexpected error during folder creation")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates","test.template")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(filepath.Join(inputFlags.Output, "test"))
	assert.Equal(t, testresult, string(resultfile), "unexpected result")
	assert.Nil(t, err, "unexpected error")
	_ = os.Remove(filepath.Join(inputFlags.Output, "test"))
	_ = os.Remove(inputFlags.Output)

	// Input template is a folder, output is a directory test
	inputFlags.Output = filepath.Join(rootDir, "outputdir2")
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.json")
	err = os.Mkdir(inputFlags.Output, os.ModePerm)
	assert.Nil(t, err, "unexpected error during folder creation")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(filepath.Join(inputFlags.Output, "test"))
	assert.Equal(t, testresult, string(resultfile), "unexpected result")
	assert.Nil(t, err, "unexpected error")
	_ = os.Remove(filepath.Join(inputFlags.Output, "test"))
	_ = os.Remove(inputFlags.Output)

}

func TestCustomFunctions(t *testing.T) {
	cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version: defaults.Version,
	}

	var resultfile []byte
	var err error

	inputFlags.Extension = ".template"

	// JSON test
	inputFlags.Output = filepath.Join(rootDir, "customfunctions_jsonresult.tmp")
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.json")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates2","customfunctions.template")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(inputFlags.Output)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, testcustomfunctionsresult, string(resultfile), "unexpected result")
	_ = os.Remove(inputFlags.Output)

	// TOML test
	inputFlags.Output = filepath.Join(rootDir, "customfunctions_tomlresult.tmp")
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.toml")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates2","customfunctions.template")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(inputFlags.Output)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, testcustomfunctionsresult, string(resultfile), "unexpected result")
	_ = os.Remove(inputFlags.Output)

	// YAML test
	inputFlags.Output = filepath.Join(rootDir, "customfunctions_yamlresult.tmp")
	inputFlags.File = filepath.Join(rootDir, "test_dictionaries","test.yaml")
	_, err = RunRoot(cmd, []string{filepath.Join(rootDir, "test_templates2","customfunctions.template")})
	assert.Nil(t, err, "unexpected error")
	resultfile, err = ioutil.ReadFile(inputFlags.Output)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, testcustomfunctionsresult, string(resultfile), "unexpected result")
	_ = os.Remove(inputFlags.Output)

}
