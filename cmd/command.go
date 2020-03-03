package cmd

import (
	"github.com/gkontos/properties-organizer/config"
	"github.com/gkontos/properties-organizer/model"
)

type Organizer struct {
	Config      *config.Application
	ConfigFiles []model.JavaConfigFileMetadata
}
