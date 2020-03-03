/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gkontos/properties-organizer/cmd"
	"github.com/gkontos/properties-organizer/config"

	"github.com/manifoldco/promptui"
)

const EXIT = "Exit"
const VIEW_PROFILE = "View Profile Configuration"
const OPTIMIZE_CONFIG = "Optimize Configuration"

var organizer *cmd.Organizer

func main() {
	v := getConfig()
	fmt.Printf("getConfig result: %v", v)

	setupApplication(&v.App)

	action, err := getAction()

	for {
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
		if action == EXIT {
			break
		}
		organizer.RunInitialLoad()
		action, err = getAction()
	}
}

func getAction() (string, error) {
	prompt := promptui.Select{
		Label: "Select Action",
		Items: []string{EXIT, VIEW_PROFILE, OPTIMIZE_CONFIG},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}

	fmt.Printf("You choose %q\n", result)
	return result, nil
}

func runLoadFromDir() {
	val, _ := PromptString("Config Dir")
	fmt.Printf("You entered %s", val)
}

func validateEmptyInput(input string) error {
	if input == "" {
		return errors.New("Invalid input")
	}
	return nil
}

func PromptString(name string) (string, error) {
	prompt := promptui.Prompt{
		Label:    name,
		Validate: validateEmptyInput,
	}
	return prompt.Run()
}

func getConfig() *config.AppConfig {

	var filename string

	if _, err := os.Stat("./config/config.toml"); err == nil {
		filename = "./config/config.toml"
	} else {
		fmt.Println("No configuration available.  Exiting.")
		os.Exit(1)
	}

	if conf, err := config.LoadAppConfig(filename); err != nil {
		log.Fatal("Unable to get configuration - ", err)
		os.Exit(1)
	} else {
		return conf
	}
	fmt.Print("returning nil")
	return nil
}

func setupApplication(appConf *config.Application) {
	fmt.Printf("in application %+v", appConf)
	fmt.Println("")
	organizer = &cmd.Organizer{}
	organizer.Config = appConf
}
