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

	if inputFlags.File == "" && inputFlags.String == "" && inputFlags.List == "" && inputFlags.Map == "" {
		return errors.New("at least one of --file, --string, --list or --map is required")
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

var dictionary map[string]interface{}

func substitute(name string) interface{} {
	return dictionary[name]
}

func counter(input interface{}) interface{} {
	var num uint64
	if numf64, ok := input.(float64); ok {
		num = uint64(numf64)
	} else {
		if numi, ok := input.(uint64); ok {
			num = numi
		} else {
			if nums, ok := input.(string); ok {
				var err error
				num, err = strconv.ParseUint(nums, 10, 64)
				if err != nil {
					panic(err)
				}
			} else {
				panic(errors.New(fmt.Sprintf("cannot convert input to number: %s", input)))
			}
		}
	}
	var result []uint64
	var i uint64
	for i = 0 ; i < num ; i++ {
		result = append(result, i)
	}
	return result
}

func RunRoot(cmd *cobra.Command, args []string) (output string, err error) {

	// Read file
	if inputFlags.File == "" {
		dictionary = make(map[string]interface{})
	} else {
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
		dictionary = viper.AllSettings()
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
			err = os.Mkdir(inputFlags.Output, os.ModePerm)
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
					return os.Mkdir(destination, pathInfo.Mode())
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
	_ = rootCmd.MarkFlagFilename("file")

	return rootCmd.Execute()
}
