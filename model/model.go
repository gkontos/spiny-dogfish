package model

type JavaConfigFileMetadata struct {
	// configurationType should be application or bootstrap
	ConfigurationType string

	// path to the configuration file
	Path string

	// the spring profile that this file belongs with
	Profile string

	// bootstrap vs application
	ApplicationContext string
}

type JavaConfig struct {
	profile     string
	application map[string]interface{}
	bootstrap   map[string]interface{}
}
