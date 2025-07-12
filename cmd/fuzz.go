package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/frida/frida-go/frida"
	"github.com/spf13/cobra"

	"github.com/nsecho/furlzz/internal/config"
	"github.com/nsecho/furlzz/mutator"
)

var (
	fuzzRe   = regexp.MustCompile("FUZZ[1-5]")
	numberRe = regexp.MustCompile("[1-5]")
)

var fuzzCmd = &cobra.Command{
	Use:   "fuzz",
	Short: "Fuzz URL scheme",
	RunE: func(cmd *cobra.Command, args []string) error {
		var validInputs []string
		var err error

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		var cfg config.Config
		f, err := os.Open(configPath)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			return err
		}

		if cfg.Base == "" {
			return errors.New("missing Base URL")
		}

		matches := fuzzRe.FindAllString(cfg.Base, -1)

		if len(matches) == 0 {
			return errors.New("missing FUZZ keywords")
		}

		uniqueFuzz := make(map[string]struct{})
		for _, match := range matches {
			uniqueFuzz[match] = struct{}{}
		}

		var unique []string
		for k := range uniqueFuzz {
			unique = append(unique, k)
		}

		sort.Slice(unique, func(i, j int) bool {
			firstNum := numberRe.FindStringSubmatch(unique[i])
			secondNum := numberRe.FindStringSubmatch(unique[j])
			first, _ := strconv.Atoi(firstNum[0])
			second, _ := strconv.Atoi(secondNum[0])
			return first < second
		})

		if len(cfg.Inputs) != len(uniqueFuzz) {
			return errors.New("invalid number of inputs; one for each FUZZ keywords")
		}

		inputSets := make(map[string][]string)

		for i := range unique {
			validInputs, err = readInputs(cfg.Inputs[i])
			if err != nil {
				return err
			}
			inputSets[unique[i]] = validInputs
		}

		if cfg.Application == "" {
			return errors.New("missing application name")
		}

		mut := mutator.NewMutator(cfg.Base, cfg.Application, cfg.Runs, cfg.Function, cfg.IgnoreCrashes, inputSets)

		var session *frida.Session = nil
		var script *frida.Script = nil

		l.Infof("Frida version: %s", frida.Version())

		// Adding support for accessing remote devices, else default is USB
		if cfg.RemoteDevice != "" {
			mgr := frida.NewDeviceManager()
			ropts := frida.NewRemoteDeviceOptions()
			dev, err := mgr.AddRemoteDevice(cfg.RemoteDevice, ropts)
			if err != nil {
				return err
			}
			defer dev.Clean()

			session, err = dev.Attach(cfg.Application, nil)
			if err != nil {
				return err
			}
			l.Infof("Attached to remote device: %s", cfg.RemoteDevice)
		} else {
			dev := frida.USBDevice()
			if dev == nil {
				return errors.New("missing USB device")
			}
			defer dev.Clean()

			// Spawn app only if not in foreground
			if err := spawnApp(dev, cfg.Application, false, cfg.SpawnTimeout); err != nil {
				return err
			}
			session, err = dev.Attach(cfg.Application, nil)
			if err != nil {
				return err
			}
			l.Infof("Attached to %s", cfg.Application)
		}

		l.Infof("Reading inputs from %s", strings.Join(cfg.Inputs, ","))

		var lastInput string
		detached := make(chan struct{})

		session.On("detached", func(reason frida.SessionDetachReason, crash *frida.Crash) {
			detached <- struct{}{}
			out := fmt.Sprintf("fcrash_%s_%s", cfg.Application, crashSHA256(lastInput))
			err := func() error {
				f, err := os.Create(filepath.Join(cfg.WorkingDir, out))
				if err != nil {
					return err
				}
				_, err = f.WriteString(lastInput)
				return err
			}()
			if err != nil {
				l.Errorf("Could not write crash file: %v", err.Error())
			} else {
				l.Infof("Written crash to: %s", out)
			}
		})

		script, err = session.CreateScript(scriptContent)
		if err != nil {
			return err
		}

		script.On("message", func(message string) {
			l.Infof("Script output: %s", message)
		})

		if err := script.Load(); err != nil {
			return err
		}

		l.Infof("Script loaded")

		method := cfg.Type
		uiapp := ""
		delegateName := ""
		sceneName := ""

		fuzzMap := cfg.Fuzz.(map[string]any)

		switch method {
		case "delegate":
			fallthrough
		case "delegate_activity":
			uiapp = fuzzMap["application"].(string)
			delegateName = fuzzMap["delegate"].(string)
		case "app":
			uiapp = fuzzMap["application"].(string)
		case "scene_activity":
			fallthrough
		case "scene_context":
			sceneName = fuzzMap["scene"].(string)
			delegateName = fuzzMap["delegate"].(string)
		}

		_ = script.ExportsCall("setup_fuzz", method, uiapp, delegateName, sceneName)

		l.Infof("Finished fuzz setup")

		ch := mut.Mutate()

	mLoop:
		for {
			select {
			case <-detached:
				mut.Close()
				break mLoop
			case mutated := <-ch:
				lastInput = mutated.Input
				l.Infof("[%s] %s\n", color.New(color.FgCyan).Sprintf("%s", mutated.Mutation), mutated.Input)

				_ = script.ExportsCall("fuzz", cfg.Type, mutated.Input)

				if cfg.Timeout > 0 {
					time.Sleep(time.Duration(cfg.Timeout) * time.Second)
				}

				// Check if script has new coverage blocks
				has, ok := script.ExportsCall("has_new_blocks").(bool)
				if ok && has {
					l.Infof("New blocks found, continuing fuzzing...")
					mut.HandleNewCoverage(mutated.MutatedInputs)
				}
			}
		}

		return err
	},
}

func spawnApp(dev frida.DeviceInt, app string, toSpawn bool, sTimeout uint) error {
	process, err := dev.FindProcessByName(app, frida.ScopeMinimal)
	if err != nil {
		return err
	}
	// If app is not open, Spawn it
	if process.PID() < 0 {
		toSpawn = true
	} else if process.PID() > 0 {
		// If app is in process but not in foreground, Spawn it
		frontApp, err := dev.FrontmostApplication(frida.ScopeMinimal)
		if err != nil {
			// We don't need to exit/return here, since frida throws generic error if no app is in foreground sending as stats
			return err
		}
		// Checking if foreground app does not match intended app, then we spawn it
		if frontApp == nil || frontApp.Name() != process.Name() {
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
			return err
		}

		for i := 0; i < int(len(appsList)); i++ {
			appName := appsList[i]
			if appName.Name() == app {
				pid, err := dev.Spawn(appName.Identifier(), fopts)
				if err != nil {
					return err
				}
				err = dev.Resume(pid)
				if err != nil {
					return err
				}
				break
			}
		}
		l.Infof("Spawning app: %s", app)
		// Sleep for supplied time before fuzzing so app spawn properly
		if sTimeout > 0 {
			time.Sleep(time.Duration(sTimeout) * time.Second)
		}
	}
	return nil
}

func init() {
	fuzzCmd.Flags().StringP("config", "c", "furlzz.json", "Path to config file")

	rootCmd.AddCommand(fuzzCmd)
}
