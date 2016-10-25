package main

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

func starthere(jobname, config_dir string) {
	var vm TunirVM
	commandfile := filepath.Join(config_dir, fmt.Sprintf("%s.txt", jobname))
	if _, err := os.Stat(commandfile); os.IsNotExist(err) {
		fmt.Println("Missing commands file for job:", jobname)
		os.Exit(100)
	}
	viper.SetConfigName(jobname)
	viper.AddConfigPath(config_dir)
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	viper.SetDefault("PORT", "22")
	viper.SetDefault("USER", "fedora")
	backend := viper.GetString("BACKEND")
	fmt.Println("Starts a new Tunir Job.\n")

	if backend == "openstack" {
		vm, _ = BootInstanceOS()
		// First 180 seconds timeout for the vm to come up
		res := Poll(180, vm)
		if !res {
			fmt.Println("Failed to ssh into the vm.")
			panic("Failed.")
		}
	} else if backend == "bare" {
		vm = TunirVM{IP: viper.GetString("IP"), KeyFile: viper.GetString("key"),
			Port: viper.GetString("PORT")}
	} else if backend == "aws" {
		vm, _ = BootInstanceAWS()
		res := Poll(180, vm)
		if !res {
			fmt.Println("Failed to ssh into the vm.")
			panic("Failed.")
		}
	}
	commands := ReadCommands(commandfile)
	result := ExecuteTests(commands, vm)
	if backend == "openstack" {
		// Time to destroy the server
		vm.Delete()
	}
	printResultSet(result)
	if !result.Status {
		os.Exit(200)
	}
	os.Exit(0)
}

func createApp() *cli.App {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "gotun"
	app.Version = "0.1.0"
	app.Usage = "The Tunir in golang."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "job",
			Value: "",
			Usage: "the job name",
		},
		cli.StringFlag{
			Name:  "config-dir",
			Value: "",
			Usage: "the directory having configuration (default current)",
		},
	}
	app.Action = func(c *cli.Context) error {
		file_path := c.GlobalString("job")
		config_dir := c.GlobalString("config-dir")
		if config_dir == "" {
			config_dir = "./"
		}
		if file_path != "" {
			starthere(file_path, config_dir)
		}
		return nil
	}

	return app
}

func main() {
	app := createApp()
	if err := app.Run(os.Args); err != nil {
		check(err)
	}
}
