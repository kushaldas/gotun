package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

func checkVariables() error {
	backend := viper.GetString("BACKEND")
	if backend == "openstack" {
		viper.SetEnvPrefix("OS")
		viper.AutomaticEnv()
		u := viper.GetString("USERNAME")
		p := viper.GetString("PASSWORD")
		t := viper.GetString("TENANT_ID")
		if (u == "") || (p == "") || (t == "") {
			return errors.New("Missing secret configuration value(s).")
		}
	} else if backend == "aws" {
		viper.SetEnvPrefix("AWS")
		viper.AutomaticEnv()
		k := viper.GetString("KEY")
		s := viper.GetString("SECRET")
		if (k == "") || (s == "") {
			return errors.New("Missing secret configuration value(s).")
		}
	}
	return nil
}

func starthere(jobname, config_dir string) {
	var vm TunirVM
	var commands []string
	var result ResultSet
	currentruninfo := make(map[string]string)

	vmdict := make(map[string]TunirVM)
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
	err = checkVariables()
	if err != nil {
		fmt.Println(err)
		os.Exit(111)
	}
	viper.SetDefault("PORT", "22")
	viper.SetDefault("USER", "fedora")
	viper.SetDefault("NUMBER", 1)
	viper.SetDefault("NODELETE", false)

	backend := viper.GetString("BACKEND")
	fmt.Println("Starts a new Tunir Job.\n")

	if backend == "openstack" {
		number := viper.GetInt("NUMBER")
		for i := 1; i <= number; i++ {
			vmname := fmt.Sprintf("gotun-%d", i)
			vm, err = BootInstanceOS(vmname)
			if err != nil {
				// We do not have an instance
				fmt.Println("We do not have an instance.")
				goto ERROR_NOIP
			}
			// First 180 seconds timeout for the vm to come up
			err = Poll(180, vm)
			if err != nil {
				fmt.Println("Failed to ssh into the vm.")
				goto ERROR_NOIP
			}
			// All good, add in the dict
			name := fmt.Sprintf("vm%d", i)
			vmdict[name] = vm
		}
	} else if backend == "bare" {
		localvms := viper.GetStringMapString("VMS")
		for k := range localvms {
			vm = TunirVM{IP: localvms[k], KeyFile: viper.GetString("key"),
				Port: viper.GetString("PORT")}
			vmdict[k] = vm
		}
	} else if backend == "aws" {
		vm, err = BootInstanceAWS()
		if err != nil {
			// We do not have an instance
			fmt.Println("We do not have an instance.")
			goto ERROR_NOIP
		}

		err = Poll(300, vm)
		if err != nil {
			fmt.Println("Failed to ssh into the vm.")
			goto ERROR_NOIP
		}
		vmdict["vm1"] = vm
	}
	// Now time to dump IP/SSH information on a file.
	for k := range vmdict {
		vm = vmdict[k]
		ip, _ := vm.GetDetails()
		currentruninfo[k] = ip

	}
	currentruninfo["user"] = viper.GetString("USER")
	currentruninfo["keyfile"] = viper.GetString("key")
	writeIPinformation(currentruninfo)

	commands = ReadCommands(commandfile)
	result = ExecuteTests(commands, vmdict)
ERROR_NOIP:
	if backend == "openstack" || backend == "aws" {
		// Time to destroy the server
		// Do it over a loop
		if viper.GetBool("NODELETE") == false {
			for k := range vmdict {
				vm = vmdict[k]
				vm.Delete()
			}
		}
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
