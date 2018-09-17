// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"
	"git.profzone.net/profzone/terra/spider"
	"git.profzone.net/profzone/terra/dht"
	"time"
	"math"
)

var (
	cfgFile         string
	configFileGroup string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "terra",
	Short: "A P2P demo application",
	Run: func(cmd *cobra.Command, args []string) {
		table := dht.DistributedHashTable{
			BucketExpiredAfter:   0,
			NodeExpiredAfter:     0,
			CheckBucketPeriod:    5 * time.Second,
			MaxTransactionCursor: math.MaxUint32,
			MaxNodes:             5000,
			K:                    8,
			BucketSize:           math.MaxInt32,
			RefreshNodeNum:       256,
			Network:              "udp4",
			LocalAddr:            ":6881",
			SeedNodes: []string{
				"router.bittorrent.com:6881",
				"router.utorrent.com:6881",
				"dht.transmissionbt.com:6881",
			},
			Self:    nil,
			Handler: spider.BTHandlePacket,
		}
		table.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Errorf("%v", err)
	}
}

func init() {
	cobra.OnInitialize(initCmdConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is ./.terra.yaml)")

	RootCmd.Flags().StringVarP(&configFileGroup, "file-group", "g", "", "")
}

// initCmdConfig reads in config file and ENV variables if set.
func initCmdConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".peer-to-world" (without extension).
		viper.AddConfigPath("./")
		viper.SetConfigName(".terra")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	env := viper.GetString("ENV")
	switch env {
	case "LOCAL":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "PROD":
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
