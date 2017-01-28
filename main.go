package main

import (
	"flag"
	"os"

	"github.com/Sirupsen/logrus"
)

var (
	// Must be set at build
	Version string
	Git     string

	debug    bool
	testMode bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debugging")
	flag.BoolVar(&testMode, "test-mode", false, "Renew certificate every 120 seconds")
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	logrus.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()
	logrus.Infof("Starting Let's Encrypt Certificate Manager %s %s", Version, Git)
	context := &Context{}
	context.InitContext()
	context.Run()
}
