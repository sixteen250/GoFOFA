package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

// stats subcommand
var statsCmd = &cli.Command{
	Name:                   "stats",
	Usage:                  "fofa stats",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "fields",
			Aliases:     []string{"f"},
			Value:       "title,country",
			Usage:       "visit fofa website for more info",
			Destination: &fieldString,
		},
		&cli.IntFlag{
			Name:        "size",
			Aliases:     []string{"s"},
			Value:       5,
			Usage:       "aggs size",
			Destination: &size,
		},
		&cli.StringFlag{
			Name:        "customFields",
			Aliases:     []string{"cf"},
			Value:       "",
			Usage:       "use custom fields",
			Destination: &customFields,
		},
	},
	Action: statsAction,
}

// statsAction stats action
func statsAction(ctx *cli.Context) error {
	// valid same config
	query := ctx.Args().First()
	if len(query) == 0 {
		return errors.New("fofa query cannot be empty")
	}
	var err error
	if customFields != "" {
		fieldString, err = getCustomFields(customFields)
		if err != nil {
			return fmt.Errorf("get custom fields error, %v", err)
		}
	}
	fields := strings.Split(fieldString, ",")
	if len(fields) == 0 {
		return errors.New("fofa fields cannot be empty")
	}

	// do search
	res, err := fofaCli.Stats(query, size, fields)
	if err != nil {
		return err
	}

	for _, obj := range res {
		color.New(color.FgBlue).Fprintln(os.Stdout, "=== ", obj.Name)
		for _, item := range obj.Items {
			color.New(color.FgHiGreen).Fprint(os.Stdout, item.Name)
			fmt.Print("\t")
			color.New(color.FgHiYellow).Fprintln(os.Stdout, item.Count)
		}
	}

	return nil
}
