package cmd

import (
	"errors"
	"github.com/frida/frida-go/frida"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

var sess *frida.Session = nil
var crashCmd = &cobra.Command{
	Use:   "crash",
	Short: "Run the application with crash",
	RunE: func(cmd *cobra.Command, args []string) error {
		sfile, err := cmd.Flags().GetString("session")
		if err != nil {
			return err
		}

		l.Infof("Reading session file %s", sfile)
		s, err := NewSession(sfile)
		if err != nil {
			return err
		}

		crash, err := cmd.Flags().GetString("crash")
		if err != nil {
			return err
		}

		inpf, err := os.Open(crash)
		if err != nil {
			return err
		}
		defer inpf.Close()

		input, _ := io.ReadAll(inpf)

		l.Infof("Read %s from crash file", string(input))

		//Adding support for accessing remote devices, else default is USB
		if s.NetworkDevice != "" {
			mgr := frida.NewDeviceManager()
			ropts := frida.NewRemoteDeviceOptions()
			dev, err := mgr.AddRemoteDevice(s.NetworkDevice, ropts)
			if err != nil {
				return err
			}
			defer dev.Clean()

			sess, err = dev.Attach(s.App, nil)
			if err != nil {
				return err
			}
		} else {
			dev := frida.USBDevice()
			if dev == nil {
				return errors.New("no USB device detected")
			}
			defer dev.Clean()

			sess, err = dev.Attach(s.App, nil)
			if err != nil {
				return err
			}
		}

		defer sess.Clean()

		l.Infof("Attached to %s", s.App)

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
		defer script.Clean()

		l.Infof("Loaded script")

		l.Infof("Sleeping for two seconds and triggering crash")

		time.Sleep(2 * time.Second)

		_ = script.ExportsCall("setup_fuzz", s.Method, s.UIApp, s.Delegate, s.Scene)

		_ = script.ExportsCall("fuzz", s.Method, string(input))

		return nil
	},
}

func init() {
	crashCmd.Flags().StringP("session", "s", "", "session path")
	crashCmd.Flags().StringP("crash", "c", "", "crash file")

	rootCmd.AddCommand(crashCmd)
}
