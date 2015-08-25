package base

import (
	"github.com/codegangsta/cli"
	"log"
	"os"
)

func Run() error {
	call := func(f func(app *Application) error) func(c *cli.Context) {
		return func(c *cli.Context) {
			a, e := New(c.GlobalString("config"))
			if e == nil {
				e = f(a)
			}
			if e != nil {
				log.Fatalf("%v", e)
			}
		}
	}

	app := cli.NewApp()
	app.Name = "itpkg"
	app.Usage = "IT-PACKAGE"
	app.Version = "v20150825"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "environment, e",
			Value:  "development",
			Usage:  "can be production, development, etc...",
			EnvVar: "ITPKG_ENV",
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.yml",
			Usage: "configuration filename",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "Start web server",
			Action:  call(func(a *Application) error { return a.Server() }),
		},
		{
			Name:    "routes",
			Aliases: []string{"ro"},
			Usage:   "Print out all defined routes in match order, with names",
			Action: func(c *cli.Context) {
				//todo
			},
		},
		{
			Name:    "redis",
			Aliases: []string{"re"},
			Usage:   "Start a console for the redis",
			Flags:   []cli.Flag{},
			Action:  call(func(a *Application) error { return a.RedisShell() }),
		},
		{
			Name:    "nginx",
			Aliases: []string{"n"},
			Usage:   "Nginx config file demo",
			Flags:   []cli.Flag{},
			Action:  call(func(a *Application) error { return a.Nginx() }),
		},
		{
			Name:    "openssl",
			Aliases: []string{"ssl"},
			Usage:   "Openssl certs command demo",
			Flags:   []cli.Flag{},
			Action:  call(func(a *Application) error { return a.Openssl() }),
		},
		{
			Name:    "db:console",
			Aliases: []string{"db"},
			Usage:   "Start a console for the database",
			Flags:   []cli.Flag{},
			Action:  call(func(a *Application) error { return a.DbShell() }),
		},
		{
			Name:  "db:seed",
			Usage: "Load the seed data",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				//todo
				//				a := Load(c.String("environment"), false)
				//				a.Seed.run()
			},
		},
		{
			Name:  "db:migrate",
			Usage: "Migrate the database",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				//todo
				//				a := Load(c.String("environment"), false)
				//				a.DbMigrate()
			},
		},
		{
			Name:   "db:drop",
			Usage:  "Drops the database",
			Flags:  []cli.Flag{},
			Action: call(func(a *Application) error { return a.DbDrop() }),
		},
		{
			Name:   "db:create",
			Usage:  "Creates the database",
			Flags:  []cli.Flag{},
			Action: call(func(a *Application) error { return a.DbCreate() }),
		},
		{
			Name:  "test:email",
			Usage: "Test mailer",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from, f",
					Usage: "from-address",
				},
				cli.StringFlag{
					Name:  "to, t",
					Usage: "to-address",
				},
			},
			Action: func(c *cli.Context) {
				//todo
				//				a := Load(c.String("environment"), false)
				//				from := c.String("from")
				//				to := c.String("to")
				//				a.TestEmail(from, to)
			},
		},
		{
			Name:   "cache:clear",
			Usage:  "Clear cache from redis",
			Flags:  []cli.Flag{},
			Action: call(func(a *Application) error { return a.ClearRedis("cache://") }),
		},
		{
			Name:   "token:clear",
			Usage:  "Clear tokens from redis",
			Flags:  []cli.Flag{},
			Action: call(func(a *Application) error { return a.ClearRedis("token://") }),
		},
	}

	return app.Run(os.Args)
}
