package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	cm "github.com/chartmuseum/go-chartmuseum/pkg/chartmuseum"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"k8s.io/helm/pkg/chartutil"
)

type (

	// Config struct map with drone plugin parameters
	Config struct {
		Server string `json:"server,omitempty"`
		Org    string `json:"org,omitempty"`
		Repo   string `json:"repo,omitempty"`
		Client *cm.Client
	}
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = "curator"
	app.Usage = "Chart Museum CLI"
	app.Version = fmt.Sprintf("0.0.1")
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "server,s",
			Value:  "",
			Usage:  "ChartMuseum API base `URL`",
			EnvVar: "CHARTMUSEUM_SERVER",
		},
		cli.StringFlag{
			Name:   "org,o",
			Value:  "",
			Usage:  "ChartMuseum Organisation",
			EnvVar: "CHARTMUSEUM_ORG",
		},
		cli.StringFlag{
			Name:   "repo,r",
			Value:  "",
			Usage:  "ChartMuseum Repo",
			EnvVar: "CHARTMUSEUM_REPO",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:      "push",
			Action:    pushCommand,
			Usage:     "Push chart-dir to server",
			ArgsUsage: "chart-dir",
		},
		cli.Command{
			Name:      "delete",
			Action:    deleteCommand,
			Usage:     "delete chart from server",
			ArgsUsage: "chart-name version",
		},
	}
	return app
}

// initConfig initializes and validates the command config
func initConfig(c *cli.Context) (*Config, error) {
	var err error
	config := &Config{
		Server: c.GlobalString("server"),
		Org:    c.GlobalString("org"),
		Repo:   c.GlobalString("repo"),
	}

	if config.Org != "" && config.Repo == "" {
		return nil, errors.Errorf("Repo required if Org is set")
	}

	// init ChartMuseum client
	if config.Client, err = cm.NewClient(config.Server, nil); err != nil {
		return nil, errors.Wrapf(err, "Could not create ChartMuseum client (server: %q)", config.Server)
	}

	return config, nil
}

func pushCommand(c *cli.Context) error {
	config, err := initConfig(c)
	if err != nil {
		fmt.Printf("%s\n", err)
		cli.ShowSubcommandHelp(c)
		return err
	}
	chartPath := c.Args().First()
	// validate chart-path is a valid chart directory
	if valid, err := chartutil.IsChartDir(chartPath); !valid {
		return errors.Wrapf(err, "Error validating chart path: %q", chartPath)
	}

	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, 60*time.Second)

	response, err := packageAndUpload(ctx, config, chartPath)
	if err != nil {
		fmt.Printf("Error while processing %q: %s\n", chartPath, err)
	} else if response.Saved {
		fmt.Printf("Succesfully Uploaded %q to %q\n", chartPath, config.Server)
	} else {
		fmt.Printf("Unexpected ChartMuseum response (Message = %q)\n", response.Message)
	}
	return nil
}

func deleteCommand(c *cli.Context) error {
	config, err := initConfig(c)
	if err != nil {
		fmt.Printf("%s\n", err)
		cli.ShowSubcommandHelp(c)
		return err
	}

	chartName := c.Args().First()
	chartVersion := c.Args().Get(1)
	ci := &cm.ChartInfo{
		Name:    &chartName,
		Version: &chartVersion,
		Org:     &config.Org,
		Repo:    &config.Repo,
	}

	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, 60*time.Second)

	response, err := config.Client.ChartService.DeleteChart(ctx, ci)
	if err != nil {
		fmt.Printf("Error while deleting %s: %s\n", ci, err)
	} else if response.Deleted {
		fmt.Printf("Succesfully deleted %s from %q\n", ci, config.Server)
	} else {
		fmt.Printf("Unexpected ChartMuseum response (Message = %q)\n", response.Message)
	}
	return nil
}

// packageAndUpload saves a helm chart directory to a compressed package and uploads it to chartmuseum
func packageAndUpload(ctx context.Context, config *Config, chart string) (*cm.Response, error) {
	tmp, err := ioutil.TempDir("", "curator-")
	if err != nil {
		return nil, errors.Wrapf(err, "Error while preparing temp Dir")
	}

	defer os.RemoveAll(tmp) // clean up

	c, err := chartutil.LoadDir(chart)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while loading Chart directory: %q", chart)
	}

	chartPackage, err := chartutil.Save(c, tmp)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while packaging Chart: %q", chart)
	}

	f, err := os.Open(chartPackage)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while opening generated Chart package: %q", chartPackage)
	}

	ci := &cm.ChartInfo{
		Name:    &c.Metadata.Name,
		Version: &c.Metadata.Version,
		Org:     &config.Org,
		Repo:    &config.Repo,
	}
	return config.Client.ChartService.UploadChart(ctx, ci, f)
}

func main() {
	app := initApp()
	app.Run(os.Args)
}
