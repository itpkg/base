package base

import (
	"github.com/codegangsta/cli"
	"log"
	"os"
)

func Run() error {

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
			Action: func(c *cli.Context) {
				a, e := New(c.GlobalString("config"))
				if e == nil {
					e = a.Server()
				}
				if e != nil {
					log.Fatalf("Error on start server: %v", e)
				}
			},
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
			Action: func(c *cli.Context) {
				//todo
				//				a := New(c.String("environment"))
				//				a.RedisShell()
			},
		},
		{
			Name:    "nginx",
			Aliases: []string{"n"},
			Usage:   "Nginx config file demo",
			Flags:   []cli.Flag{},
			Action: func(c *cli.Context) {
				if a, e := New(c.GlobalString("config")); e == nil {
					a.Nginx()
				} else {
					log.Fatalf("%v", e)
				}
			},
		},
		{
			Name:    "openssl",
			Aliases: []string{"ssl"},
			Usage:   "Openssl certs command demo",
			Flags:   []cli.Flag{},
			Action: func(c *cli.Context) { //todo
				if a, e := New(c.GlobalString("config")); e == nil {
					a.Openssl()
				} else {
					log.Fatalf("%v", e)
				}
			},
		},
		{
			Name:    "db:console",
			Aliases: []string{"db"},
			Usage:   "Start a console for the database",
			Flags:   []cli.Flag{},
			Action: func(c *cli.Context) {
				//todo
				//				a := New(c.String("environment"))
				//				a.DbShell()
			},
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
			Name:  "db:drop",
			Usage: "Drops the database",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				//todo
				//				a := New(c.String("environment"))
				//				a.DbDrop()
			},
		},
		{
			Name:  "db:create",
			Usage: "Creates the database",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				//todo
				//				a := New(c.String("environment"))
				//				a.DbCreate()
			},
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
			Name:  "cache:clear",
			Usage: "Clear cache from redis",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				//todo
				//				a := Load(c.String("environment"), false)
				//				if e := a.clearRedis("cache://*"); e != nil {
				//					log.Fatalf("Error on clear cache: %v", e)
				//				}
			},
		},
		{
			Name:  "token:clear",
			Usage: "Clear tokens from redis",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				//todo
				//				a := Load(c.String("environment"), false)
				//				if e := a.clearRedis("token://*"); e != nil {
				//					log.Fatalf("Error on clear tokens: %v", e)
				//				}
			},
		},
	}

	return app.Run(os.Args)
}
