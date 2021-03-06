package main

import (
	"encoding/json"

	"code.cloudfoundry.org/go-envstruct"
)

type Config struct {
	// Command is the command that will run on the local task or within the
	// new one.
	Command string `env:"COMMAND, required"`

	// CreateTask defines if the task is a new one or the current task.
	CreateTask bool `env:CREATE_TASK"`

	// ExpectedHeaders are the headers are used to distinguish tasks.
	ExpectedHeaders []string `env:"EXPECTED_HEADERS"`

	// HttpProxy is not used directly, however the CAPI client assumes its
	// going through a proxy for auth.
	HttpProxy string `env:"HTTP_PROXY, required"`

	// ScriptAppName is the app name that has the droplet where the script
	// lives.
	ScriptAppName string `env:"SCRIPT_APP_NAME"`

	VcapApplication VcapApplication `env:"VCAP_APPLICATION, required"`
}

type VcapApplication struct {
	CAPIAddr        string   `json:"cf_api"`
	ApplicationID   string   `json:"application_id"`
	ApplicationName string   `json:"application_name"`
	ApplicationURIs []string `json:"application_uris"`
	SpaceID         string   `json:"space_id"`
}

func (a *VcapApplication) UnmarshalEnv(data string) error {
	return json.Unmarshal([]byte(data), a)
}

func LoadConfig() (Config, error) {
	cfg := Config{}

	if err := envstruct.Load(&cfg); err != nil {
		return Config{}, err
	}

	if cfg.ScriptAppName == "" {
		cfg.ScriptAppName = cfg.VcapApplication.ApplicationName
	}

	return cfg, nil
}
