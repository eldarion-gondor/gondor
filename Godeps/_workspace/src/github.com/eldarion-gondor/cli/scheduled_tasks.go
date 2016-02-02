package gondorcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/olekukonko/tablewriter"
)

func scheduledTasksListCmd(c *CLI, ctx *cli.Context) {
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	scheduledTasks, err := api.ScheduledTasks.List(&*instance.URL)
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Schedule", "Timezone", "Command"})
	for i := range scheduledTasks {
		scheduledTask := scheduledTasks[i]
		table.Append([]string{
			*scheduledTask.Name,
			*scheduledTask.Schedule,
			*scheduledTask.Timezone,
			*scheduledTask.Command,
		})
	}
	table.Render()
}

func scheduledTasksCreateCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor scheduled-tasks create [--instance,--timezone] --name --schedule -- <executable> <arg-or-option>...")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	if ctx.String("name") == "" {
		usage("--name not defined")
	}
	if ctx.String("schedule") == "" {
		usage("--schedule not defined")
	}
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	name := ctx.String("name")
	schedule := ctx.String("schedule")
	timezone := ctx.String("timezone")
	command := strings.Join(ctx.Args(), " ")
	scheduledTask := gondor.ScheduledTask{
		Instance: instance.URL,
		Name:     &name,
		Schedule: &schedule,
		Timezone: &timezone,
		Command:  &command,
	}
	if err := api.ScheduledTasks.Create(&scheduledTask); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s scheduled task has been created.", *scheduledTask.Name))
}

func scheduledTasksDeleteCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor scheduled-tasks delete <name>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	name := ctx.Args()[0]
	if err := api.ScheduledTasks.DeleteByName(*instance.URL, name); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s scheduled task has been deleted.", name))
}
