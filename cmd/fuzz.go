package cmd

import (
	"errors"
	"github.com/fatih/color"
	"github.com/frida/frida-go/frida"
	"github.com/nsecho/furlzz/mutator"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var fuzzCmd = &cobra.Command{
	Use:   "fuzz",
	Short: "Fuzz URL scheme",
	RunE: func(cmd *cobra.Command, args []string) error {
		var validInputs [][]byte
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

		method, err := cmd.Flags().GetString("method")
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

		l.Infof("Attached to %s", app)

		var lastInput []byte

		sess.On("detached", func(reason frida.SessionDetachReason, crash *frida.Crash) {
			l.Infof("Session detached; reason=%s", reason.String())
			out := crashSHA256(lastInput)
			err := func() error {
				f, err := os.Create(out)
				if err != nil {
					return err
				}
				f.Write(lastInput)
				return nil
			}()
			if err != nil {
				l.Errorf("Error writing crash file: %v", err)
			} else {
				l.Infof("Written crash to: %s", out)
			}
			s := Session{
				App:      app,
				Base:     base,
				Function: fn,
				Method:   method,
			}
			if err := s.WriteToFile(); err != nil {
				l.Errorf("Error writing session file: %v", err)
			} else {
				l.Infof("Written session file")
			}
			os.Exit(1)
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

		l.Infof("Loaded script")

		m := mutator.NewMutator([]byte(base), runs, fn, validInputs...)
		ch := m.Mutate()

		for mutated := range ch {
			//lastInput = mutated.Input
			l.Infof("[%s] %s\n", color.New(color.FgCyan).Sprintf("%s", mutated.Mutation), mutated.Input)
			//_ = script.ExportsCall("fuzz", method, mutated.Input)
			if timeout > 0 {
				time.Sleep(time.Duration(timeout) * time.Second)
			}
		}
		return nil
	},
}

func init() {
	fuzzCmd.Flags().StringP("app", "a", "Gadget", "Application name to attach to")
	fuzzCmd.Flags().StringP("base", "b", "", "base URL to fuzz")
	fuzzCmd.Flags().StringP("input", "i", "", "path to input directory")
	fuzzCmd.Flags().StringP("function", "f", "", "apply the function to mutated input (url, base64)")
	fuzzCmd.Flags().StringP("method", "m", "delegate", "method of opening url (delegate, app)")
	fuzzCmd.Flags().UintP("runs", "r", 0, "number of runs")
	fuzzCmd.Flags().UintP("timeout", "t", 1, "sleep X seconds between each case")

	rootCmd.AddCommand(fuzzCmd)
}
