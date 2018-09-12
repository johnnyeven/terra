package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"git.profzone.net/terra/cmd"
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	cmd.Execute()
}
