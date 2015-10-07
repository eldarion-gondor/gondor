package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/bytefmt"
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
		if len(ctx.Args()) == 0 {
			usage("too few arguments")
		}
		source = ctx.Args()[0]
	}
	fmt.Printf("-----> Preparing deployment of %s to %s\n", source, instance.Label)
	fmt.Printf("       Creating release... ")
	// 1. create a release for the instance
	release, err := api.Releases.Create(instance)
	if err != nil {
		fatal(err.Error())
	}
	fmt.Printf("%s\n", release.Tag)
	cleanup := func(err error) {
		if err := api.Releases.Delete(release); err != nil {
			fatal(err.Error())
		}
		if err != nil {
			fatal(err.Error())
		}
	}
	// 2. create a build
	build, err := api.Builds.Create(instance, release)
	if err != nil {
		cleanup(err)
	}
	// 3. perform build from source blob
	fmt.Printf("       Running git archive --format=tar %s... ", source)
	f, err := ioutil.TempFile("", "g3a-")
	if err != nil {
		fmt.Println("error")
		fmt.Printf("       %s\n", err)
		cleanup(nil)
	}
	cmd := exec.Command("git", "archive", "--format=tar", source)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("error")
		fmt.Printf("       %s\n", err)
		os.Remove(f.Name())
		cleanup(nil)
	}
	w := bufio.NewWriter(f)
	if err := cmd.Start(); err != nil {
		fmt.Println("error")
		fmt.Printf("       %s\n", err)
		os.Remove(f.Name())
		cleanup(nil)
	}
	go io.Copy(w, stdout)
	if err := cmd.Wait(); err != nil {
		fmt.Println("failed")
		fmt.Printf("       %s\n", err)
		os.Remove(f.Name())
		cleanup(nil)
	}
	fmt.Println("done")
	w.Flush()
	var size uint64
	fi, err := os.Stat(f.Name())
	if err == nil {
		size = uint64(fi.Size())
	}
	msg := "       Uploading tarball%s... "
	if size > 0 {
		msg = fmt.Sprintf(msg, fmt.Sprintf(" (%s)", bytefmt.ByteSize(size)))
	} else {
		msg = fmt.Sprintf(msg, "")
	}
	fmt.Print(msg)
	f.Seek(0, 0)
	endpoint, err := build.Perform(f)
	if err != nil {
		fmt.Println("error")
		fmt.Printf("       %s\n", err)
		os.Remove(f.Name())
		cleanup(nil)
	}
	fmt.Println("done")
	os.Remove(f.Name())
	re := remoteExec{
		endpoint:   endpoint,
		enableTty:  false,
		httpClient: getHttpClient(ctx),
		tlsConfig:  getTLSConfig(ctx),
		callback: func(ok bool, err error) {
			if err != nil {
				fmt.Println("error")
				fmt.Printf("       %s\n", err)
				cleanup(nil)
			}
			if !ok {
				fmt.Println("failed")
				cleanup(nil)
			}
			fmt.Println("ok")
		},
	}
	fmt.Printf("-----> Attaching to build process... ")
	exitCode, err := re.execute()
	if err != nil {
		fatal(err.Error())
	}
	if exitCode > 0 {
		os.Exit(exitCode)
	}
	// 4. create a deployment for the instance pointed at the release
	fmt.Printf("\n-----> Deploying... ")
	deployment, err := api.Deployments.Create(instance, release)
	if err != nil {
		fmt.Println("failed")
		fmt.Printf("       %s\n", err)
		os.Exit(1)
	}
	if err := deployment.Wait(); err != nil {
		fmt.Println("error")
		fmt.Printf("       %s\n", err)
		os.Exit(1)
	}
	fmt.Println("done")
}
