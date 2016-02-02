package gondorcli

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/mitchellh/go-homedir"
	"github.com/pivotal-golang/bytefmt"
)

type versionInfo struct {
	Version     string
	DownloadURL string
}

type CLI struct {
	Name    string
	Version string
	Author  string
	Email   string
	Usage   string

	Config *GlobalConfig

	api *gondor.Client
}

func (c *CLI) cmd(cmdFunc func(*CLI, *cli.Context)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		configPath, err := homedir.Expand("~/.config/gondor")
		if err != nil {
			fatal(err.Error())
		}
		if err := LoadGlobalConfig(c, ctx, configPath); err != nil {
			fatal(err.Error())
		}
		// there are some cases when the command gets called within bash
		// autocomplete which can be a bad thing!
		for i := range ctx.Args() {
			if strings.Contains(ctx.Args()[i], "generate-bash-completion") {
				os.Exit(0)
			}
		}
		cmdFunc(c, ctx)
	}
}

func (c *CLI) stdCmd(cmdFunc func(*CLI, *cli.Context)) func(*CLI, *cli.Context) {
	return func(c *CLI, ctx *cli.Context) {
		c.checkVersion()
		if !c.IsAuthenticated() {
			fatal(fmt.Sprintf("you are not authenticated. Run `%s login` to authenticate.", c.Name))
		}
		cmdFunc(c, ctx)
	}
}

