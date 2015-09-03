package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type GlobalConfig struct {
	Auth struct {
		Username    string `yaml:"username,omitempty"`
		AccessToken string `yaml:"access_token,omitempty"`
	} `yaml:"auth,omitempty"`
	loaded   bool
	filename string
}

var gcfg GlobalConfig

func LoadGlobalConfig(filename string) error {
	gcfg.filename = filename
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &gcfg)
	if err != nil {
		return err
	}
	gcfg.loaded = true
	return nil
}

func (cfg *GlobalConfig) Save() error {
	data, err := yaml.Marshal(&gcfg)
	if err != nil {
		return err
	}
	configDir := filepath.Dir(cfg.filename)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create %s: %s", filepath.Dir(cfg.filename), err)
		}
	}
	if err := ioutil.WriteFile(cfg.filename, data, 0600); err != nil {
		return fmt.Errorf("unable to write %s: %s", cfg.filename, err)
	}
	return nil
}

type SiteConfig struct {
	Identifier string `yaml:"site"`

	loaded   bool
	filename string
}

var siteCfg SiteConfig

func FindSiteConfig() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "gondor.yml"), nil
}

func LoadSiteConfigFromFile(filename string, dst interface{}) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("cannot find gondor.yml")
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, dst)
	if err != nil {
		return err
	}
	return nil
}

func LoadSiteConfig() error {
	filename, err := FindSiteConfig()
	if err != nil {
		return err
	}
	siteCfg.filename = filename
	if err := LoadSiteConfigFromFile(filename, &siteCfg); err != nil {
		return err
	}
	siteCfg.loaded = true
	return nil
}

func MustLoadSiteConfig() {
	if err := LoadSiteConfig(); err != nil {
		fatal(err.Error())
	}
}
