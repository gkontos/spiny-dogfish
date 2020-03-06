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
	"fmt"
	"os"

	log "github.com/gkontos/bivalve-chronicle"
	"github.com/gkontos/java-properties-pruner/cmd"
	"github.com/gkontos/java-properties-pruner/config"

	"github.com/manifoldco/promptui"
)

const (
	exitAction           = "Exit"
	viewProfileAction    = "View Profile Configuration"
	optimizeConfigAction = "Optimize Configuration"
)

var organizer *cmd.Organizer

func main() {
	v := getConfig()

	setupLogging()
	log.Info("Welcome to the Properties Compactor")
	log.Debugf("getConfig result: %v", v)

	setupApplication(&v.App)

	action, err := getAction()

	organizer.LoadConfigFileMetadata()

	for {
		if err != nil {
			log.Errorf("Error: %v", err)
		}
		if action == exitAction {
			break
		}
		if action == viewProfileAction {
			organizer.RunInitialLoad()
		}
		if action == optimizeConfigAction {
			organizer.PruneProperties()
		}
		action, err = getAction()
	}
}

func getAction() (string, error) {
	prompt := promptui.Select{
		Label: "Select Action",
		Items: []string{exitAction, viewProfileAction, optimizeConfigAction},
	}

	_, result, err := prompt.Run()

	if err != nil {
		log.Errorf("Prompt failed %v\n", err)
		return "", err
	}

	log.Infof("You choose %q\n", result)
	return result, nil
}

func getConfig() *config.AppConfig {

	var filename string

	if _, err := os.Stat("./config/config.toml"); err == nil {
		filename = "./config/config.toml"
	} else {
		panic("No configuration available.  Exiting.")
	}

	if conf, err := config.LoadAppConfig(filename); err != nil {
		panic(fmt.Sprintf("Unable to get configuration - %s", err))
	} else {
		return conf
	}
}

func setupApplication(appConf *config.Application) {
	log.Debugf("in application %+v", appConf)
	organizer = &cmd.Organizer{}
	organizer.Config = appConf
}

func setupLogging() {
	logconf := &log.LogConfig{
		Output:         "stdout",
		Level:          "info",
		DisplayMinimal: true,
		TerminalOutput: true,
	}
	log.Configure(logconf)
}
