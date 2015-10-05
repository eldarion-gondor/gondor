package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
)

func deployCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor deploy [--instance] [<git-ref>]")
		fatal(msg)
	}
	// 0. prepare API
	var source string
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	source, ok := siteCfg.instances[instance.Label]
	if !ok {
		if len(ctx.Args()) < 2 {
			usage("too few arguments")
		}
		source = ctx.Args()[0]
	}
	fmt.Printf("-----> Preparing deployment of %s to %s...", source, instance.Label)
	// 1. create a release for the instance
	release, err := api.Releases.Create(instance)
	if err != nil {
		fatal(err.Error())
	}
	cleanup := func(err error) {
		if err := api.Releases.Delete(release); err != nil {
			fatal(err.Error())
		}
		fatal(err.Error())
	}
	// 2. create a build
	build, err := api.Builds.Create(instance, release)
	if err != nil {
		cleanup(err)
	}
	// 3. perform build from source blob
	cmd := exec.Command("git", "archive", "--format=tar", source)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cleanup(err)
	}
	if err := cmd.Start(); err != nil {
		cleanup(err)
	}
	endpoint, err := build.Perform(stdout)
	if err != nil {
		cleanup(err)
	}
	if err := cmd.Wait(); err != nil {
		cleanup(err)
	}
	re := remoteExec{
		endpoint:   endpoint,
		enableTty:  false,
		httpClient: getHttpClient(ctx),
		tlsConfig:  getTLSConfig(ctx),
	}
	exitCode, err := re.execute()
	if err != nil {
		fatal(err.Error())
	}
	if exitCode > 0 {
		os.Exit(exitCode)
	}
	// 4. create a deployment for the instance pointed at the release
	fmt.Println("\n-----> Deploying...")
	deployment, err := api.Deployments.Create(instance, release)
	if err != nil {
		cleanup(err)
	}
	if err := deployment.Wait(); err != nil {
		fatal(err.Error())
	}
}
