package cmd

import (
	"errors"
	"fmt"
	"github.com/freshautomations/stemplate/defaults"
	"github.com/freshautomations/stemplate/exit"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type FlagsType struct {
	Env       bool
	File      string
	String    string
	List      string
	Map       string
	Output    string
	Extension string
	All       bool
}

var inputFlags FlagsType

func CheckArgs(cmd *cobra.Command, args []string) (err error) {
	validateArgs := cobra.ExactArgs(1)
	if err = validateArgs(cmd, args); err != nil {
		return
	}

	if inputFlags.File == "" && inputFlags.String == "" && inputFlags.List == "" && inputFlags.Map == "" && ! inputFlags.Env {
		return errors.New("at least one of --file, --string, --list --env or --map is required")
	}

	for _, item := range strings.Split(args[0], ",") {
		_, err = os.Stat(item)
		if err != nil {
			return
		}
	}

	if inputFlags.File != "" {
		_, err = os.Stat(inputFlags.File)
	}

	return err
}

func interface2uint64(input interface{}) (uint64, error) {
	// Might be the right type already
	if xnum, ok := input.(uint64); ok {
		return xnum, nil
	}
	// JSON represents numbers as float64
	if xnum, ok := input.(float64); ok {
		return uint64(xnum), nil
	}
	// YAML represents numbers as int
	if xnum, ok := input.(int); ok {
		return uint64(xnum), nil
	}
	// TOML represents numbers as int64
	if xnum, ok := input.(int64); ok {
		return uint64(xnum), nil
	}
	// Some users might quote their numbers
	if xnum, ok := input.(string); ok {
		num, err := strconv.ParseUint(xnum, 10, 64)
			return num, err
	}
	return 0, errors.New(fmt.Sprintf("cannot convert input to number: %s", input))
}

var dictionary map[string]interface{}

func substitute(name string) interface{} {
	return dictionary[name]
}

func counter(input interface{}) (result []uint64, err error) {
	var num uint64
	num, err = interface2uint64(input)
	var i uint64
	for i = 0 ; i < num ; i++ {
		result = append(result, i)
	}
	return
}

func left(s string, input interface{}) (string, error) {
	i, err := interface2uint64(input)
	if err != nil {
		return "", err
	}
	return s[0:i], err
}

func right(s string, input interface{}) (string, error) {
	i, err := interface2uint64(input)
	if err != nil {
		return "", err
	}
	return s[len(s)-int(i):], err
}

func mid(s string, inputb interface{}, inputl interface{}) (string, error) {
	b, err := interface2uint64(inputb)
	if err != nil {
		return "", err
	}
	l, err := interface2uint64(inputl)
	if err != nil {
		return "", err
	}
	return s[b:b+l], err
}

func add(a interface{}, b interface{}) (result uint64, err error) {
	var ax, bx uint64
	ax, err = interface2uint64(a)
	if err != nil {
		return
	}
	bx, err = interface2uint64(b)
	if err != nil {
		return
	}
	result = ax + bx
	return
}

func sub(a interface{}, b interface{}) (result uint64, err error) {
	var ax, bx uint64
	ax, err = interface2uint64(a)
	if err != nil {
		return
	}
	bx, err = interface2uint64(b)
	if err != nil {
		return
	}
	result = ax - bx
	return
}

func RunRoot(cmd *cobra.Command, args []string) (output string, err error) {
// Priorities least to most: env, file, string, list, map

	dictionary = make(map[string]interface{})

	// Read --env
	if inputFlags.Env {
		for _, envVar := range os.Environ() {
			equals := strings.Index(envVar,"=")
			if equals < 1 || equals==len(envVar) {
				// Invalid string
				continue
			}
			name := envVar[0:equals]
			value := envVar[equals+1:]
			dictionary[name] = value
		}
	}

	// Read --file
	if inputFlags.File != "" {
		viper.SetConfigFile(inputFlags.File)
		err = viper.ReadInConfig()
		if err != nil {
			if _, IsUnsupportedExtension := err.(viper.UnsupportedConfigError); IsUnsupportedExtension {
				viper.SetConfigType("toml")
				err = viper.ReadInConfig()
				if err != nil {
					return
				}
			} else {
				return
			}
		}
		for k, v := range viper.AllSettings() {
			if dictionary[k] == nil {
				dictionary[k] = v
			}
		}
	}

	// Read --string
	if inputFlags.String != "" {
		for _, envVar := range strings.Split(inputFlags.String, ",") {
			dictionary[envVar] = os.Getenv(envVar)
		}
	}

	// Read --list
	if inputFlags.List != "" {
		for _, envVar := range strings.Split(inputFlags.List, ",") {
			dictionary[envVar] = strings.Split(os.Getenv(envVar), ",")
		}
	}
	// Read --map
	if inputFlags.Map != "" {
		for _, envVar := range strings.Split(inputFlags.Map, ",") {
			tempMap := make(map[string]string)
			for _, mapItem := range strings.Split(os.Getenv(envVar), ",") {
				m := strings.Split(mapItem, "=")
				if len(m) < 2 {
					// something's not right, there's no equal sign (=) in the variable
					return "", errors.New(fmt.Sprintf("Missing =. %s does not contain a map: %s", envVar, mapItem))
				} else {
					tempMap[m[0]] = strings.Join(m[1:], "=")
				}
			}
			dictionary[envVar] = tempMap
		}
	}

	// Read and parse template files and directories
	var tmpl *template.Template
	funcMaps := template.FuncMap{
		"substitute": substitute,
		"counter": counter,
		"left": left,
		"right": right,
		"mid": mid,
		"add": add,
		"sub": sub,
	}

	// Input template
	templateInput := args[0]
	templateIsComplex := true // Assuming we have a list of files and directories
	templateIsDir := false
	if templateInfo, checkErr := os.Stat(templateInput); checkErr == nil {
		templateIsDir = templateInfo.IsDir()
		templateIsComplex = false
	}

	// Output path
	outputIsDir := false
	if inputFlags.Output != "" {
		outputInfo, checkErr := os.Stat(inputFlags.Output)
		outputExist := checkErr == nil
		if outputExist {
			outputIsDir = outputInfo.IsDir()
		}

		if (templateIsComplex || templateIsDir) && !outputExist {
			err = os.MkdirAll(inputFlags.Output, os.ModePerm)
			if err != nil {
				return
			}
		}
		if (templateIsComplex || templateIsDir) && outputExist && !outputIsDir {
			err = errors.New("cannot copy template folder into file")
			return
		}
	}

	for _, templateFileOrDir := range strings.Split(templateInput, ",") {
		err = filepath.Walk(templateFileOrDir, func(currentPath string, pathInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			var destination string
			out := os.Stdout
			if inputFlags.Output != "" {
				// (file-to-file) source is a simple file, destination is a folder or a file
				if !templateIsComplex && !templateIsDir {
					if pathInfo.IsDir() { // source is under multiple folders
						return nil
					}
					if outputIsDir {
						destination = filepath.Join(inputFlags.Output, filepath.Base(currentPath))
					} else {
						destination = inputFlags.Output
					}
				}
				// (dir-to-dir) source is one directory, use the contents only
				if !templateIsComplex && templateIsDir {
					relativeRoot := filepath.Clean(templateFileOrDir)
					cleanCurrentPath := filepath.Clean(currentPath)
					if currentPath == templateFileOrDir || relativeRoot == cleanCurrentPath { // do not copy the source's root folder
						return nil
					}
					relativePath := filepath.Clean(strings.Replace(cleanCurrentPath, relativeRoot, "", 1))
					destination = filepath.Join(inputFlags.Output, relativePath)
				}
				// (multi-to-dir) source is a list of files and directories, copy source folders too
				if templateIsComplex {
					destination = filepath.Join(inputFlags.Output, currentPath)
				}
				// if the current path is a directory, create it at output (should only run when multi|dir-to-dir)
				if pathInfo.IsDir() {
					return os.MkdirAll(destination, pathInfo.Mode())
				}
				// If extension does not match and we do not process all files in the template directory, then copy file and move on
				if (templateIsComplex || templateIsDir) && !inputFlags.All && filepath.Ext(destination) != inputFlags.Extension {
					return os.Link(currentPath, destination)
				}
				// Cut off .template extension
				extension := filepath.Ext(destination)
				if filepath.Ext(destination) == inputFlags.Extension {
					destination = destination[0 : len(destination)-len(extension)]
				}
				// Create and open file
				out, err = os.Create(destination)
				defer out.Close()
				if err != nil {
					return err
				}
			} else { // Print to screen instead of file
				// If the current path is a directory, move on
				if pathInfo.IsDir() {
					return nil
				}
				// If extension does not match and we do not process all files in the template directory, then print the file and move on
				if (templateIsComplex || templateIsDir) && !inputFlags.All && filepath.Ext(currentPath) != inputFlags.Extension {
					regularFileContent, openError := ioutil.ReadFile(currentPath)
					_,_ = fmt.Fprint(out, regularFileContent)
					return openError
				}
			}

			// Prepare template reading
			tmpl, err = template.New(filepath.Base(currentPath)).Funcs(funcMaps).ParseFiles(currentPath)
			if err != nil {
				return err
			}

			// Execute template and print results to destination output
			return tmpl.Execute(out, dictionary)
		})
		if err != nil {
			return
		}
	}

	return
}

func runRootWrapper(cmd *cobra.Command, args []string) {
	if result, err := RunRoot(cmd, args); err != nil {
		exit.Fail(err)
	} else {
		exit.Succeed(result)
	}
}

func Execute() error {
	var rootCmd = &cobra.Command{
		Version: defaults.Version,
		Use:     "stemplate",
		Short:   "STemplate - simple template parser for Shell",
		Long: `A simple template parser for the Linux Shell.
Source and documentation is available at https://github.com/freshautomations/stemplate`,
		Args: CheckArgs,
		Run:  runRootWrapper,
	}
	rootCmd.Use = "stemplate <template>"
	pflag.StringVarP(&inputFlags.Output, "output", "o", "", "Send results to this file instead of stdout")
	pflag.StringVarP(&inputFlags.File, "file", "f", "", "Filename that contains data structure")
	pflag.StringVarP(&inputFlags.String, "string", "s", "", "Comma-separated list of environment variable names that contain strings")
	pflag.StringVarP(&inputFlags.List, "list", "l", "", "Comma-separated list of environment variable names that contain comma-separated strings")
	pflag.StringVarP(&inputFlags.Map, "map", "m", "", "Comma-separated list of environment variable names that contain comma-separated strings of key=value pairs")
	pflag.StringVarP(&inputFlags.Extension, "extension", "t", ".template", "Extension for template files when template input or output is a directory. Default: .template")
	pflag.BoolVarP(&inputFlags.All, "all", "a", false, "Consider all files in a directory templates, regardless of extension.")
	pflag.BoolVarP(&inputFlags.Env, "env", "e", false, "Import all environment variables for templates as strings.")
	_ = rootCmd.MarkFlagFilename("file")

	return rootCmd.Execute()
}
