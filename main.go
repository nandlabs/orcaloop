package main

import (
	"os"

	"oss.nandlabs.io/golly/cli"
	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/l3" // Add this line
	"oss.nandlabs.io/orcaloop/config"
	"oss.nandlabs.io/orcaloop/service"
)

const (
	ConfigFile = "config-file"
)

var logger = l3.Get()

func main() {

	app := cli.NewCLI()

	startCmd := &cli.Command{
		Name:        "start",
		Description: "Starts the  Service",
		// Aliases:     []string{"st"}, TODO
		Handler: func(ctx *cli.Context) (err error) {
			configFile, exists := ctx.GetFlag(ConfigFile)
			logger.Info(exists)
			var options *config.Orcaloop
			if exists && configFile != "" {

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
				logger.InfoF("No Configuration File found using default configuration")
				options = config.DefaultConfig()

			}
			logger.InfoF("Starting Orcaloop service")
			// builtin.InitActions()

			err = service.Init(options)
			if err != nil {
				logger.ErrorF("Failed to initialize the service", err)
			}
			err = service.StartAndWait()
			panic(err)
		},
		Flags: []cli.Flag{
			{
				Name:    ConfigFile,
				Aliases: []string{"cf"},
				Default: "",
				Usage:   "Configuration File",
			},
		},
	}

	app.AddCommand(startCmd)

	if err := app.Execute(); err != nil {
		logger.ErrorF("Error executing the command", err)
	}
}