func (c *CLI) Run() {
	app := cli.NewApp()
	app.Name = c.Name
	app.Version = c.Version
	app.Author = c.Author
	app.Email = c.Email
	app.Usage = c.Usage
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "cloud",
			Value:  "",
			Usage:  "cloud used for this invocation",
			EnvVar: "GONDOR_CLOUD",
		},
		cli.StringFlag{
			Name:   "cluster",
			Value:  "",
			Usage:  "cluster used for this invocation",
			EnvVar: "GONDOR_CLUSTER",
		},
		cli.StringFlag{
			Name:   "resource-group",
			Value:  "",
			Usage:  "resource group used for this invocation",
			EnvVar: "GONDOR_RESOURCE_GROUP",
		},
		cli.StringFlag{
			Name:   "site",
			Value:  "",
			Usage:  "site used for this invocation",
			EnvVar: "GONDOR_SITE",
		},
		cli.BoolFlag{
			Name:  "log-http",
			Usage: "log HTTP interactions",
		},
	}
	app.Action = func(ctx *cli.Context) {
		c.checkVersion()
		cli.ShowAppHelp(ctx)
	}
	app.Commands = []cli.Command{
		{
			Name:   "login",
			Usage:  "authenticate with a Gondor cluster",
			Action: c.cmd(loginCmd),
		},
		{
			Name:   "logout",
			Usage:  "invalidate any existing credentials with the Gondor cluster",
			Action: c.cmd(logoutCmd),
		},
		{
			Name:   "upgrade",
			Usage:  "upgrade the client to latest version supported by your cloud",
			Action: c.cmd(upgradeCmd),
		},
		{
			Name:  "resource-groups",
			Usage: "manage resource groups",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				cli.ShowSubcommandHelp(ctx)
			}),
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "show resource groups to which you belong",
					Action: c.cmd(c.stdCmd(resourceGroupListCmd)),
				},
			},
		},
		{
			Name:  "keypairs",
			Usage: "manage keypairs",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				cli.ShowSubcommandHelp(ctx)
			}),
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "List keypairs",
					Action: c.cmd(c.stdCmd(keypairsListCmd)),
				},
				{
					Name:  "create",
					Usage: "create a keypair",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "name of keypair",
						},
					},
					Action: c.cmd(c.stdCmd(keypairsCreateCmd)),
				},
				{
					Name:  "attach",
					Usage: "attach keypair to service",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
						cli.StringFlag{
							Name:  "keypair",
							Value: "",
							Usage: "name of keypair",
						},
						cli.StringFlag{
							Name:  "service",
							Value: "",
							Usage: "service path",
						},
					},
					Action: c.cmd(c.stdCmd(keypairsAttachCmd)),
				},
				{
					Name:  "detach",
					Usage: "detach keypair from service",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
						cli.StringFlag{
							Name:  "service",
							Value: "",
							Usage: "service name",
						},
					},
					Action: c.cmd(c.stdCmd(keypairsDetachCmd)),
					BashComplete: func(ctx *cli.Context) {
					},
				},
				{
					Name:   "delete",
					Usage:  "delete a keypair by name",
					Action: c.cmd(c.stdCmd(keypairsDeleteCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						resourceGroup := c.GetResourceGroup(ctx)
						keypairs, err := api.KeyPairs.List(&*resourceGroup.URL)
						if err != nil {
							return
						}
						for i := range keypairs {
							fmt.Println(*keypairs[i].Name)
						}
					},
				},
			},
		},
		{
			Name:  "sites",
			Usage: "manage sites",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				cli.ShowSubcommandHelp(ctx)
			}),
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "show sites in the resource group",
					Action: c.cmd(c.stdCmd(sitesListCmd)),
				},
				{
					Name:  "init",
					Usage: "create a site, production instance and write config",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "optional name for site",
						},
					},
					Action: c.cmd(c.stdCmd(sitesInitCmd)),
				},
				{
					Name:   "create",
					Usage:  "create a site in the resource group",
					Action: c.cmd(c.stdCmd(sitesCreateCmd)),
				},
				{
					Name:   "delete",
					Usage:  "delete a site in the resource group",
					Action: c.cmd(c.stdCmd(sitesDeleteCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						resourceGroup := c.GetResourceGroup(ctx)
						sites, err := api.Sites.List(&*resourceGroup.URL)
						if err != nil {
							return
						}
						for i := range sites {
							fmt.Println(*sites[i].Name)
						}
					},
				},
				{
					Name:   "env",
					Usage:  "",
					Action: c.cmd(c.stdCmd(sitesEnvCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						resourceGroup := c.GetResourceGroup(ctx)
						sites, err := api.Sites.List(&*resourceGroup.URL)
						if err != nil {
							return
						}
						for i := range sites {
							fmt.Println(*sites[i].Name)
						}
					},
				},
				{
					Name:  "users",
					Usage: "manage users",
					Action: c.cmd(func(c *CLI, ctx *cli.Context) {
						cli.ShowSubcommandHelp(ctx)
					}),
					Subcommands: []cli.Command{
						{
							Name:   "list",
							Usage:  "List users for the site",
							Action: c.cmd(c.stdCmd(sitesUsersListCmd)),
						},
						{
							Name:   "add",
							Usage:  "Add a user to site with a given role",
							Action: c.cmd(c.stdCmd(sitesUsersAddCmd)),
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "role",
									Value: "dev",
									Usage: "desired role for user",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:  "instances",
			Usage: "manage instances",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				cli.ShowSubcommandHelp(ctx)
			}),
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Usage: "create new instance",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "kind",
							Value: "",
							Usage: "kind of instance",
						},
					},
					Action: c.cmd(c.stdCmd(instancesCreateCmd)),
				},
				{
					Name:   "list",
					Usage:  "",
					Action: c.cmd(c.stdCmd(instancesListCmd)),
				},
				{
					Name:   "delete",
					Usage:  "",
					Action: c.cmd(c.stdCmd(instancesDeleteCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						site := c.GetSite(ctx)
						instances, err := api.Instances.List(&*site.URL)
						if err != nil {
							fatal(err.Error())
						}
						for i := range instances {
							fmt.Println(*instances[i].Label)
						}
					},
				},
				{
					Name:  "env",
					Usage: "",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(instancesEnvCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						site := c.GetSite(ctx)
						instances, err := api.Instances.List(&*site.URL)
						if err != nil {
							fatal(err.Error())
						}
						for i := range instances {
							fmt.Println(*instances[i].Label)
						}
					},
				},
			},
		},
		{
			Name:  "services",
			Usage: "manage services",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				cli.ShowSubcommandHelp(ctx)
			}),
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Usage: "create new service",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "name of the service",
						},
						cli.StringFlag{
							Name:  "version",
							Value: "",
							Usage: "version for the new service",
						},
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(servicesCreateCmd)),
				},
				{
					Name:  "list",
					Usage: "",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(servicesListCmd)),
				},
				{
					Name:  "delete",
					Usage: "",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(servicesDeleteCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						instance := c.GetInstance(ctx, nil)
						services, err := api.Services.List(&*instance.URL)
						if err != nil {
							fatal(err.Error())
						}
						for i := range services {
							fmt.Println(*services[i].Name)
						}
					},
				},
				{
					Name:   "env",
					Usage:  "",
					Action: c.cmd(c.stdCmd(servicesEnvCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						instance := c.GetInstance(ctx, nil)
						services, err := api.Services.List(&*instance.URL)
						if err != nil {
							fatal(err.Error())
						}
						for i := range services {
							fmt.Println(*services[i].Name)
						}
					},
				},
				{
					Name:  "scale",
					Usage: "scale up/down a service on an instance",
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "replicas",
							Usage: "desired number of replicas",
						},
					},
					Action: c.cmd(c.stdCmd(servicesScaleCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						instance := c.GetInstance(ctx, nil)
						services, err := api.Services.List(&*instance.URL)
						if err != nil {
							fatal(err.Error())
						}
						for i := range services {
							fmt.Println(*services[i].Name)
						}
					},
				},
				{
					Name:   "restart",
					Usage:  "restart a service on a given instance",
					Action: c.cmd(c.stdCmd(servicesRestartCmd)),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := c.GetAPIClient(ctx)
						instance := c.GetInstance(ctx, nil)
						services, err := api.Services.List(&*instance.URL)
						if err != nil {
							fatal(err.Error())
						}
						for i := range services {
							fmt.Println(*services[i].Name)
						}
					},
				},
			},
		},
		{
			Name:  "run",
			Usage: "run a one-off process",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "instance",
					Value: "",
					Usage: "instance label",
				},
			},
			Action: c.cmd(c.stdCmd(runCmd)),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := c.GetAPIClient(ctx)
				instance := c.GetInstance(ctx, nil)
				services, err := api.Services.List(&*instance.URL)
				if err != nil {
					fatal(err.Error())
				}
				for i := range services {
					fmt.Println(*services[i].Name)
				}
			},
		},
		{
			Name:  "deploy",
			Usage: "create a new release and deploy",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "instance",
					Value: "",
					Usage: "instance label",
				},
			},
			Action: c.cmd(c.stdCmd(deployCmd)),
		},
		{
			Name:  "hosts",
			Usage: "manage hosts for an instance",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				cli.ShowSubcommandHelp(ctx)
			}),
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List hosts for an instance",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(hostsListCmd)),
				},
				{
					Name:  "create",
					Usage: "Create a host for an instance",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(hostsCreateCmd)),
				},
				{
					Name:  "delete",
					Usage: "Delete a host from an instance",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(hostsDeleteCmd)),
				},
			},
		},
		{
			Name:  "scheduled-tasks",
			Usage: "manage scheduled tasks for an instance",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				cli.ShowSubcommandHelp(ctx)
			}),
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List scheduled tasks for an instance",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(scheduledTasksListCmd)),
				},
				{
					Name:  "create",
					Usage: "Create a scheduled task for an instance",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
						cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "scheduled task name",
						},
						cli.StringFlag{
							Name:  "timezone",
							Value: "UTC",
							Usage: "scheduled task timezone (default: UTC)",
						},
						cli.StringFlag{
							Name:  "schedule",
							Value: "",
							Usage: "scheduled task schedule (cron syntax)",
						},
					},
					Action: c.cmd(c.stdCmd(scheduledTasksCreateCmd)),
				},
				{
					Name:  "delete",
					Usage: "Delete a scheduled task from an instance",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: c.cmd(c.stdCmd(scheduledTasksDeleteCmd)),
				},
			},
		},
		{
			Name:   "open",
			Usage:  "open instance URL in browser",
			Action: c.cmd(c.stdCmd(openCmd)),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := c.GetAPIClient(ctx)
				instance := c.GetInstance(ctx, nil)
				services, err := api.Services.List(&*instance.URL)
				if err != nil {
					fatal(err.Error())
				}
				for i := range services {
					if *services[i].Kind == "web" {
						fmt.Println(*services[i].Name)
					}
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "instance",
					Value: "",
					Usage: "instance label",
				},
			},
		},
		{
			Name:  "logs",
			Usage: "view logs for an instance or service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "instance",
					Value: "",
					Usage: "instance label",
				},
				cli.IntFlag{
					Name:  "lines",
					Value: 20,
					Usage: "number of lines to query",
				},
			},
			Action: c.cmd(c.stdCmd(logsCmd)),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := c.GetAPIClient(ctx)
				instance := c.GetInstance(ctx, nil)
				services, err := api.Services.List(&*instance.URL)
				if err != nil {
					fatal(err.Error())
				}
				for i := range services {
					fmt.Println(*services[i].Name)
				}
			},
		},
		{
			Name:  "metrics",
			Usage: "view metrics for a given service",
			Action: c.cmd(func(c *CLI, ctx *cli.Context) {
				api := c.GetAPIClient(ctx)
				site := c.GetSite(ctx)
				if len(ctx.Args()) != 1 {
					fatal("missing service")
				}
				parts := strings.Split(ctx.Args()[0], "/")
				instanceLabel := parts[0]
				serviceName := parts[1]
				instance, err := api.Instances.Get(*site.URL, instanceLabel)
				if err != nil {
					fatal(err.Error())
				}
				service, err := api.Services.Get(*instance.URL, serviceName)
				if err != nil {
					fatal(err.Error())
				}
				series, err := api.Metrics.List(*service.URL)
				if err != nil {
					fatal(err.Error())
				}
				for i := range series {
					s := series[i]
					fmt.Printf("%s = ", s.Name)
					for j := range s.Points {
						value := s.Points[j][2]
						switch *s.Name {
						case "filesystem/limit_bytes_gauge", "filesystem/usage_bytes_gauge", "memory/usage_bytes_gauge", "memory/working_set_bytes_gauge":
							fmt.Printf("%s ", bytefmt.ByteSize(uint64(value)))
							break
						default:
							fmt.Printf("%d ", value)
						}
					}
					fmt.Println("")
				}
			}),
		},
	}
	app.Run(os.Args)
}

