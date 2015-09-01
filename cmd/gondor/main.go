package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/pivotal-golang/bytefmt"
)

var version string
var app *cli.App

type versionInfo struct {
	Version     string
	DownloadURL string
}

func stdCmd(cmdFunc func(*cli.Context)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		checkVersion()
		if !isAuthenticated() {
			fatal("you are not authenticated. Run `gondor login` to authenticate.")
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
	app.Name = "gondor"
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
			EnvVar: "GONDOR_API_URL",
		},
		cli.StringFlag{
			Name:   "resource-group",
			Value:  "",
			Usage:  "Scope requests to given resource group (default is resource group of site otherwise personal)",
			EnvVar: "GONDOR_RESOURCE_GROUP",
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
						api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
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
							Name:  "service",
							Value: "",
							Usage: "service path",
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
						api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
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
			Name:  "create",
			Usage: "[site] create a new instance or service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "instance-kind",
					Value: "dev",
				},
				cli.StringFlag{
					Name:  "service-name",
					Value: "",
					Usage: "name of the service",
				},
				cli.StringFlag{
					Name:  "service-version",
					Value: "",
					Usage: "version for the new service",
				},
			},
			Action: stdCmd(createCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
				}
			},
		},
		{
			Name:   "delete",
			Usage:  "[site] delete instance/service for the given site/instance",
			Action: stdCmd(deleteCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
					for j := range site.Instances[i].Services {
						fmt.Printf("%s/%s\n", site.Instances[i].Label, site.Instances[i].Services[j].Name)
					}
				}
			},
		},
		{
			Name:   "list",
			Usage:  "[site] display instances/services for the given site/instance",
			Action: stdCmd(listCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
				}
			},
		},
		{
			Name:  "scale",
			Usage: "[site] scale up/down a service on an instance",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "replicas",
					Usage: "desired number of replicas",
				},
			},
			Action: stdCmd(scaleCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					for j := range site.Instances[i].Services {
						fmt.Printf("%s/%s\n", site.Instances[i].Label, site.Instances[i].Services[j].Name)
					}
				}
			},
		},
		{
			Name:   "restart",
			Usage:  "[site] restart a service on a given instance",
			Action: stdCmd(restartCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					for j := range site.Instances[i].Services {
						fmt.Printf("%s/%s\n", site.Instances[i].Label, site.Instances[i].Services[j].Name)
					}
				}
			},
		},
		{
			Name:            "run",
			Usage:           "[site] run a one-off process",
			SkipFlagParsing: true,
			Action:          stdCmd(runCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
				}
			},
		},
		{
			Name:   "deploy",
			Usage:  "[site] create a new release and deploy",
			Action: stdCmd(deployCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
				}
			},
		},
		{
			Name:  "hosts",
			Usage: "[site] manage hosts for an instance",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "List hosts for an instance",
					Action: stdCmd(hostsListCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
						site := getSite(ctx, api)
						for i := range site.Instances {
							fmt.Println(site.Instances[i].Label)
						}
					},
				},
				{
					Name:   "create",
					Usage:  "Create a host for an instance",
					Action: stdCmd(hostsCreateCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
						site := getSite(ctx, api)
						for i := range site.Instances {
							fmt.Println(site.Instances[i].Label)
						}
					},
				},
				{
					Name:   "delete",
					Usage:  "Delete a host from an instance",
					Action: stdCmd(hostsDeleteCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
						site := getSite(ctx, api)
						for i := range site.Instances {
							fmt.Println(site.Instances[i].Label)
							// todo
						}
					},
				},
			},
		},
		{
			Name:  "pg",
			Usage: "[site] manage database",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
			Subcommands: []cli.Command{
				{
					Name:            "run",
					Usage:           "Run a one-off process against the database",
					SkipFlagParsing: true,
					Action:          stdCmd(pgRunCmd),
					BashComplete: func(ctx *cli.Context) {
						if len(ctx.Args()) > 0 {
							return
						}
						api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
						site := getSite(ctx, api)
						for i := range site.Instances {
							fmt.Println(site.Instances[i].Label)
						}
					},
				},
			},
		},
		{
			Name:   "open",
			Usage:  "[site] open instance URL in browser",
			Action: stdCmd(openCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
				}
			},
		},
		{
			Name:   "env",
			Usage:  "[site] manage environment variables",
			Action: stdCmd(envCmd),
			BashComplete: func(ctx *cli.Context) {
				if len(ctx.Args()) > 0 {
					return
				}
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
					for j := range site.Instances[i].Services {
						fmt.Printf("%s/%s\n", site.Instances[i].Label, site.Instances[i].Services[j].Name)
					}
				}
			},
		},
		{
			Name:  "logs",
			Usage: "[site] view logs for an instance or service",
			Flags: []cli.Flag{
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
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
				site := getSite(ctx, api)
				for i := range site.Instances {
					fmt.Println(site.Instances[i].Label)
					for j := range site.Instances[i].Services {
						fmt.Printf("%s/%s\n", site.Instances[i].Label, site.Instances[i].Services[j].Name)
					}
				}
			},
		},
		{
			Name:  "metrics",
			Usage: "[site] view metrics for a given service",
			Action: stdCmd(func(ctx *cli.Context) {
				MustLoadSiteConfig()
				api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
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
		{
			Name:  "users",
			Usage: "[site] manage users for a site",
			Action: func(ctx *cli.Context) {
				checkVersion()
				cli.ShowSubcommandHelp(ctx)
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "List users for the site",
					Action: stdCmd(usersListCmd),
				},
				{
					Name:   "add",
					Usage:  "Add a user to site with a given role",
					Action: stdCmd(usersAddCmd),
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
	if strings.Contains(app.Version, "-dev") {
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
				"You are using an older version (%s; latest: %s) of this client.\nTo upgrade run `gondor upgrade`.\n",
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
		fatal("site not defined in gondor.yml")
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
		}
	}
	return resourceGroup
}

func getSite(ctx *cli.Context, api *gondor.Client) *gondor.Site {
	resourceGroup := getResourceGroup(ctx, api)
	_, siteName := parseSiteIdentifier(siteCfg.Identifier)
	site, err := api.Sites.Get(siteName, resourceGroup)
	if err != nil {
		fatal(err.Error())
	}
	return site
}
