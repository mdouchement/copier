package main

import (
	"fmt"
	"io"
	"os"

	"github.com/juju/errors"
	"github.com/mdouchement/copier"
	"github.com/mdouchement/copier/cmd/util"
	"gopkg.in/urfave/cli.v2"
)

var (
	listCommand = &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "list recursively",
		Action:  listAction,
		Flags:   listFlags,
	}

	listFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "o, output-file",
			Usage: "Specify the the output file",
		},
	}
)

func listAction(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		path = "." // current directory
	}

	var w io.Writer
	if o := c.String("o"); o != "" {
		copier.MkdirAllWithFilename(o)
		dst, err := os.Create(o)
		if err != nil {
			return errors.Annotate(err, "listAction create file")
		}

		defer dst.Close()
		w = dst
	} else {
		w = os.Stdout
	}

	filenames, err := util.ListFiles(path)
	if err != nil {
		return errors.Annotate(err, "listAction")
	}

	for _, filename := range filenames {
		if _, err := io.WriteString(w, fmt.Sprintf("%s%s", filename, util.Newline)); err != nil {
			return errors.Annotate(err, "listAction")
		}
	}

	return nil
}
