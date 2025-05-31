package main

import (
	_ "embed"
	"fmt"
	"github.com/frida/frida-go/frida"
	"github.com/nsecho/furlzz/cmd"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	agentFilename = "_agent.js"
)

var sc string

//go:embed script.ts
var scriptContent []byte

//go:embed package.json
var packageJSON []byte

//go:embed package-lock.json
var packageLockJSON []byte

var tempFiles = map[string][]byte{
	"script.ts":         scriptContent,
	"package.json":      packageJSON,
	"package-lock.json": packageLockJSON,
}

func main() {
	tempDir := filepath.Join(os.TempDir(), "furlzz")
	os.MkdirAll(tempDir, os.ModePerm)

	// we don't have agent compiled
	if _, err := os.Stat(filepath.Join(tempDir, agentFilename)); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(tempDir, "script.ts")); os.IsNotExist(err) {
			for fl, data := range tempFiles {
				os.WriteFile(filepath.Join(tempDir, fl), data, os.ModePerm)
			}
		}
		if _, err = os.Stat(filepath.Join(tempDir, "node_modules")); os.IsNotExist(err) {
			// Install modules
			pwd, _ := os.Getwd()
			os.Chdir(tempDir)
			command := exec.Command("npm", "install")
			if err := command.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to install node modules: %v\n", err)
				os.Exit(1)
			}
			os.Chdir(pwd)

			c := frida.NewCompiler()
			c.On("diagnostics", func(diag string) {
				fmt.Fprintf(os.Stderr, "Diagnostics: %s\n", diag)
				os.Exit(1)
			})

			bopts := frida.NewCompilerOptions()
			bopts.SetProjectRoot(tempDir)
			bopts.SetJSCompression(frida.JSCompressionTerser)

			bundle, err := c.Build("script.ts", bopts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to compile script: %v\n", err)
				os.Exit(1)
			}
			os.WriteFile(filepath.Join(tempDir, agentFilename), []byte(bundle), os.ModePerm)
			sc = bundle
		}
	} else {
		scriptContent, err = ioutil.ReadFile(filepath.Join(tempDir, agentFilename))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read agent: %v\n", err)
			os.Exit(1)
		}
		sc = string(scriptContent)
	}

	if err := cmd.Execute(sc); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
