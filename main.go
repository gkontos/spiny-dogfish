package main

import (
	"fmt"
	"os"

	log "github.com/gkontos/bivalve-chronicle"
	"github.com/gkontos/java-properties-pruner/cmd"
	"github.com/gkontos/java-properties-pruner/config"
	"github.com/gkontos/java-properties-pruner/model"

	"github.com/manifoldco/promptui"
)

const (
	exitAction           = "Exit"
	viewProfileAction    = "View Profile Configuration"
	optimizeConfigAction = "Optimize Configuration"
)

var organizer *cmd.Pruner

func main() {
	v := getConfig()

	setupLogging()
	log.Info("Welcome to the Properties Compactor")
	log.Debugf("getConfig result: %v", v)

	setupApplication(&v.App)

	action, err := getAction()

	organizer.ConfigFiles = make(map[int8][]model.JavaConfigFileMetadata)

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

	if _, err := os.Stat("config.toml"); err == nil {
		filename = "config.toml"
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
	organizer = &cmd.Pruner{}
	organizer.Config = appConf
}

func setupLogging() {
	logconf := &log.LogConfig{
		Output:         "stdout",
		Level:          "debug",
		DisplayMinimal: true,
		TerminalOutput: true,
	}
	log.Configure(logconf)
}
