package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/pivotal-golang/bytefmt"
)

func deployCmd(ctx *cli.Context) {
	// 0. prepare API
	var source string
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	instance := getInstance(ctx, api, nil)
	buildLabel := fmt.Sprintf("%s-%%s", filepath.Base(filepath.Dir(siteCfg.filename)))
	if ctx.Args().First() != "" {
		source = ctx.Args().First()
		var sourceLabel string
		if len(source) > 8 {
			sourceLabel = source[:8]
		} else {
			sourceLabel = source
		}
		buildLabel = fmt.Sprintf(buildLabel, sourceLabel)
	} else {
		var ok bool
		source, ok = siteCfg.instances[*instance.Label]
		if !ok {
			fatal(fmt.Sprintf("no branch could be found for %s", *instance.Label))
		} else {
			buildLabel = fmt.Sprintf(buildLabel, fmt.Sprintf("%s-%s", siteCfg.vcs.Branch, siteCfg.vcs.Commit[:8]))
		}
	}
	if siteCfg.Deploy == nil {
		fatal("gondor.yml is missing the deploy configuration.")
	}
	fmt.Printf("-----> Preparing build of %s (%s)\n", source, buildLabel)
	cleanup := func(err error) {
		if err != nil {
			fatal(err.Error())
		}
	}
	// 1. create a build
	build := &gondor.Build{
		Site:         site.URL,
		Label:        &buildLabel,
		BuildpackURL: &siteCfg.BuildpackURL,
	}
	if err := api.Builds.Create(build); err != nil {
		cleanup(err)
	}
	// 2. perform build from source blob
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
	// 3. create a deployment for the instance pointed at the release
	fmt.Printf("\n-----> Deploying to %s... ", *instance.Label)
	errc := make(chan error)
	for _, serviceName := range siteCfg.Deploy.Services {
		service, err := api.Services.Get(*instance.URL, serviceName)
		if err != nil {
			fmt.Println("error")
			fmt.Printf("       %s\n", err)
			os.Exit(1)
		}
		go func() {
			deployment := &gondor.Deployment{
				Service: service.URL,
				Build:   build.URL,
			}
			if err := api.Deployments.Create(deployment); err != nil {
				errc <- err
				return
			}
			if err := deployment.Wait(); err != nil {
				errc <- err
				return
			}
			errc <- nil
		}()
	}
	if err := <-errc; err != nil {
		fmt.Println("error")
		fmt.Printf("       %s\n", err)
		os.Exit(1)
	}
	fmt.Println("done")
}
