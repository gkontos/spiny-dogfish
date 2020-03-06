package cmd

import (
	"github.com/gkontos/java-properties-pruner/config"
	"github.com/gkontos/java-properties-pruner/model"
)

type Organizer struct {
	Config      *config.Application
	ConfigFiles []model.JavaConfigFileMetadata
}
