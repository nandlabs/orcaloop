package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"oss.nandlabs.io/golly/cli"
	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/l3"
	"oss.nandlabs.io/orcaloop/config"
	"oss.nandlabs.io/orcaloop/service"
)

const (
	ConfigFile = "config-file"
)

var logger = l3.Get()

func main() {
	app := &cli.App{
		Name:    "orcaloop",
		Usage:   "Orcaloop",
		Version: "v0.0.1",
		Action: func(ctx *cli.Context) error {
			cli.ShowCommandHelp(ctx)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "start",
				Usage:   "Starts the  Service",
				Aliases: []string{"st"},
				Action: func(ctx *cli.Context) (err error) {
					cf := ctx.GetFlag(ConfigFile)
					var options *config.Orcaloop
					if cf != nil {
						configFile, ok := cf.(string)
						if ok {
							logger.InfoF("Using Configuration File %v", configFile)
							mime := ioutils.GetMimeFromExt(configFile)
							var c codec.Codec
							var f *os.File
							f, err = os.Open(configFile)
							if err != nil {
								logger.ErrorF("Unable to open the file", err)
								return
							}
							c, err = codec.GetDefault(mime)
							if err != nil {
								logger.ErrorF("Unable to determine the file content", err)
								return
							}
							options = &config.Orcaloop{}

							err = c.Read(f, options)
							if err != nil {
								logger.ErrorF("Unable to read the file", err)
								return
							}

						} else {
							msg := fmt.Sprintf("Invalid Configuration File %v", cf)
							logger.Error(msg)
							err = errors.New(msg)
							return
						}

					} else {
						logger.InfoF("No Configuration File found using default configuration")
						options = config.DefaultConfig()

					}
					logger.InfoF("Starting Orcaloop service")

					err = service.Init(options)
					if err != nil {
						logger.ErrorF("Failed to initialize the service", err)
					}
					err = service.StartAndWait()
					panic(err)
				},
			},
		},
		Flags: []*cli.Flag{
			{
				Name:    ConfigFile,
				Aliases: []string{"cf"},
				Default: "",
				Usage:   "Configuration File",
			},
		},
	}

	if err := app.Execute(os.Args); err != nil {
		log.Fatal(err)
	}
}
