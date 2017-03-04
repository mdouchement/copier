package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v2"
)

var app *cli.App

func init() {
	app = &cli.App{}
	app.Name = "Copier"
	app.Version = "dev"
	app.Usage = ""

	app.Commands = []*cli.Command{
		listCommand,
		copyCommand,
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
