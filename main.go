package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"github.com/johnnyeven/terra/cmd"
)

func main() {
	logrus.SetOutput(os.Stdout)

	cmd.Execute()
}
