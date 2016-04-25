package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/koofr/haproxystatsd"
)

func main() {

	hostname, _ := os.Hostname()

	app := cli.NewApp()
	app.Name = "haproxystatsd"
	app.Usage = "parse haproxy logs and send metrics to statsd"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "statsdAddr",
			Value: "127.0.0.1:8125",
			Usage: "address of statsd server",
		},
		cli.StringFlag{
			Name:  "bind",
			Value: "127.0.0.1:10514",
			Usage: "syslog server bind address",
		},
		cli.StringFlag{
			Name:  "logPattern",
			Value: "",
			Usage: "optional: specify own regex for log parsing. Read docs first!",
		},
		cli.StringFlag{
			Name:  "bucketPrefixTpl",
			Value: "",
			Usage: "optional: specify own template for bucket prefixes. Read docs first!",
		},
		cli.StringFlag{
			Name:  "nodeTag",
			Value: hostname,
			Usage: "optional: specify own tag for node. Defaults to hostname",
		},
		cli.BoolFlag{
			Name:  "dryRun",
			Usage: "just parse messages and send to stdout instead of statsd",
		},
	}

	app.Run(os.Args)
}

func run(c *cli.Context) {

	statsdAddr := c.GlobalString("statsdAddr")
	bind := c.GlobalString("bind")
	logPattern := c.GlobalString("logPattern")
	bucketPrefixTpl := c.GlobalString("bucketPrefixTpl")
	nodeTag := c.GlobalString("nodeTag")
	dryRun := c.GlobalBool("dryRun")

	cfg := haproxystatsd.Config{
		StatsdAddr:     statsdAddr,
		SyslogBindAddr: bind,
		NodeTag:        nodeTag,
		LogPattern:     logPattern,
		BucketTemplate: bucketPrefixTpl,
		DryRun:         dryRun,
	}

	fmt.Printf("Config %+v\n", cfg)

	hs, err := haproxystatsd.New(&cfg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	if err = hs.Boot(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(2)
	}
	hs.Wait()
}
