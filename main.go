package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"git.profzone.net/terra/cmd"
)

func main() {
	logrus.SetOutput(os.Stdout)

	cmd.Execute()
}
