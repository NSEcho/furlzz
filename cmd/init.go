package cmd

import (
	"encoding/json"
	"github.com/nsecho/furlzz/internal/config"
	"github.com/spf13/cobra"
	"os"
)

const (
	defaultType string = "application"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new furlzz project",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		tp, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		var dt any

		switch tp {
		case "delegate":
			dt = config.DelegateType{
				Delegate:    "AppDelegate",
				Application: "UIApplication",
			}
		case "scene_activity":
			dt = config.SceneActivityType{
				SceneDelegate: "AppSceneDelegate",
				SceneName:     "AppScene",
			}
		case "scene_context":
			dt = config.SceneActivityType{
				SceneDelegate: "AppSceneDelegate",
				SceneName:     "AppScene",
			}
		case "delegate_activity":
			dt = config.DelegateActivityType{
				Delegate:    "AppDelegate",
				Application: "UIApplication",
			}
		default:
			dt = config.ApplicationType{
				Application: "UIApplication",
			}
		}

		cfg := config.Config{
			Application:   "ExampleApp",
			Base:          "application://example?a=FUZZ",
			IgnoreCrashes: false,
			Inputs: []string{
				"input_1",
			},
			RemoteDevice: "",
			Runs:         0,
			SpawnTimeout: 0,
			Timeout:      1,
			WorkingDir:   ".",
			Function:     "url",
			Type:         defaultType,
			Fuzz:         dt,
		}

		f, err := os.Create(c)
		if err != nil {
			return err
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")

		if err := enc.Encode(cfg); err != nil {
			return err
		}

		l.Infof("Initialized new furlzz project @ %s", c)

		return nil
	},
}

func init() {
	initCmd.Flags().StringP("config", "c", "furlzz.json", "Path to config file furlzz.json")
	initCmd.Flags().StringP("type", "t", defaultType, "Type of URL method to fuzz")
	rootCmd.AddCommand(initCmd)
}
