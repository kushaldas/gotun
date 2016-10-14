package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	backend := viper.GetString("backend")
	fmt.Println(backend)
}
