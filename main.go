package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cahaba-ts/cahaba/commands"
	"github.com/urfave/cli"
)

const version = `2022.0`

func main() {
	app := cli.NewApp()
	app.Name = "cahaba"
	app.Description = "Generate or Build Volumes (Light Novels)"
	app.Commands = []cli.Command{
		{
			Name:  "version",
			Usage: "Show version of cahaba",
			Action: func(c *cli.Context) error {
				fmt.Println("Version: ", version)
				return nil
			},
		},
		{
			Name:   "build",
			Usage:  "Build a volume in a directory or the current directory",
			Action: commands.Build,
		},
		{
			Name:   "new",
			Usage:  "Generate a new volume in a directory or the current directory",
			Action: commands.New,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Run: ", err)
	}
}
