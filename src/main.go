package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/mitchellh/go-homedir"
	"github.com/pivotal-golang/bytefmt"
)

var version string
var app *cli.App
var api *gondor.Client

type versionInfo struct {
	Version     string
	DownloadURL string
}

func stdCmd(cmdFunc func(*cli.Context)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		checkVersion()
		if !isAuthenticated() {
			fatal("you are not authenticated. Run `g3a login` to authenticate.")
		}
		// there are some cases when the command gets called within bash
		// autocomplete which can be a bad thing!
		for i := range ctx.Args() {
			if strings.Contains(ctx.Args()[i], "generate-bash-completion") {
				os.Exit(0)
			}
		}
		cmdFunc(ctx)
	}
}

func main() {
	app = cli.NewApp()
	app.Name = "g3a"
	app.Version = version
	app.Author = "Eldarion, Inc."
	app.Email = "development@eldarion.com"
	app.Usage = "command-line tool for interacting with the Gondor API"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "api-url",
			Value:  "https://api.us2.gondor.io",
			Usage:  "API URL endpoint",
			EnvVar: "G3A_API_URL",
		},
		cli.StringFlag{
			Name:   "ca-cert",
			Value:  "",
			Usage:  "certificate authority certificate file path",
			EnvVar: "G3A_CA_CERT",
		},
		cli.StringFlag{
			Name:   "resource-group",
			Value:  "",
			Usage:  "resource group used for this invocation",
			EnvVar: "G3A_RESOURCE_GROUP",
		},
		cli.StringFlag{
			Name:   "site",
			Value:  "",
			Usage:  "site used for this invocation",
			EnvVar: "G3A_SITE",
		},
		cli.BoolFlag{
			Name:  "log-http",
			Usage: "log HTTP interactions",
		},
	}
	app.Action = func(ctx *cli.Context) {
		checkVersion()
		cli.ShowAppHelp(ctx)
	}
	app.Commands = []cli.Command{
		{
			Name:   "login",
			Usage:  "authenticate with Gondor using OAuth 2",
			Action: loginCmd,
		},
		{
			Name:   "logout",
			Usage:  "log out of Gondor by revoking OAuth 2 access token",
			Action: logoutCmd,
		},
		{
			Name:   "upgrade",
			Usage:  "upgrade the client to latest version",
			Action: upgradeCmd,
		},
		{
			Name:  "resource-groups",
			Usage: "manage resource groups",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "show resource groups to which you belong",
					Action: stdCmd(resourceGroupListCmd),
				},
			},
		},
		{
			Name:  "keypairs",
			Usage: "manage keypairs",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "List keypairs",
					Action: stdCmd(keypairsListCmd),
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
					Action: stdCmd(keypairsCreateCmd),
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
					Action: stdCmd(keypairsAttachCmd),
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
					Action: stdCmd(keypairsDetachCmd),
					BashComplete: func(ctx *cli.Context) {
					},
				},
				{
					Name:   "delete",
					Usage:  "delete a keypair by name",
					Action: stdCmd(keypairsDeleteCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						resourceGroup := getResourceGroup(ctx, api)
						keypairs, err := api.KeyPairs.List(resourceGroup)
						if err != nil {
							return
						}
						for i := range keypairs {
							fmt.Println(keypairs[i].Name)
						}
					},
				},
			},
		},
		{
			Name:  "sites",
			Usage: "manage sites",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "show sites in the resource group",
					Action: stdCmd(sitesListCmd),
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
					Action: stdCmd(sitesInitCmd),
				},
				{
					Name:   "create",
					Usage:  "create a site in the resource group",
					Action: stdCmd(sitesCreateCmd),
				},
				{
					Name:   "delete",
					Usage:  "delete a site in the resource group",
					Action: stdCmd(sitesDeleteCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						resourceGroup := getResourceGroup(ctx, api)
						sites, err := api.Sites.List(resourceGroup)
						if err != nil {
							return
						}
						for i := range sites {
							fmt.Println(sites[i].Name)
						}
					},
				},
				{
					Name:   "env",
					Usage:  "",
					Action: stdCmd(sitesEnvCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						resourceGroup := getResourceGroup(ctx, api)
						sites, err := api.Sites.List(resourceGroup)
						if err != nil {
							return
						}
						for i := range sites {
							fmt.Println(sites[i].Name)
						}
					},
				},
				{
					Name:  "users",
					Usage: "manage users",
					Action: func(ctx *cli.Context) {
						checkVersion()
						cli.ShowSubcommandHelp(ctx)
					},
					Subcommands: []cli.Command{
						{
							Name:   "list",
							Usage:  "List users for the site",
							Action: stdCmd(sitesUsersListCmd),
						},
						{
							Name:   "add",
							Usage:  "Add a user to site with a given role",
							Action: stdCmd(sitesUsersAddCmd),
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
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
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
					Action: stdCmd(instancesCreateCmd),
				},
				{
					Name:   "list",
					Usage:  "",
					Action: stdCmd(instancesListCmd),
				},
				{
					Name:   "delete",
					Usage:  "",
					Action: stdCmd(instancesDeleteCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						site := getSite(ctx, api)
						for i := range site.Instances {
							fmt.Println(site.Instances[i].Label)
						}
					},
				},
				{
					Name:   "env",
					Usage:  "",
					Action: stdCmd(instancesEnvCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						site := getSite(ctx, api)
						for i := range site.Instances {
							fmt.Println(site.Instances[i].Label)
						}
					},
				},
			},
		},
		{
			Name:  "services",
			Usage: "manage services",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
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
					Action: stdCmd(servicesCreateCmd),
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
					Action: stdCmd(servicesListCmd),
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
					Action: stdCmd(servicesDeleteCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						instance := getInstance(ctx, api, nil)
						for i := range instance.Services {
							fmt.Println(instance.Services[i].Name)
						}
					},
				},
				{
					Name:   "env",
					Usage:  "",
					Action: stdCmd(servicesEnvCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						instance := getInstance(ctx, api, nil)
						for i := range instance.Services {
							fmt.Println(instance.Services[i].Name)
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
					Action: stdCmd(servicesScaleCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						instance := getInstance(ctx, api, nil)
						for i := range instance.Services {
							fmt.Println(instance.Services[i].Name)
						}
					},
				},
				{
					Name:   "restart",
					Usage:  "restart a service on a given instance",
					Action: stdCmd(servicesRestartCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := getAPIClient(ctx)
						instance := getInstance(ctx, api, nil)
						for i := range instance.Services {
							fmt.Println(instance.Services[i].Name)
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
			Action: stdCmd(runCmd),
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
			Action: stdCmd(deployCmd),
		},
		{
			Name:  "hosts",
			Usage: "manage hosts for an instance",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
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
					Action: stdCmd(hostsListCmd),
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
					Action: stdCmd(hostsCreateCmd),
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
					Action: stdCmd(hostsDeleteCmd),
				},
			},
		},
		{
			Name:  "scheduled-tasks",
			Usage: "manage scheduled tasks for an instance",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
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
					Action: stdCmd(scheduledTasksListCmd),
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
					Action: stdCmd(scheduledTasksCreateCmd),
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
					Action: stdCmd(scheduledTasksDeleteCmd),
				},
			},
		},
		{
			Name:  "pg",
			Usage: "manage database",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
			Subcommands: []cli.Command{
				{
					Name:  "run",
					Usage: "Run a one-off process against the database",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "instance",
							Value: "",
							Usage: "instance label",
						},
					},
					Action: stdCmd(pgRunCmd),
				},
			},
		},
		{
			Name:   "open",
			Usage:  "open instance URL in browser",
			Action: stdCmd(openCmd),
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
			Action: stdCmd(logsCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := getAPIClient(ctx)
				instance := getInstance(ctx, api, nil)
				for i := range instance.Services {
					fmt.Println(instance.Services[i].Name)
				}
			},
		},
		{
			Name:  "metrics",
			Usage: "view metrics for a given service",
			Action: stdCmd(func(ctx *cli.Context) {
				api := getAPIClient(ctx)
				site := getSite(ctx, api)
				if len(ctx.Args()) != 1 {
					fatal("missing service")
				}
				parts := strings.Split(ctx.Args()[0], "/")
				instanceLabel := parts[0]
				serviceName := parts[1]
				instance, err := api.Instances.Get(site, instanceLabel)
				if err != nil {
					fatal(err.Error())
				}
				service, err := api.Services.Get(instance, serviceName)
				if err != nil {
					fatal(err.Error())
				}
				series, err := api.Metrics.List(service)
				if err != nil {
					fatal(err.Error())
				}
				for i := range series {
					s := series[i]
					fmt.Printf("%s = ", s.Name)
					for j := range s.Points {
						value := s.Points[j][2]
						switch s.Name {
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
	configPath, err := homedir.Expand("~/.config/gondor/config")
	if err != nil {
		fatal(err.Error())
	}
	if err := LoadGlobalConfig(configPath); err != nil {
		fatal(err.Error())
	}
	app.Run(os.Args)
}

func getAPIClient(ctx *cli.Context) *gondor.Client {
	if api == nil {
		if !gcfg.loaded || gcfg.Client.ID == "" {
			gcfg.SetClientConfig(&gondor.Config{
				ID:          "KtcICiPMAII8FAeArUoDB97zmjqltllyUDev8HOS",
				BaseURL:     ctx.GlobalString("api-url"),
				IdentityURL: "https://identity.gondor.io",
			})
		}
		httpClient := getHttpClient(ctx)
		api = gondor.NewClient(gcfg.GetClientConfig(), httpClient)
		if ctx.GlobalBool("log-http") {
			api.EnableHTTPLogging(true)
		}
	}
	return api
}

func getTLSConfig(ctx *cli.Context) *tls.Config {
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

func getHttpClient(ctx *cli.Context) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: getTLSConfig(ctx),
	}
	return &http.Client{Transport: tr}
}

func checkVersion() {
	var shouldCheck bool
	var outs io.Writer
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		outs = os.Stdout
		shouldCheck = true
	} else if terminal.IsTerminal(int(os.Stderr.Fd())) {
		outs = os.Stderr
		shouldCheck = true
	}
	if strings.Contains(app.Version, "dev") {
		shouldCheck = false
	}
	if shouldCheck {
		newVersion, err := checkForUpgrade(app.Version)
		if err != nil {
			fmt.Fprintf(outs, errize(fmt.Sprintf(
				"Failed checking for upgrade: %s\n",
				err.Error(),
			)))
		}
		if newVersion != nil {
			fmt.Fprintf(outs, heyYou(fmt.Sprintf(
				"You are using an older version (%s; latest: %s) of this client.\nTo upgrade run `g3a upgrade`.\n",
				app.Version,
				newVersion.Version,
			)))
		}
	}
}

func checkForUpgrade(currentVersion string) (*versionInfo, error) {
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
	if newVersion != currentVersion {
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

func getResourceGroup(ctx *cli.Context, api *gondor.Client) *gondor.ResourceGroup {
	var resourceGroup *gondor.ResourceGroup
	var err error
	if ctx.GlobalString("resource-group") != "" {
		resourceGroup, err = api.ResourceGroups.GetByName(ctx.GlobalString("resource-group"))
		if err != nil {
			fatal(err.Error())
		}
	} else {
		if err := LoadSiteConfig(); err == nil {
			resourceGroupName, _ := parseSiteIdentifier(siteCfg.Identifier)
			resourceGroup, err = api.ResourceGroups.GetByName(resourceGroupName)
			if err != nil {
				fatal(err.Error())
			}
		} else {
			user, err := api.AuthenticatedUser()
			if err != nil {
				fatal(err.Error())
			}
			if user.ResourceGroup == nil {
				fatal("you do not have a personal resource group.")
			}
			resourceGroup = user.ResourceGroup
		}
	}
	return resourceGroup
}

func getSite(ctx *cli.Context, api *gondor.Client) *gondor.Site {
	var err error
	var siteName string
	var resourceGroup *gondor.ResourceGroup
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
			resourceGroup = getResourceGroup(ctx, api)
			siteName = siteFlag
		}
	} else {
		LoadSiteConfig()
		resourceGroup = getResourceGroup(ctx, api)
		_, siteName = parseSiteIdentifier(siteCfg.Identifier)
	}
	site, err := api.Sites.Get(siteName, resourceGroup)
	if err != nil {
		fatal(err.Error())
	}
	return site
}

func getInstance(ctx *cli.Context, api *gondor.Client, site *gondor.Site) *gondor.Instance {
	if site == nil {
		site = getSite(ctx, api)
	}
	var branch string
	output, err := exec.Command("git", "symbolic-ref", "HEAD").Output()
	if err == nil {
		bits := strings.Split(strings.TrimSpace(string(output)), "/")
		if len(bits) == 3 {
			branch = bits[2]
		}
	}
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
	instance, err := api.Instances.Get(site, label)
	if err != nil {
		fatal(err.Error())
	}
	return instance
}
