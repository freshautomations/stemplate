package cmd

import (
	"github.com/freshautomations/stemplate/defaults"
	"github.com/freshautomations/stemplate/exit"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"text/template"
)

// --output <filename>
var outputFlag string

func CheckArgs(cmd *cobra.Command, args []string) error {
	validateArgs := cobra.ExactArgs(2)
	if err := validateArgs(cmd, args); err != nil {
		return err
	}

	dict := args[0]
	_, err := os.Stat(dict)
	if err == nil {
		templ := args[1]
		_, err = os.Stat(templ)
	}
	return err
}

func RunRoot(cmd *cobra.Command, args []string) (output string, err error) {
	// Get input
	dictionary := args[0]
	templateFile := args[1]

	// Read dictionary
	viper.SetConfigFile(dictionary)
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

	// Read and parse template
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return
	}

	// Execute template and print results
	out := os.Stdout
	if outputFlag != "" {
		f, fErr := os.Create(outputFlag)
		defer f.Close()
		if fErr != nil {
			return
		}
		out = f
	}
	err = tmpl.Execute(out, viper.AllSettings())

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
	rootCmd.Use = "stemplate <dictionary> <template>"
	pflag.StringVarP(&outputFlag, "output", "o", "", "Send results to this file instead of stdout")

	return rootCmd.Execute()
}
