package main

import (
	"fmt"
	"github.com/blinkops/blink-http/openapi"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"gen"},
				Subcommands: []*cli.Command{
					{
						Name:    "named-actions",
						Aliases: []string{"na"},
						Usage:   "generate named actions from openapi file",
						Action:  generateNamedActions,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "path",
								Aliases:  []string{"p"},
								Required: true,
								Usage:    "relative path to the openapi/mask files and also for the generated output. for example, './plugins/github/'",
							},
							&cli.StringFlag{
								Name:        "mask",
								Aliases:     []string{"m"},
								Value:       "mask.yaml",
								Usage:       "mask file name",
								DefaultText: "mask.yaml",
							},
							&cli.StringFlag{
								Name:        "openapi",
								Aliases:     []string{"o"},
								Value:       "openapi.yaml",
								Usage:       "openapi file name",
								DefaultText: "openapi.yaml",
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateNamedActions(c *cli.Context) error {
	return openapi.GenerateNamedActions(
		c.String("path"),
		c.String("openapi"),
		c.String("mask"),
	)
}
