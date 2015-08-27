package base

import (
	"github.com/codegangsta/cli"
	"log"
	"os"
)

func Run() error {
	callC := func(f func(cfg *Configuration, ctx *cli.Context) error) func(c *cli.Context) {
		return func(c *cli.Context) {
			config, err := Load(c.GlobalString("config"))
			if err == nil {
				err = f(config, c)
			}
			if err == nil {
				log.Println("DONE!!!")

			} else {
				log.Fatalln(err)
			}
		}
	}
	callA := func(f func(app Application, ctx *cli.Context) error) func(c *cli.Context) {
		return func(c *cli.Context) {
			a, e := New(c.GlobalString("config"))
			if e == nil {
				e = f(a, c)
			}
			if e == nil {
				log.Println("DONE!!!")
			} else {
				log.Fatalln(e)
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
			Action: callA(func(a Application, c *cli.Context) error {
				a.Server()
				return nil
			}),
		},
		{
			Name:    "worker",
			Aliases: []string{"w"},
			Usage:   "Run background job",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "threads, t",
					Usage: "threads poll size",
					Value: 5,
				},
				cli.IntFlag{
					Name:  "port, p",
					Usage: "stats port",
					Value: 11111,
				},
			},
			Action: callA(func(a Application, c *cli.Context) error {
				a.Worker(c.Int("port"), c.Int("threads"))
				return nil
			}),
		},
		{
			Name:    "routes",
			Aliases: []string{"ro"},
			Usage:   "Print out all defined routes in match order, with names",
			Action: callA(func(a Application, c *cli.Context) error {
				a.Routes()
				return nil
			}),
		},
		{
			Name:    "redis",
			Aliases: []string{"re"},
			Usage:   "Start a console for the redis",
			Flags:   []cli.Flag{},
			Action: callC(func(cfg *Configuration, ctx *cli.Context) error {
				cmd, args := cfg.RedisShell()
				return Shell(cmd, args...)
			}),
		},
		{
			Name:    "nginx",
			Aliases: []string{"n"},
			Usage:   "Nginx config file demo",
			Flags:   []cli.Flag{},
			Action: callA(func(a Application, c *cli.Context) error {
				a.Nginx()
				return nil
			}),
		},
		{
			Name:    "openssl",
			Aliases: []string{"ssl"},
			Usage:   "Openssl certs command demo",
			Flags:   []cli.Flag{},
			Action: callA(func(a Application, c *cli.Context) error {
				a.Openssl()
				return nil
			}),
		},
		{
			Name:    "db:console",
			Aliases: []string{"db"},
			Usage:   "Start a console for the database",
			Flags:   []cli.Flag{},
			Action: callC(func(cfg *Configuration, ctx *cli.Context) error {
				cmd, args := cfg.DbShell()
				return Shell(cmd, args...)
			}),
		},
		{
			Name:  "db:seed",
			Usage: "Load the seed data",
			Flags: []cli.Flag{},
			Action: callA(func(a Application, c *cli.Context) error {
				return a.Seed()
			}),
		},
		{
			Name:  "db:migrate",
			Usage: "Migrate the database",
			Flags: []cli.Flag{},
			Action: callA(func(a Application, c *cli.Context) error {
				a.Migrate()
				return nil
			}),
		},
		{
			Name:  "db:drop",
			Usage: "Drops the database",
			Flags: []cli.Flag{},
			Action: callC(func(cfg *Configuration, ctx *cli.Context) error {
				cmd, args := cfg.DbDrop()
				return Shell(cmd, args...)
			}),
		},
		{
			Name:  "db:create",
			Usage: "Creates the database",
			Flags: []cli.Flag{},
			Action: callC(func(cfg *Configuration, ctx *cli.Context) error {
				cmd, args := cfg.DbCreate()
				return Shell(cmd, args...)
			}),
		},
		{
			Name:   "cache:clear",
			Usage:  "Clear cache records",
			Flags:  []cli.Flag{},
			Action: callA(func(a Application, c *cli.Context) error { return a.Clear("cache://") }),
		},
		{
			Name:  "token:clear",
			Usage: "Clear tokens records",
			Flags: []cli.Flag{},
			Action: callA(func(a Application, c *cli.Context) error {
				return a.Clear("token://")
			}),
		},
		{
			Name:   "assets:clear",
			Usage:  "Clear assets resources",
			Flags:  []cli.Flag{},
			Action: callA(func(a Application, c *cli.Context) error { return a.Clear("assets://") }),
		},
	}

	return app.Run(os.Args)
}