func (c *CLI) IsAuthenticated() bool {
	return c.Config.loaded && c.Config.Identity != nil
}

func (c *CLI) SetCloud(cloud *Cloud) {
	c.Config.Cloud = cloud
}

func (c *CLI) SetCluster(cluster *Cluster) {
	c.Config.Cluster = cluster
}

func (c *CLI) SetIdentity(identity *Identity) {
	c.Config.Identity = identity
}

func (c *CLI) GetAPIClient(ctx *cli.Context) *gondor.Client {
	if c.api == nil {
		httpClient := c.GetHttpClient(ctx)
		c.api = gondor.NewClient(c.Config.GetClientConfig(), httpClient)
		if ctx.GlobalBool("log-http") {
			c.api.EnableHTTPLogging(true)
		}
	}
	return c.api
}

func (c *CLI) GetTLSConfig(ctx *cli.Context) *tls.Config {
	var pool *x509.CertPool
	if ctx.GlobalString("ca-cert") != "" {
		pem, err := ioutil.ReadFile(ctx.GlobalString("ca-cert"))
		if err != nil {
			// warn user
		} else {
			pool = x509.NewCertPool()
			if ok := pool.AppendCertsFromPEM(pem); !ok {
				pool = nil
				// warn user
			}
		}
	}
	return &tls.Config{RootCAs: pool}
}

