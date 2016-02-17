package gondorcli

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"

	"gopkg.in/yaml.v2"
)

const configFilename = "gondor.yml"

type ErrConfigNotFound struct{}

func (err ErrConfigNotFound) Error() string {
	return "gondor.yml does not exist"
}

type OAuth2Config struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Identity struct {
	Provider string       `json:"provider"`
	Username string       `json:"username"`
	OAuth2   OAuth2Config `json:"oauth2"`
}

type Cloud struct {
	Name           string                `json:"name"`
	Identity       CloudIdentityProvider `json:"identity"`
	CurrentCluster string                `json:"current-cluster"`
	Clusters       []*Cluster            `json:"clusters"`
}

func (c *Cloud) GetClusterByName(name string) (*Cluster, error) {
	var ret *Cluster
	var found bool
	if c.Clusters != nil {
		for _, cluster := range c.Clusters {
			if cluster.Name == name {
				found = true
				ret = cluster
				break
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("%q cluster not found.", name)
	}
	return ret, nil
}

func (c *Cloud) GetCurrentCluster() *Cluster {
	var ret *Cluster
	if c.Clusters != nil {
		for _, cluster := range c.Clusters {
			if cluster.Name == c.CurrentCluster {
				ret = cluster
				break
			}
		}
	}
	return ret
}

type CloudIdentityProvider struct {
	Type     string `json:"type"`
	Location string `json:"location"`
	ClientID string `json:"client-id"`
}

type Cluster struct {
	Name                     string `json:"name"`
	Location                 string `json:"location"`
	CertificateAuthority     string `json:"certificate-authority"`
	CertificateAuthorityData []byte `json:"certificate-authority-data"`
	InsecureSkipVerify       bool   `json:"insecure-skip-verify"`
}

func (cluster *Cluster) GetCertificateAuthority() (*x509.Certificate, error) {
	var caData []byte
	if cluster.CertificateAuthority != "" {
		var err error
		caData, err = ioutil.ReadFile(cluster.CertificateAuthority)
		if err != nil {
			return nil, err
		}
	} else if cluster.CertificateAuthorityData != nil {
		caData = cluster.CertificateAuthorityData
	} else {
		return nil, nil
	}
	return x509.ParseCertificate(caData)
}

type GlobalConfig struct {
	Identities   []*Identity `json:"identities"`
	CurrentCloud string      `json:"current-cloud"`
	Clouds       []*Cloud    `json:"clouds"`

	Cloud    *Cloud
	Cluster  *Cluster
	Identity *Identity

	root   string
	loaded bool
	once   sync.Once
}

func (cfg *GlobalConfig) GetCloudByName(name string) (*Cloud, error) {
	var ret *Cloud
	var found bool
	if cfg.Clouds != nil {
		for _, c := range cfg.Clouds {
			if c.Name == name {
				found = true
				ret = c
				break
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("%q cloud not found.", name)
	}
	return ret, nil
}

func (cfg *GlobalConfig) GetCurrentCloud() *Cloud {
	var ret *Cloud
	if cfg.Clouds != nil {
		for _, c := range cfg.Clouds {
			if c.Name == cfg.CurrentCloud {
				ret = c
				break
			}
		}
	}
	return ret
}

func LoadGlobalConfig(c *CLI, ctx *cli.Context, root string) error {
	var err error
	if c.Config == nil {
		c.Config = &GlobalConfig{
			root: root,
		}
	} else {
		c.Config.root = root
	}
	c.Config.once.Do(func() {
		// create config directories if they do not exist
		if _, err := os.Stat(root); os.IsNotExist(err) {
			if err := os.Mkdir(root, 0700); err != nil {
				err = fmt.Errorf("failed to create %s: %s", root, err)
				return
			}
		} else {
			// identity.json
			data, err := ioutil.ReadFile(path.Join(root, "identity.json"))
			if err == nil {
				if err := json.Unmarshal(data, &c.Config); err != nil {
					return
				}
			} else {
				if !os.IsNotExist(err) {
					return
				}
			}
			// clouds.json
			data, err = ioutil.ReadFile(path.Join(root, "clouds.json"))
			if err == nil {
				if err := json.Unmarshal(data, &c.Config); err != nil {
					return
				}
			} else {
				if !os.IsNotExist(err) {
					return
				}
			}
		}
		if c.Config.Cloud == nil {
			var cloud *Cloud
			cloudName := ctx.GlobalString("cloud")
			if cloudName != "" {
				if cloud, err = c.Config.GetCloudByName(cloudName); err != nil {
					return
				}
			} else {
				cloud = c.Config.GetCurrentCloud()
				if cloud == nil {
					err = fmt.Errorf("current cloud not specified; use --cloud or set current-cloud in %s", path.Join(root, "clouds.json"))
					return
				}
			}
			c.Config.Cloud = cloud
		}
		if c.Config.Cluster == nil {
			var cluster *Cluster
			clusterName := ctx.GlobalString("cluster")
			if clusterName != "" {
				if cluster, err = c.Config.Cloud.GetClusterByName(clusterName); err != nil {
					return
				}
			} else {
				cluster = c.Config.Cloud.GetCurrentCluster()
				if cluster == nil {
					err = fmt.Errorf("current cluster not specified; use --cluster or set current-cluster in %s of %s", c.Config.Cloud.Name, path.Join(root, "clouds.json"))
					return
				}
			}
			c.Config.Cluster = cluster
		}
		if c.Config.Identity == nil {
			var identity *Identity
			for _, i := range c.Config.Identities {
				if i.Provider == c.Config.Cloud.Identity.Location {
					identity = i
					break
				}
			}
			c.Config.Identity = identity
		}
		c.Config.loaded = true
	})
	return err
}

type clientConfigPersister struct {
	cfg *GlobalConfig
}

func (p *clientConfigPersister) Persist(config *gondor.Config) error {
	p.cfg.Identity = &Identity{
		Provider: p.cfg.Cloud.Identity.Location,
		Username: config.Auth.Username,
		OAuth2: OAuth2Config{
			AccessToken:  config.Auth.AccessToken,
			RefreshToken: config.Auth.RefreshToken,
		},
	}
	m := make(map[string]*Identity)
	for i := range p.cfg.Identities {
		m[p.cfg.Identities[i].Provider] = p.cfg.Identities[i]
	}
	m[p.cfg.Cloud.Identity.Location] = p.cfg.Identity
	var identities []*Identity
	for _, i := range m {
		identities = append(identities, i)
	}
	p.cfg.Identities = identities
	if err := p.cfg.SaveIdentities(); err != nil {
		return err
	}
	return nil
}

func (cfg *GlobalConfig) GetClientConfig() *gondor.Config {
	config := gondor.Config{}
	config.ID = cfg.Cloud.Identity.ClientID
	config.BaseURL = fmt.Sprintf("https://%s", cfg.Cluster.Location)
	config.IdentityURL = fmt.Sprintf("https://%s", cfg.Cloud.Identity.Location)
	if cfg.Identity != nil {
		config.Auth.Username = cfg.Identity.Username
		config.Auth.AccessToken = cfg.Identity.OAuth2.AccessToken
		config.Auth.RefreshToken = cfg.Identity.OAuth2.RefreshToken
	}
	config.Persister = &clientConfigPersister{cfg: cfg}
	return &config
}

func (cfg *GlobalConfig) SaveIdentities() error {
	c := struct {
		Identities []*Identity `json:"identities"`
	}{
		Identities: cfg.Identities,
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, data, "", "  ")
	out.WriteString("\n")
	filename := path.Join(cfg.root, "identity.json")
	if err := ioutil.WriteFile(filename, out.Bytes(), 0600); err != nil {
		return fmt.Errorf("unable to write %s: %s", filename, err)
	}
	return nil
}

type DeployConfig struct {
	Services []string `yaml:"services"`
}

type VCSMetadata struct {
	Branch string
	Commit string
}

type SiteConfig struct {
	Identifier   string            `yaml:"site"`
	BuildpackURL string            `yaml:"buildpack,omitempty"`
	Branches     map[string]string `yaml:"branches,omitempty"`
	Deploy       *DeployConfig     `yaml:"deploy,omitempty"`

	instances map[string]string

	once     sync.Once
	filename string
	vcs      VCSMetadata
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
		return ErrConfigNotFound{}
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
	var err error
	siteCfg.once.Do(func() {
		filename, err := FindSiteConfig()
		if err != nil {
			return
		}
		siteCfg.filename = filename
		if err := LoadSiteConfigFromFile(filename, &siteCfg); err != nil {
			return
		}
		// git metadata
		var branch string
		output, err := exec.Command("git", "symbolic-ref", "HEAD").Output()
		if err == nil {
			bits := strings.Split(strings.TrimSpace(string(output)), "/")
			if len(bits) == 3 {
				branch = bits[2]
			}
		}
		var commit string
		output, err = exec.Command("git", "rev-parse", branch).Output()
		if err == nil {
			commit = strings.TrimSpace(string(output))
		}
		siteCfg.vcs = VCSMetadata{
			Branch: branch,
			Commit: commit,
		}
		// reverse the branches mapping
		siteCfg.instances = make(map[string]string)
		for branch := range siteCfg.Branches {
			siteCfg.instances[siteCfg.Branches[branch]] = branch
		}
	})
	return err
}

func MustLoadSiteConfig() {
	if err := LoadSiteConfig(); err != nil {
		fatal(err.Error())
	}
}
