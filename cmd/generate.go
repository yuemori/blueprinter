/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuemori/blueprinter/internal/logger"
	"github.com/yuemori/blueprinter/internal/runner"
)

var (
	verbose                              bool
	template, workdir, glob, ignore, out string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate <path/to/package> <container struct name>",
	Short: "Generate DI container code",
	Long:  "Generate DI container code",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			logger.SetVerbose(true)
		}

		var b bytes.Buffer

		packagePath := args[0]
		structName := args[1]

		globs := []string{}
		if glob != "" {
			globs = strings.Split(glob, ",")
		}

		ignores := []string{}
		if ignore != "" {
			ignores = strings.Split(ignore, ",")
		}

		t := runner.DefaultTemplate

		if template != "" {
			bytes, err := os.ReadFile(template)
			if err != nil {
				log.Fatal(err)
			}

			t = string(bytes)
		}

		cfg := &runner.Config{
			Template:         t,
			Dest:             &b,
			WorkDir:          workdir,
			Globs:            globs,
			Ignores:          ignores,
			ContainerName:    structName,
			ContainerPackage: packagePath,
		}

		errs := runner.Run(cfg)

		if errs != nil {
			for _, err := range errs {
				fmt.Println(err)
			}

			os.Exit(1)
		}

		dest := os.Stdout

		if out != "" {
			fp, err := os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0o664)
			if err != nil {
				log.Fatal(err)
			}
			dest = fp
			defer dest.Close()
		}

		if _, err := dest.Write(b.Bytes()); err != nil {
			log.Fatal(err)
		}

		if out != "" {
			logger.Infof("Generated code is written to %s", out)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringVarP(&template, "template", "t", "", "Template file for generating code. If not speicied, use default template")
	generateCmd.PersistentFlags().StringVarP(&workdir, "workdir", "w", ".", "Workdir for generating code. If not specified, use current directory")
	generateCmd.PersistentFlags().StringVarP(&ignore, "ignore", "i", "", "Glob pattern for ignoring files")
	generateCmd.PersistentFlags().StringVarP(&out, "out", "o", "", "Output file for generated code. If not specified, output to stdout")
	generateCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")
}
