package cmd

import (
	"errors"
	"fmt"
	"github.com/freshautomations/stemplate/defaults"
	"github.com/freshautomations/stemplate/exit"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
	"text/template"
)

type FlagsType struct {
	File string
	String string
	List string
	Map string
	Output string
}

var flags FlagsType

func CheckArgs(cmd *cobra.Command, args []string) error {
	validateArgs := cobra.ExactArgs(1)
	if err := validateArgs(cmd, args); err != nil {
		return err
	}

	if flags.File == "" && flags.String == "" && flags.List == "" && flags.Map == "" {
		return errors.New("at least one of --file, --string, --list or --map is required")
	}

	_, err := os.Stat(args[0])
	if err != nil {
		return err
	}

	if flags.File != "" {
		_, err = os.Stat(flags.File)
	}

	return err
}

func RunRoot(cmd *cobra.Command, args []string) (output string, err error) {

	// Read file
	var dictionary map[string]interface{}
	if flags.File == "" {
		dictionary = make(map[string]interface{})
	} else {
		viper.SetConfigFile(flags.File)
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
	if flags.String != "" {
		for _, envVar := range strings.Split(flags.String, ",") {
			dictionary[envVar] = os.Getenv(envVar)
		}
	}

	// Read --list
	if flags.List != "" {
		for _, envVar := range strings.Split(flags.List, ",") {
			dictionary[envVar] = strings.Split(os.Getenv(envVar), ",")
		}
	}
	// Read --map
	if flags.Map != "" {
		for _, envVar := range strings.Split(flags.Map, ",") {
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

	// Read and parse template
	templateFile := args[0]
	var tmpl *template.Template
	tmpl, err = template.ParseFiles(templateFile)
	if err != nil {
		return
	}

	// Execute template and print results
	out := os.Stdout
	if flags.Output != "" {
		f, fErr := os.Create(flags.Output)
		defer f.Close()
		if fErr != nil {
			return
		}
		out = f
	}
	err = tmpl.Execute(out, dictionary)

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
	pflag.StringVarP(&flags.Output, "output", "o", "", "Send results to this file instead of stdout")
	pflag.StringVarP(&flags.File, "file", "f", "", "Filename that contains data structure")
	pflag.StringVarP(&flags.String, "string", "s", "", "Comma-separated list of environment variable names that contain strings")
	pflag.StringVarP(&flags.List, "list", "l", "", "Comma-separated list of environment variable names that contain comma-separated strings")
	pflag.StringVarP(&flags.Map, "map", "m", "", "Comma-separated list of environment variable names that contain comma-separated strings of key=value pairs")
	_ = rootCmd.MarkFlagFilename("file")

	return rootCmd.Execute()
}
