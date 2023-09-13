package cmd

import (
	"crypto/sha256"
	"fmt"
	"github.com/nsecho/furlzz/logger"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

var scriptContent string

var l = *logger.NewLogger()

var rootCmd = &cobra.Command{
	Use:   "furlzz",
	Short: "Fuzz iOS URL schemes",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute(sc string) error {
	scriptContent = sc
	return rootCmd.Execute()
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

func crashSHA256(inp string) string {
	h := sha256.New()
	h.Write([]byte(inp))
	return fmt.Sprintf("%x", h.Sum(nil))
}
