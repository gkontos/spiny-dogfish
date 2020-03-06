package cmd

import (
	"github.com/gkontos/java-properties-pruner/config"
	"github.com/gkontos/java-properties-pruner/model"
	"github.com/manifoldco/promptui"
)

// Organizer is the application context; this seems to be used somewhat eradically
type Organizer struct {
	Config      *config.Application
	ConfigFiles []model.JavaConfigFileMetadata
}

func promptString(name string) (string, error) {
	prompt := promptui.Prompt{
		Label:    name,
		Validate: validateEmptyInput,
	}
	return prompt.Run()
}
