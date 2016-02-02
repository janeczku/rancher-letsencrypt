package main

import (
	"os"

	"github.com/Sirupsen/logrus"
)

var (
	// Must be set at build
	Version string
	Git     string
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	logrus.SetOutput(os.Stdout)
}

func main() {
	logrus.Infof("Starting Let's Encrypt Certificate Manager %s %s", Version, Git)
	context := &Context{}
	context.InitContext()
	context.Run()
}
