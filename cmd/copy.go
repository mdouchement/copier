package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/juju/errors"
	"github.com/mdouchement/copier"
	"github.com/mdouchement/copier/cmd/util"
	"gopkg.in/urfave/cli.v2"
)

var (
	copyCommand = &cli.Command{
		Name:    "copy",
		Aliases: []string{"c"},
		Usage:   "copy directory",
		Action:  copyAction,
		Flags:   copyFlags,
	}

	copyFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "from-list",
			Usage: "Specify the list of source files",
		},
		&cli.StringFlag{
			Name:  "speed",
			Usage: "Specify the copy speed (e.g 2MBps - default: 512KBps)",
		},
		&cli.StringFlag{
			Name:  "timeout",
			Usage: "Specify the timeout for copy one file (e.g 1h - default: 10m)",
		},
	}
)

func copyAction(c *cli.Context) error {
	if err := validateCopyOptions(c); err != nil {
		return errors.Annotate(err, "copyAction")
	}

	var filenames []string
	var destination string
	if fromList := c.String("from-list"); fromList != "" {
		destination = c.Args().First()

		lff, err := loadFromFile(fromList)
		if err != nil {
			return errors.Annotate(err, "loadFromFile")
		}
		filenames = lff
	} else {
		destination = c.Args().Get(1)

		lf, err := util.ListFiles(c.Args().First())
		if err != nil {
			return errors.Annotate(err, "listFiles")
		}
		filenames = lf
	}

	if !copier.Exists(destination) {
		return fmt.Errorf("No such directory %s", destination)
	}

	supervisor, err := copier.NewSupervisor(filenames, destination)
	if err != nil {
		return errors.Annotate(err, "copier.NewSupervisor")
	}

	if s := c.String("speed"); s != "" {
		s, err := util.ParseSpeed(s)
		if err != nil {
			return err
		}
		supervisor.Speed = float64(s)
	}

	if t := c.String("timeout"); t != "" {
		timeout, err := util.ParseTimeout(t)
		if err != nil {
			return err
		}
		supervisor.ExecTimeout = timeout
	}

	logfile := filepath.Join(destination, "copier.log")
	fmt.Println("Logging to", logfile)
	if err := util.StartLogger(supervisor.Logger(), logfile); err != nil {
		return errors.Annotate(err, "copier.NewSupervisor")
	}

	copyProgressBar(supervisor)

	if err := supervisor.Execute(); err != nil {
		return errors.Annotate(err, "start logger")
	}

	return nil
}

func validateCopyOptions(c *cli.Context) error {
	if c.IsSet("from-list") && c.Args().Len() > 1 {
		return errors.Annotate(errors.New("to many arguments. only target_directory must be specified when --from-list option is used"), "validateCopyOptions")
	}
	if !c.IsSet("from-list") && c.Args().Len() != 2 {
		return errors.Annotate(errors.New("to many arguments. needs source_directory and target_directory arguments"), "validateCopyOptions")
	}

	return nil
}

func loadFromFile(filename string) ([]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	str := strings.TrimSpace(string(data))

	return strings.Split(str, util.Newline), nil
}

func copyProgressBar(s *copier.Supervisor) {
	go func() {
		for {
			select {
			case <-s.Done():
				return
			case pi := <-s.Progress:
				if pi.ProxyReader == nil {
					color.Green("File %s: %s", pi.Name, pi.Status)
				} else {
					color.Green("File %s", pi.Name)
					bar := pb.New64(pi.Size).SetUnits(pb.U_BYTES)
					bar.ShowSpeed = true

					bar.Start()
					var size int64
					for n := range pi.ProxyReader.ReadChunk() {
						bar.Add(n)
						size += int64(n)
					}
					bar.Finish()
					if size != pi.Size {
						color.Red("===> Copy failed")
					}
				}
			}
		}
	}()
}
