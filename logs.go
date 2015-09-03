package main

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/mgutz/ansi"
)

var blue func(string) string = ansi.ColorFunc("blue+b")
var red func(string) string = ansi.ColorFunc("red+b")

func logsCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor logs [--lines] (<instance-label> | <service-identifier>)")
		fatal(msg)
	}

	MustLoadSiteConfig()

	var err error
	var site *gondor.Site
	var instance *gondor.Instance
	var service *gondor.Service
	var instanceLabel, serviceName string

	api := getAPIClient(ctx)
	site = getSite(ctx, api)

	if len(ctx.Args()) >= 1 {
		arg := ctx.Args()[0]
		if strings.Contains(arg, "/") {
			parts := strings.Split(arg, "/")
			instanceLabel = parts[0]
			serviceName = parts[1]
		} else {
			instanceLabel = arg
		}
	}

	if instanceLabel == "" {
		usage("too few arguments")
	} else {
		instance, err = api.Instances.Get(site, instanceLabel)
		if err != nil {
			fatal(err.Error())
		}
		if serviceName != "" {
			service, err = api.Services.Get(instance, serviceName)
			if err != nil {
				fatal(err.Error())
			}
		}
	}

	var records []*gondor.LogRecord

	if instance != nil && service == nil {
		records, err = api.Logs.ListByInstance(instance, ctx.Int("lines"))
		if err != nil {
			fatal(err.Error())
		}
	} else if service != nil {
		records, err = api.Logs.ListByService(service, ctx.Int("lines"))
		if err != nil {
			fatal(err.Error())
		}
	}

	var color func(string) string

	for i := range records {
		record := records[i]
		switch record.Stream {
		case "stdout":
			color = blue
			break
		case "stderr":
			color = red
			break
		}
		fmt.Printf(
			"%s %s\n",
			color(fmt.Sprintf(
				"[%s; %s]",
				record.Timestamp,
				record.Tag,
			)),
			strings.TrimSpace(record.Message),
		)
	}
}
