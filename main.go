package main

import (
	"github.com/joho/godotenv"
	"github.com/netbrain/darknetw/cfg"
	"github.com/netbrain/darknetw/cmd/generate"
	"github.com/netbrain/darknetw/cmd/serve"
	"github.com/netbrain/darknetw/cmd/train"
	"github.com/netbrain/darknetw/cmd/validate"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	err := godotenv.Load()
	app := &cli.App{
		Name:  "darknetw",
		Usage: "a wrapper around darknet that exposes a REST API for inference, labelling and training.",
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Usage:  "serve the darknetw api service",
				Action: serveAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "host",
						Usage:   "http host",
						EnvVars: []string{"DARKNETW_HOST"},
						Value:   "localhost",
					},
					&cli.StringFlag{
						Name:    "port",
						Usage:   "http port",
						EnvVars: []string{"DARKNETW_PORT"},
						Value:   "8080",
					},
					&cli.StringFlag{
						Name:     "config",
						Usage:    "darknet config file",
						EnvVars:  []string{"DARKNETW_NN_CONFIG"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "weights",
						Usage:    "darknet weights file",
						EnvVars:  []string{"DARKNETW_NN_WEIGHTS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "data",
						Usage:    "darknet data file",
						EnvVars:  []string{"DARKNETW_NN_DATA"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "storage",
						Usage:    "input/output directory",
						EnvVars:  []string{"DARKNETW_STORAGE"},
						Required: false,
					},
				},
			},
			{
				Name:   "train",
				Usage:  "train the neural network (equivalent of darknet detector train)",
				Action: trainAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Usage:    "darknet config file",
						EnvVars:  []string{"DARKNETW_NN_CONFIG"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "weights",
						Usage:    "darknet weights file",
						EnvVars:  []string{"DARKNETW_NN_WEIGHTS"},
						Required: false,
					},
					&cli.StringFlag{
						Name:     "data",
						Usage:    "darknet data file",
						EnvVars:  []string{"DARKNETW_NN_DATA"},
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "clear",
						Usage:   "clear the number of iterations and force training",
						EnvVars: []string{"DARKNETW_NN_CLEAR"},
						Value:   false,
					},
					&cli.StringFlag{
						Name:     "storage",
						Usage:    "input/output directory",
						EnvVars:  []string{"DARKNETW_STORAGE"},
						Required: false,
					},
				},
			},
			{
				Name:   "validate",
				Usage:  "validate the accuracy of the neural network (equivalent of darknet detector map)",
				Action: validateAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Usage:    "darknet config file",
						EnvVars:  []string{"DARKNETW_NN_CONFIG"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "weights",
						Usage:    "darknet weights file",
						EnvVars:  []string{"DARKNETW_NN_WEIGHTS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "data",
						Usage:    "darknet data file",
						EnvVars:  []string{"DARKNETW_NN_DATA"},
						Required: true,
					},
				},
			},
			{
				Name:   "generate",
				Usage:  "will create a simple computer generated test dataset with circles and rectangles in a random fashion",
				Action: generateAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "output",
						Usage:    "output directory",
						EnvVars:  []string{"DARKNETW_GEN_OUTPUT"},
						Required: true,
					},
					&cli.IntFlag{
						Name:     "images",
						Usage:    "number of images to create",
						EnvVars:  []string{"DARKNETW_GEN_IMAGES"},
						Required: true,
					},
					&cli.IntFlag{
						Name:     "seed",
						Usage:    "a seed number for rng",
						EnvVars:  []string{"DARKNETW_GEN_SEED"},
						Required: false,
					},
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func ctxToCfg(ctx *cli.Context) *cfg.AppConfig {
	return &cfg.AppConfig{
		ServerConfig: &cfg.ServerConfig{
			Host: ctx.String("host"),
			Port: ctx.String("port"),
		},
		NeuralNetworkConfig: &cfg.NeuralNetworkConfig{
			ConfigFile:  ctx.String("config"),
			WeightsFile: ctx.String("weights"),
			DataFile:    ctx.String("data"),
			Clear:       ctx.Bool("clear"),
		},
		Storage: ctx.String("storage"),
	}
}

func serveAction(context *cli.Context) error {
	return serve.Run(ctxToCfg(context))
}

func trainAction(ctx *cli.Context) error {
	return train.Run(ctxToCfg(ctx))
}

func validateAction(ctx *cli.Context) error {
	return validate.Run(ctxToCfg(ctx))
}

func generateAction(ctx *cli.Context) error {
	return generate.Run(ctx.String("output"), ctx.Int("images"), ctx.Int("seed"))
}
