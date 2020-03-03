package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/gkontos/properties-organizer/model"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const JAVA_CLASSPATH_RESOURCE = "src/main/resources"
const DEFAULT_PROFILE = "default"

var fileTypes = []string{"yaml", "yml", "properties"}

// filenames corresponding to the applicationContext's loaded by spring
var fileNames = []string{"application", "bootstrap"}

func (env *Organizer) RunInitialLoad() {

	configClassPathLocation := env.Config.ProjectRoot + "/" + JAVA_CLASSPATH_RESOURCE
	fmt.Printf("\rScanning %s\r", configClassPathLocation)
	fmt.Println("")
	files := make([]model.JavaConfigFileMetadata, 0)
	configFiles := env.getFiles(configClassPathLocation, files)
	env.ConfigFiles = configFiles
	fmt.Printf("Profiles Found: %v", uniqueProfiles(configFiles))
	fmt.Println("")

	for _, profile := range uniqueProfiles(configFiles) {
		fmt.Printf("found profile: %v", profile)
		fmt.Println("")
	}

	if runProfile, err := PromptString("Spring Profile (single profile or a comma separated list)"); err != nil {
		fmt.Printf("Error: %v", err)
		fmt.Println("")
	} else {
		for _, context := range fileNames {
			profileProperties := env.consolidateProfileAndContext(runProfile, context)
			d, err := yaml.Marshal(&profileProperties)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			fmt.Printf("CONFIGURATION FOR %s", context)
			fmt.Println("")
			fmt.Printf("--- t dump:\n%s\n\n", string(d))
		}
	}
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

func (env *Organizer) getFiles(searchDir string, files []model.JavaConfigFileMetadata) []model.JavaConfigFileMetadata {
	var fileInfo []os.FileInfo
	var err error
	fmt.Printf("Scanning directory %s ", searchDir)
	fmt.Println("")
	if fileInfo, err = ioutil.ReadDir(searchDir); err != nil {
		log.Printf("Unable to read directory at %s: %v", searchDir, err)
		return files
	}
	for _, file := range fileInfo {
		if file.IsDir() {
			fmt.Printf("Directory Found %s ", file.Name())
			fmt.Println("")
			return env.getFiles(searchDir+"/"+file.Name(), files)

		}
		parts := strings.Split(file.Name(), ".")
		if len(parts) > 1 {
			if _, found := Find(fileTypes, parts[1]); found {
				fileNameParts := strings.Split(parts[0], "-")
				if _, isJavaConfig := Find(fileNames, fileNameParts[0]); isJavaConfig {
					configFile := model.JavaConfigFileMetadata{}
					configFile.ConfigurationType = parts[len(parts)-1]
					configFile.Path = searchDir + "/" + file.Name()
					if len(fileNameParts) > 1 {
						configFile.Profile = fileNameParts[len(fileNameParts)-1]
					} else {
						configFile.Profile = DEFAULT_PROFILE
					}
					configFile.ApplicationContext = fileNameParts[0]

					files = append(files, configFile)
				}
			}
		}
	}

	return files
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func uniqueProfiles(files []model.JavaConfigFileMetadata) []string {
	uniqueProfiles := make([]string, 0)
	profileMap := make(map[string]bool)
	for _, configFile := range files {
		profileMap[configFile.Profile] = true
	}
	for key := range profileMap {
		uniqueProfiles = append(uniqueProfiles, key)
	}
	sort.Strings(uniqueProfiles)
	return uniqueProfiles
}

func (env *Organizer) consolidateProfileAndContext(profile string, context string) map[string]interface{} {

	commaRegex := regexp.MustCompile(`,\s+`)
	profiles := commaRegex.Split(profile, -1)
	profiles = append([]string{DEFAULT_PROFILE}, profiles...)

	profileProperties := make(map[string]interface{})
	// for each profiles, create a union of the configuration
	for _, profile := range profiles {
		if applicationMetadata, err := env.getConfigFileMetaByProfileAndContext(profile, context); err != nil {
			fmt.Printf("Error loading default profile, %", err)
			fmt.Println("")
		} else {
			props := loadFromFile(applicationMetadata)
			if len(profileProperties) == 0 {
				profileProperties = props
			} else {
				profileProperties = mergeMaps(profileProperties, props)
			}
		}
	}
	return profileProperties

}

func loadFromFile(fileMetadata model.JavaConfigFileMetadata) map[string]interface{} {

	pathParts := strings.Split(fileMetadata.Path, "/")
	configName := strings.Split(pathParts[len(pathParts)-1], ".")[0]
	pathParts = pathParts[:len(pathParts)-1]
	path := strings.Join(pathParts, "/")
	viper.SetConfigName(configName)                     // name of config file (without extension)
	viper.SetConfigType(fileMetadata.ConfigurationType) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)                           // path to look for the config file in
	err := viper.ReadInConfig()                         // Find and read the config file
	if err != nil {                                     // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}
	return viper.AllSettings()
}

func (env *Organizer) getConfigFileMetaByProfileAndContext(profile string, context string) (model.JavaConfigFileMetadata, error) {

	for _, file := range env.ConfigFiles {
		if file.ApplicationContext == context && file.Profile == profile {
			return file, nil
		}
	}
	return model.JavaConfigFileMetadata{}, fmt.Errorf("error config not found for profile %s and context: %s ", profile, context)
}

func mergeMaps(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	m3 := make(map[string]interface{})
	for k, v := range m1 {
		m3[k] = v
	}
	for k, v := range m2 {
		// if the key exists within the map
		if _, ok := m3[k]; ok {
			if reflect.TypeOf(m3[k]).Kind() == reflect.Map {
				//  if the value is a map, we need to merge the maps found
				m3[k] = mergeMaps(m3[k].(map[string]interface{}), v.(map[string]interface{}))
			} else {
				//  if the value is not a map we will overwrite the existing value with that found in m2
				m3[k] = v
			}
		} else {
			m3[k] = v
		}
	}
	return m3
}
