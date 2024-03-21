package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frida/frida-go/frida"
	"github.com/nsecho/furlzz/internal/tui"
	"github.com/nsecho/furlzz/mutator"
	"github.com/spf13/cobra"
)

var fuzzCmd = &cobra.Command{
	Use:   "fuzz",
	Short: "Fuzz URL scheme",
	RunE: func(cmd *cobra.Command, args []string) error {
		var validInputs []string
		var err error

		base, err := cmd.Flags().GetString("base")
		if err != nil {
			return err
		}

		input, err := cmd.Flags().GetString("input")
		if err != nil {
			return err
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

		delegate, err := cmd.Flags().GetString("delegate")
		if err != nil {
			return err
		}

		uiapp, err := cmd.Flags().GetString("uiapp")
		if err != nil {
			return err
		}

		scene, err := cmd.Flags().GetString("scene")
		if err != nil {
			return err
		}

		crash, err := cmd.Flags().GetBool("crash")
		if err != nil {
			return err
		}

		app, err := cmd.Flags().GetString("app")
		if err != nil {
			return err
		}

		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		m := tui.NewModel()
		m.Crash = crash
		m.Runs = runs
		m.Timeout = timeout
		m.App = app
		m.Device = "usb"
		m.Function = fn
		m.Method = method
		m.Delegate = delegate
		m.UIApp = uiapp
		m.Scene = scene
		m.Base = base
		m.Input = input
		m.ValidInputs = validInputs

		p := tea.NewProgram(m)

		var sess *frida.Session = nil
		var script *frida.Script = nil
		hasCrashed := false

		go func() {
			if base == "" {
				sendErr(p, "Base cannot be empty")
				return
			}

			if input == "" && strings.Contains(base, "FUZZ") {
				sendErr(p, "Input directory cannot be empty")
				return
			}

			if app == "" {
				sendErr(p, "App cannot be empty")
				return
			}

			//Adding support for accessing remote devices, else default is USB
			if network != "" {
				mgr := frida.NewDeviceManager()
				ropts := frida.NewRemoteDeviceOptions()
				dev, err := mgr.AddRemoteDevice(network, ropts)
				if err != nil {
					sendErr(p, err.Error())
					return
				}
				defer dev.Clean()

				//Spawn app only if not in foreground
				spawnApp(dev, app, p, false)
				sess, err = dev.Attach(app, nil)
				if err != nil {
					sendErr(p, err.Error())
					return
				}
				sendStats(p, "Attached to Remote device")
			} else {
				dev := frida.USBDevice()
				if dev == nil {
					sendErr(p, "No USB device detected")
					return
				}
				defer dev.Clean()

				//Spawn app only if not in foreground
				spawnApp(dev, app, p, false)
				sess, err = dev.Attach(app, nil)
				if err != nil {
					sendErr(p, err.Error())
					return
				}
				sendStats(p, "Attached to USB device")
			}

			sendStats(p, fmt.Sprintf("Reading inputs from %s", input))
			sendStats(p, fmt.Sprintf("Attached to %s", app))

			var lastInput string

			sess.On("detached", func(reason frida.SessionDetachReason, crash *frida.Crash) {
				// Add sleep here so that we can wait for the context to get cancelled
				time.Sleep(3 * time.Second)
				defer p.Send(tui.SessionDetached{})
				if hasCrashed {
					sendStats(p, fmt.Sprintf("Session detached; reason=%s", reason.String()))
					out := fmt.Sprintf("fcrash_%s_%s", app, crashSHA256(lastInput))
					err := func() error {
						f, err := os.Create(out)
						if err != nil {
							return err
						}
						_, err = f.WriteString(lastInput)
						return err
					}()
					if err != nil {
						sendErr(p, fmt.Sprintf("Could not write crash file: %s", err.Error()))
					} else {
						sendStats(p, fmt.Sprintf("Written crash to: %s", out))
					}
					s := Session{
						App:      app,
						Base:     base,
						Delegate: delegate,
						Function: fn,
						Method:   method,
						Scene:    scene,
						UIApp:    uiapp,
					}
					if err := s.WriteToFile(); err != nil {
						sendErr(p, fmt.Sprintf("Could not write session file: %s", err.Error()))
					} else {
						sendStats(p, "Written session file")
					}
				}
			})

			script, err = sess.CreateScript(scriptContent)
			if err != nil {
				sendErr(p, fmt.Sprintf("Could not create script: %s", err.Error()))
				return
			}

			script.On("message", func(message string) {
				sendStats(p, fmt.Sprintf("Script output: %s", message))
			})

			if err := script.Load(); err != nil {
				sendErr(p, fmt.Sprintf("Could not load script: %s", err.Error()))
				return
			}

			sendStats(p, "Script loaded")

			_ = script.ExportsCall("setup_fuzz", method, uiapp, delegate, scene)
			sendStats(p, "Finished fuzz setup")

			mut := mutator.NewMutator(base, app, runs, fn, crash, validInputs...)
			ch := mut.Mutate()

			for mutated := range ch {
				lastInput = mutated.Input
				p.Send(tui.MutatedMsg(mutated))
				ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
				if err := script.ExportsCallWithContext(ctx, "fuzz", method, mutated.Input); err == frida.ErrContextCancelled {
					hasCrashed = true
					break
				}
				if timeout > 0 {
					time.Sleep(time.Duration(timeout) * time.Second)
				}
			}
		}()

		_, err = p.Run()
		return err
	},
}

func sendStats(p *tea.Program, msg string) {
	p.Send(tui.StatsMsg(msg))
}

func sendErr(p *tea.Program, msg string) {
	p.Send(tui.ErrMsg(msg))
}

func spawnApp(dev *frida.Device, app string, p *tea.Program, toSpawn bool) {
	process, err := dev.FindProcessByName(app, frida.ScopeMinimal)
	if err != nil {
		sendErr(p, err.Error())
		return
	}
	//If app is not open, Spawn it
	if process.PID() < 0 {
		toSpawn = true
	} else if process.PID() > 0 {
		//If app is in process but not in foreground, Spawn it
		// sometimes crash makes app go in the background
		frontApp, err := dev.FrontmostApplication(frida.ScopeMinimal)
		if err != nil {
			sendErr(p, err.Error())
			//We don't need to exit/return here, since frida throws generic error if no app is in foreground
		}
		if frontApp == nil || frontApp.Name() != process.Name() {
			//need to kill app in foreground, else fuzzing seems to stop for some reason
			dev.Kill(process.PID())
			toSpawn = true
		}
	}

	if toSpawn == true {
		fopts := frida.NewSpawnOptions()
		fopts.SetArgv([]string{
			"",
		})
		appsList, err := dev.EnumerateApplications("", frida.ScopeMinimal)
		if err != nil {
			return
		}

		for i := 0; i < int(len(appsList)); i++ {
			appName := appsList[i]
			if appName.Name() == app {
				pid, err := dev.Spawn(appName.Identifier(), fopts)
				if err != nil {
					sendErr(p, err.Error())
					return
				}
				dev.Resume(pid)
				break
			}
		}
	}
}

func init() {
	fuzzCmd.Flags().StringP("app", "a", "Gadget", "Application name to attach to")
	fuzzCmd.Flags().StringP("base", "b", "", "base URL to fuzz")
	fuzzCmd.Flags().StringP("input", "i", "", "path to input directory")
	fuzzCmd.Flags().StringP("function", "f", "", "apply the function to mutated input (url, base64)")
	fuzzCmd.Flags().StringP("method", "m", "delegate", "method of opening url (delegate, app)")
	fuzzCmd.Flags().StringP("delegate", "d", "", "UISceneDelegate class name")
	fuzzCmd.Flags().StringP("uiapp", "u", "", "UIApplication class name")
	fuzzCmd.Flags().StringP("scene", "s", "", "UIScene class name")
	fuzzCmd.Flags().BoolP("crash", "c", false, "ignore previous crashes")
	fuzzCmd.Flags().UintP("runs", "r", 0, "number of runs")
	fuzzCmd.Flags().UintP("timeout", "t", 1, "sleep X seconds between each case")
	fuzzCmd.Flags().StringP("network", "n", "", "Connect to Device Remotely")

	rootCmd.AddCommand(fuzzCmd)
}