func (c *CLI) GetHttpClient(ctx *cli.Context) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: c.GetTLSConfig(ctx),
	}
	return &http.Client{Transport: tr}
}

func (c *CLI) checkVersion() {
	var shouldCheck bool
	var outs io.Writer
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		outs = os.Stdout
		shouldCheck = true
	} else if terminal.IsTerminal(int(os.Stderr.Fd())) {
		outs = os.Stderr
		shouldCheck = true
	}
	if strings.Contains(c.Version, "dev") {
		shouldCheck = false
	}
	if shouldCheck {
		newVersion, err := c.CheckForUpgrade()
		if err != nil {
			fmt.Fprintf(outs, errize(fmt.Sprintf(
				"Failed checking for upgrade: %s\n",
				err.Error(),
			)))
		}
		if newVersion != nil {
			fmt.Fprintf(outs, heyYou(fmt.Sprintf(
				"You are using an older version (%s; latest: %s) of this client.\nTo upgrade run `%s upgrade`.\n",
				c.Version,
				newVersion.Version,
				c.Name,
			)))
		}
	}
}

func (c *CLI) CheckForUpgrade() (*versionInfo, error) {
	r, err := http.Get("https://api.us2.gondor.io/v2/client/")
	if err != nil {
		return nil, err
	}
	var versionJson interface{}
	err = json.NewDecoder(r.Body).Decode(&versionJson)
	if err != nil {
		return nil, err
	}
	newVersion := versionJson.(map[string]interface{})["version"].(string)
	downloadUrl := versionJson.(map[string]interface{})["download_url"].(map[string]interface{})[runtime.GOOS].(map[string]interface{})[runtime.GOARCH].(string)
	if newVersion != c.Version {
		return &versionInfo{
			newVersion,
			downloadUrl,
		}, nil
	}
	return nil, nil
}

