package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Usage:       "Dump and export digital manuals for your Volkswagen",
		Description: "Work around Volkswagen's lack of proper offline manuals",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if c.Bool("verbose") {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}

			return ctx, nil
		},
		Commands: []*cli.Command{
			{
				Name:   "dump",
				Action: DumpAction,
				Usage:  "Dump all raw manual pages",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "vin",
						Usage:    "Car VIN",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "language",
						Usage:    "Language code (e.g. en_GB)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output-dir",
						Aliases:  []string{"O"},
						Usage:    "Output directory",
						Required: true,
					},
				},
			},
			{
				Name:   "render",
				Action: RenderAction,
				Usage:  "Render a raw dump",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input-dir",
						Usage:    "Input directory",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output-dir",
						Aliases:  []string{"O"},
						Usage:    "Output directory",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "format",
						Usage:    "Output format (e.g. html)",
						Required: true,
					},
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		slog.Error("Fatal error", slog.Any("error", err))
		os.Exit(1)
	}
}
