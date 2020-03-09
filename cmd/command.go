package cmd

import (
	"github.com/gkontos/java-properties-pruner/config"
	"github.com/gkontos/java-properties-pruner/model"
	"github.com/manifoldco/promptui"
)

// Pruner is the application context; this seems to be used somewhat eradically
type Pruner struct {
	Config *config.Application
	// config files are either classpath or default
	ConfigFiles map[int8][]model.JavaConfigFileMetadata
}

const (
	classpathFileKey = 0
	externalFileKey  = 1
)

func promptString(name string) (string, error) {
	prompt := promptui.Prompt{
		Label:    name,
		Validate: validateEmptyInput,
	}
	return prompt.Run()
}
