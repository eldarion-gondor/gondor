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
	var err error
	var instance *gondor.Instance
	var service *gondor.Service

	api := getAPIClient(ctx)
	instance = getInstance(ctx, api, nil)

	if len(ctx.Args()) == 1 {
		service, err = api.Services.Get(instance, ctx.Args()[0])
		if err != nil {
			fatal(err.Error())
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
