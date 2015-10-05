package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eldarion-gondor/gondor-go"

	"gopkg.in/yaml.v2"
)

const configFilename = "gondor.yml"

type GlobalConfig struct {
	Version string `yaml:"version,omitempty"`
	Client  struct {
		ID          string `yaml:"id,omitempty"`
		BaseURL     string `yaml:"base_url,omitempty"`
		IdentityURL string `yaml:"identity_url,omitempty"`
		Auth        struct {
			Username     string `yaml:"username,omitempty"`
			AccessToken  string `yaml:"access_token,omitempty"`
			RefreshToken string `yaml:"refresh_token,omitempty"`
		} `yaml:"auth,omitempty"`
	} `yaml:"client,omitempty"`
	loaded   bool
	filename string
}

var gcfg GlobalConfig

func LoadGlobalConfig(filename string) error {
	gcfg.filename = filename
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		gcfg.Version = "1"
		return nil
	}
	cfg, err := ReadGlobalConfig(filename)
	if err != nil {
		return err
	}
	if cfg.Version != "1" {
		return fmt.Errorf("global config must be v1. Delete %q and log in again.", filename)
	}
	cfg.loaded = true
	gcfg = *cfg
	return nil
}

func ReadGlobalConfig(filename string) (*GlobalConfig, error) {
	var cfg GlobalConfig
	cfg.filename = filename
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	cfg.loaded = true
	return &cfg, nil
}

type clientConfigPersister struct {
	cfg *GlobalConfig
}

func (p *clientConfigPersister) Persist(config *gondor.Config) error {
	p.cfg.SetClientConfig(config)
	if err := p.cfg.Save(); err != nil {
		return err
	}
	return nil
}

func (cfg *GlobalConfig) GetClientConfig() *gondor.Config {
	config := gondor.Config{}
	config.ID = cfg.Client.ID
	config.BaseURL = cfg.Client.BaseURL
	config.IdentityURL = cfg.Client.IdentityURL
	config.Auth.Username = cfg.Client.Auth.Username
	config.Auth.AccessToken = cfg.Client.Auth.AccessToken
	config.Auth.RefreshToken = cfg.Client.Auth.RefreshToken
	config.Persister = &clientConfigPersister{cfg: cfg}
	return &config
}

func (cfg *GlobalConfig) SetClientConfig(config *gondor.Config) {
	cfg.Client.ID = config.ID
	cfg.Client.BaseURL = config.BaseURL
	cfg.Client.IdentityURL = config.IdentityURL
	cfg.Client.Auth.Username = config.Auth.Username
	cfg.Client.Auth.AccessToken = config.Auth.AccessToken
	cfg.Client.Auth.RefreshToken = config.Auth.RefreshToken
}

func (cfg *GlobalConfig) Save() error {
	data, err := yaml.Marshal(cfg)
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
	Identifier string            `yaml:"site"`
	Branches   map[string]string `yaml:"branches,omitempty"`

	instances map[string]string

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
	// reverse the branches mapping
	siteCfg.instances = make(map[string]string)
	for branch := range siteCfg.Branches {
		siteCfg.instances[siteCfg.Branches[branch]] = branch
	}
	siteCfg.loaded = true
	return nil
}

func MustLoadSiteConfig() {
	if err := LoadSiteConfig(); err != nil {
		fatal(err.Error())
	}
}
