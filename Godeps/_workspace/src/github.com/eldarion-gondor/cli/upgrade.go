package gondorcli

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/inconshreveable/go-update"
)

func upgradeCmd(c *CLI, ctx *cli.Context) {
	newVersion, err := c.CheckForUpgrade()
	if err != nil {
		fmt.Printf(errize(fmt.Sprintf(
			"Failed checking for upgrade: %s\n",
			err.Error(),
		)))
	}
	if newVersion != nil && !strings.Contains(c.Version, "-dev") {
		err, _ := update.New().FromUrl(newVersion.DownloadURL)
		if err != nil {
			fatal(err.Error())
		}
		success(fmt.Sprintf("client has been upgraded to %s", newVersion.Version))
	} else {
		fmt.Println("You are using the latest version.")
	}
}
