package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	LastNamespace string `json:"last_namespace"`
	LastContext   string `json:"last_context"`
}

func getCfgPath() (string, error) {
	cfgPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	cfgPath = filepath.Join(cfgPath, "kubectx")
	err = os.MkdirAll(cfgPath, 0600)
	if err != nil {
		return "", err
	}

	cfgPath = filepath.Join(cfgPath, "config.json")

	_, err = os.Lstat(cfgPath)
	if os.IsNotExist(err) {
		var err error
		buf := &bytes.Buffer{}
		defaultCfg := Config{
			LastNamespace: "default",
			LastContext:   "",
		}
		err = json.NewEncoder(buf).Encode(&defaultCfg)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(cfgPath, buf.Bytes(), 0600)
		if err != nil {
			return "", err
		}
	}
	return cfgPath, nil
}

func WriteCfg(cfg *Config) error {
	cfgPath, err := getCfgPath()
	if err != nil {
		return err
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cfgPath, b, 0600)
	if err != nil {
		return err
	}

	return nil
}

func ReadCfg() (*Config, error) {
	cfg := &Config{}

	cfgPath, err := getCfgPath()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	// ignore
	_ = json.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}

	// if cfg.CurrentContext != "" {
	// 	cfg.CurNS = "default"
	// }

	return cfg, nil
}
