package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vahan90/skywoker/internal/logger"
	"github.com/vahan90/skywoker/internal/scanner"
)

// bestPractices is a function that checks Kubernetes workloads for best practices
// It takes a cli.Context as an argument and returns an error
// It checks the workload type and logs the namespace and workload type
// It then calls the scanner.ScanCluster function with the namespace and workload type
// It returns nil
func bestPractices(c *cli.Context) error {
	// check value for workload-type, exit if not valid
	workloadType := strings.ToLower(c.String("workload-type"))

	if workloadType != "all" && workloadType != "deployment" && workloadType != "statefulset" && workloadType != "cronjob" {
		logger.Errorf("Invalid workload type: %s, please either use `all`, `deployment`, `statefulset`, or `cronjob` the default type is all.", workloadType)
	}

	if c.Bool("verbose") {
		logger.LogLevel = logger.LogLevelInfo
	} else {
		logger.LogLevel = logger.LogLevelError
	}

	if c.String("namespace") == "" {
		fmt.Println("Scanning all namespaces for", workloadType, "workloads")
	} else {
		fmt.Println("Scanning namespace", c.String("namespace"), "for", workloadType, "workloads")
	}

	scanner.ScanCluster(c.String("namespace"), workloadType)
	return nil
}

func main() {
	app := &cli.App{
		Name:    "skywoker",
		Version: "0.1.0",
		Usage:   "Kubernetes deployment checker",
		Commands: []*cli.Command{
			{
				Name:    "best-practices",
				Aliases: []string{"bp"},
				Usage:   "Check Kubernetes workloads for best practices",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "namespace",
						Usage:   "namespace to check",
						Aliases: []string{"n", "ns"},
					},
					&cli.StringFlag{
						Name:        "workload-type",
						Usage:       "workload type to check from the following options: all, deployment.",
						Aliases:     []string{"w"},
						DefaultText: "all",
						Value:       "all",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Usage:   "verbose output",
						Aliases: []string{"v"},
					},
				},
				Action: bestPractices,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Errorf("Error: %v", err)
	}
}