func parseSiteIdentifier(value string) (string, string) {
	if value == "" {
		fatal("site not defined (either --site or in gondor.yml)")
	}
	if strings.Count(value, "/") != 1 {
		fatal(fmt.Sprintf("invalid site value: %q", value))
	}
	parts := strings.Split(value, "/")
	return parts[0], parts[1]
}

func (c *CLI) GetResourceGroup(ctx *cli.Context) *gondor.ResourceGroup {
	api := c.GetAPIClient(ctx)
	if ctx.GlobalString("resource-group") != "" {
		resourceGroup, err := api.ResourceGroups.GetByName(ctx.GlobalString("resource-group"))
		if err != nil {
			fatal(err.Error())
		}
		return resourceGroup
	}
	if err := LoadSiteConfig(); err == nil {
		resourceGroupName, _ := parseSiteIdentifier(siteCfg.Identifier)
		resourceGroup, err := api.ResourceGroups.GetByName(resourceGroupName)
		if err != nil {
			fatal(err.Error())
		}
		return resourceGroup
	} else if _, ok := err.(ErrConfigNotFound); !ok {
		fatal(fmt.Sprintf("failed to load gondor.yml\n%s", err.Error()))
	}
	user, err := api.AuthenticatedUser()
	if err != nil {
		fatal(err.Error())
	}
	if user.ResourceGroup == nil {
		fatal("you do not have a personal resource group.")
	}
	return user.ResourceGroup
}

func (c *CLI) GetSite(ctx *cli.Context) *gondor.Site {
	var err error
	var siteName string
	var resourceGroup *gondor.ResourceGroup
	api := c.GetAPIClient(ctx)
	siteFlag := ctx.GlobalString("site")
	if siteFlag != "" {
		if strings.Count(siteFlag, "/") == 1 {
			parts := strings.Split(siteFlag, "/")
			resourceGroup, err = api.ResourceGroups.GetByName(parts[0])
			if err != nil {
				fatal(err.Error())
			}
			siteName = parts[1]
		} else {
			resourceGroup = c.GetResourceGroup(ctx)
			siteName = siteFlag
		}
	} else {
		LoadSiteConfig()
		resourceGroup = c.GetResourceGroup(ctx)
		_, siteName = parseSiteIdentifier(siteCfg.Identifier)
	}
	site, err := api.Sites.Get(siteName, &*resourceGroup.URL)
	if err != nil {
		fatal(err.Error())
	}
	return site
}

func (c *CLI) GetInstance(ctx *cli.Context, site *gondor.Site) *gondor.Instance {
	api := c.GetAPIClient(ctx)
	if site == nil {
		site = c.GetSite(ctx)
	}
	branch := siteCfg.vcs.Branch
	label := ctx.String("instance")
	if label == "" {
		if branch != "" {
			var ok bool
			label, ok = siteCfg.Branches[branch]
			if !ok {
				fatal(fmt.Sprintf("unable to map %q to an instance. Please provide --instance or map it to an instance in gondor.yml.", branch))
			}
		} else {
			fatal("instance not defined (missing --instance?).")
		}
	}
	instance, err := api.Instances.Get(*site.URL, label)
	if err != nil {
		fatal(err.Error())
	}
	return instance
}
