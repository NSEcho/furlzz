package config

// DelegateType responds to config for -[AppDelegate application:openURL:options:]
type DelegateType struct {
	Delegate    string `json:"delegate"`
	Application string `json:"application"`
}

type ApplicationType struct {
	Application string `json:"application"`
}

// SceneActivityType is used when the application is using either of these two methods:
// 1. -[UISceneDelegate scene:continueUserActivity]
// 2. -[UISceneDelegate scene:openURLContexts:]
type SceneActivityType struct {
	SceneDelegate string `json:"delegate"`
	SceneName     string `json:"scene"`
}

type DelegateActivityType struct {
	Delegate    string `json:"delegate"`
	Application string `json:"application"`
}

type Config struct {
	Application   string   `json:"app"`
	Base          string   `json:"base"`
	IgnoreCrashes bool     `json:"ignore_crashes"`
	Inputs        []string `json:"inputs"`
	RemoteDevice  string   `json:"remote_device"`
	Runs          uint     `json:"runs"`
	SpawnTimeout  uint     `json:"spawn_timeout"`
	Timeout       uint     `json:"timeout"`
	WorkingDir    string   `json:"working_dir"`
	Function      string   `json:"function"`
	Type          string   `json:"type"`
	Fuzz          any      `json:"fuzz"`
}
