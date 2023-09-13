package main

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/frida/frida-go/frida"
	"github.com/nsecho/furlzz/logger"
	"github.com/nsecho/furlzz/mutator"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed script/script.js
var scriptContent string

var rootCmd = &cobra.Command{
	Use:   "furlzz",
	Short: "Fuzz iOS URL schemes",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.NewLogger()

		var validInputs []string
		var err error

		base, err := cmd.Flags().GetString("base")
		if err != nil {
			return err
		}
		if base == "" {
			return errors.New("base URL cannot be empty")
		}

		input, err := cmd.Flags().GetString("input")
		if err != nil {
			return err
		}

		if input == "" && strings.Contains(base, "FUZZ") {
			return errors.New("input directory cannot be empty when using FUZZ keyword")
		}

		if strings.Contains(base, "FUZZ") {
			validInputs, err = readInputs(input)
			if err != nil {
				return err
			}
		}

		runs, err := cmd.Flags().GetUint("runs")
		if err != nil {
			return err
		}

		fn, err := cmd.Flags().GetString("function")
		if err != nil {
			return err
		}

		timeout, err := cmd.Flags().GetUint("timeout")
		if err != nil {
			return err
		}

		l.Infof("Fuzzing base URL \"%s\"", base)
		if strings.Contains(base, "FUZZ") {
			l.Infof("Read %d inputs from %s directory",
				len(validInputs), input)
		} else {
			l.Infof("Fuzzing base URL")
		}

		if runs == 0 {
			l.Infof("Fuzzing indefinitely")
		} else {
			l.Infof("Fuzzing with %d mutated inputs", runs)
		}

		if timeout != 0 {
			l.Infof("Sleeping %d seconds between each fuzz case", timeout)
		}

		app, err := cmd.Flags().GetString("app")
		if err != nil {
			return err
		}

		if app == "" {
			return errors.New("error: app cannot be empty")
		}

		dev := frida.USBDevice()
		if dev == nil {
			return errors.New("no USB device detected")
		}
		defer dev.Clean()

		sess, err := dev.Attach(app, nil)
		if err != nil {
			return err
		}

		sess.On("detached", func(reason frida.SessionDetachReason, crash *frida.Crash) {
			l.Infof("session detached; reason=%s", reason.String())
		})

		script, err := sess.CreateScript(scriptContent)
		if err != nil {
			return err
		}

		script.On("message", func(message string) {
			l.Infof("script output: %s", message)
		})

		if err := script.Load(); err != nil {
			return err
		}

		m := mutator.NewMutator(base, runs, fn, validInputs...)
		ch := m.Mutate()

		method, err := cmd.Flags().GetString("method")
		if err != nil {
			return err
		}

		for mutated := range ch {
			l.Infof("[%s] %s\n", color.New(color.FgCyan).Sprintf("%s", mutated.Mutation), mutated.Input)
			_ = script.ExportsCall("fuzz", method, mutated.Input)
			if timeout > 0 {
				time.Sleep(time.Duration(timeout) * time.Second)
			}
		}
		return nil
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func main() {
	rootCmd.Flags().StringP("app", "a", "Gadget", "Application name to attach to")
	rootCmd.Flags().StringP("base", "b", "", "base URL to fuzz")
	rootCmd.Flags().StringP("input", "i", "", "path to input directory")
	rootCmd.Flags().StringP("function", "f", "", "apply the function to mutated input (url, base64)")
	rootCmd.Flags().StringP("method", "m", "delegate", "method of opening url (delegate, app)")
	rootCmd.Flags().UintP("runs", "r", 0, "number of runs")
	rootCmd.Flags().UintP("timeout", "t", 1, "sleep X seconds between each case")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func readInputs(dirPath string) ([]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var validInputs []string

	for _, fl := range files {
		if fl.IsDir() {
			continue
		}
		data, err := func() ([]byte, error) {
			f, err := os.Open(filepath.Join(dirPath, fl.Name()))
			if err != nil {
				return nil, err
			}
			defer f.Close()

			data, _ := io.ReadAll(f)
			return data, nil
		}()
		if err != nil {
			return nil, err
		}
		validInputs = append(validInputs, string(data))
	}
	return validInputs, nil
}
