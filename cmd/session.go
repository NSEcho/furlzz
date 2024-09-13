package cmd

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

func NewSession(sessionFile string) (*Session, error) {
	var s Session

	f, err := os.Open(sessionFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

type Session struct {
	App           string `yaml:"app"`
	Base          string `yaml:"base"`
	Delegate      string `yaml:"delegate"`
	Function      string `yaml:"fn"`
	Method        string `yaml:"method"`
	NetworkDevice string `yaml:"networkDevice"`
	Scene         string `yaml:"scene"`
	UIApp         string `yaml:"uiapp"`
}

func (s *Session) WriteToFile(wd string) error {
	t := time.Now()
	outputFilename := fmt.Sprintf("session_%s", t.Format("2006_01_02_15:04:05"))

	f, err := os.Create(filepath.Join(wd, outputFilename))
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(&s)
}
