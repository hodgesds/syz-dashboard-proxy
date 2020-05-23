// Copyright © 2020 Daniel Hodges <hodges.daniel.scott@gmail.com>
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
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	proxy "github.com/hodgesds/syz-dashboard-proxy"
	"github.com/spf13/cobra"
)

var (
	port    int
	forward []string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "syz-dashboard-proxy",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		proxy := proxy.New(forward)
		r := gin.Default()
		r.POST("/api", proxy.Proxy)
		r.GET("/metrics", proxy.Metrics)
		r.POST("/null", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "ok",
			})
		})
		r.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "ok",
			})
		})
		r.Run(fmt.Sprintf(":%d", port))
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().IntVarP(
		&port, "port", "p",
		8724,
		"HTTP Port",
	)
	RootCmd.PersistentFlags().StringSliceVarP(
		&forward, "forward", "f",
		[]string{},
		"Proxy forward",
	)
}
