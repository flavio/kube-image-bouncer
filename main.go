package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/flavio/kube-image-bouncer/handlers"
)

func main() {
	var cert, key string
	var port int
	var debug bool

	app := cli.NewApp()
	app.Name = "kube-image-bouncer"
	app.Usage = "webhook endpoint for kube image policy admission controller"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "cert, c",
			Usage:       "Path to the certificate to use",
			EnvVar:      "BOUNCER_CERTIFICATE",
			Destination: &cert,
		},
		cli.StringFlag{
			Name:        "key, k",
			Usage:       "Path to the key to use",
			EnvVar:      "BOUNCER_KEY",
			Destination: &key,
		},
		cli.IntFlag{
			Name:        "port, p",
			Value:       1323,
			Usage:       "Port to listen to",
			EnvVar:      "BOUNCER_PORT",
			Destination: &port,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Enable extra debugging",
			EnvVar:      "BOUNCER_DEBUG",
			Destination: &debug,
		},
	}

	app.Action = func(c *cli.Context) error {
		e := echo.New()
		e.POST("/image_policy", handlers.PostImagePolicy())
		e.POST("/", handlers.PostValidatingAdmission())

		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))

		if debug {
			e.Logger.SetLevel(log.DEBUG)
		}

		var err error
		if cert != "" && key != "" {
			err = e.StartTLS(fmt.Sprintf(":%d", port), cert, key)
		} else {
			err = e.Start(fmt.Sprintf(":%d", port))
		}

		if err != nil {
			return cli.NewExitError(err, 1)
		}

		return nil
	}

	app.Run(os.Args)
}
