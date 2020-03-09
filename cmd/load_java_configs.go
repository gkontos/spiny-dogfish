package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"

	log "github.com/gkontos/bivalve-chronicle"

	"github.com/gkontos/java-properties-pruner/model"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	javaClasspathResourcePath = "src/main/resources"
	defaultProfileKey         = "default"
)

var fileTypes = []string{"yaml", "yml", "properties"}

// filenames corresponding to the applicationContext's loaded by spring
var fileNames = []string{"application", "bootstrap"}

// RunInitialLoad will pull in configurations from the configured locations
func (appCtx *Pruner) RunInitialLoad() {

	for _, profile := range uniqueProfiles(appCtx.ConfigFiles) {
		log.Infof("found profile: %v", profile)
	}
	appCtx.displayCombinedProfile()
}

// LoadConfigFileMetadata will load the file metadata for configs
func (appCtx *Pruner) LoadConfigFileMetadata() {
	configClassPathLocation := appCtx.Config.ProjectRoot + "/" + javaClasspathResourcePath
	log.Infof("Scanning %s", configClassPathLocation)
	files := make([]model.JavaConfigFileMetadata, 0)
	configFiles := appCtx.getFiles(configClassPathLocation, files)
	appCtx.ConfigFiles[classpathFileKey] = configFiles

	log.Infof("Scanning %s", appCtx.Config.ExternalConfiguration)
	files = make([]model.JavaConfigFileMetadata, 0)
	configFiles = appCtx.getFiles(appCtx.Config.ExternalConfiguration, files)
	appCtx.ConfigFiles[externalFileKey] = configFiles
}

func (appCtx *Pruner) displayCombinedProfile() {
	if runProfile, err := promptString("Spring Profile (single profile or a comma separated list)"); err != nil {
		log.Errorf("Error: %v", err)
	} else {
		for _, context := range fileNames {
			profileProperties := appCtx.unionProfileAndContext(runProfile, context)
			d, err := yaml.Marshal(&profileProperties)
			if err != nil {
				panic(fmt.Sprintf("error: %v", err))
			}
			log.Infof("CONFIGURATION FOR %s", context)
			log.Infof("--- t dump:\n%s\n\n", string(d))
		}
	}
}

func validateEmptyInput(input string) error {
	if input == "" {
		return errors.New("Invalid input")
	}
	return nil
}

// getFiles will recursively crawl the search directory and return a list of the configuration files at that location
func (appCtx *Pruner) getFiles(searchDir string, files []model.JavaConfigFileMetadata) []model.JavaConfigFileMetadata {
	var fileInfo []os.FileInfo
	var err error
	log.Infof("Scanning directory %s ", searchDir)
	if fileInfo, err = ioutil.ReadDir(searchDir); err != nil {
		log.Errorf("Unable to read directory at %s: %v", searchDir, err)
		return files
	}
	for _, file := range fileInfo {
		if file.IsDir() {
			log.Infof("Directory Found %s ", file.Name())
			return appCtx.getFiles(searchDir+"/"+file.Name(), files)

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
						configFile.Profile = defaultProfileKey
					}
					configFile.ApplicationContext = fileNameParts[0]

					files = append(files, configFile)
				}
			}
		}
	}

	return files
}

// Find will return the index of an item within a slice if it exists.  If the element is not in the slice, Find will return -1
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func uniqueProfiles(files map[int8][]model.JavaConfigFileMetadata) []string {
	uniqueProfiles := make([]string, 0)
	profileMap := make(map[string]bool)
	for _, v := range files {

		for _, configFile := range v {
			profileMap[configFile.Profile] = true
		}
	}
	for key := range profileMap {
		uniqueProfiles = append(uniqueProfiles, key)
	}
	sort.Strings(uniqueProfiles)
	return uniqueProfiles
}

func (appCtx *Pruner) unionProfileAndContext(profile string, context string) map[string]interface{} {

	commaRegex := regexp.MustCompile(`,\s+`)
	profiles := commaRegex.Split(profile, -1)
	profiles = append([]string{defaultProfileKey}, profiles...)

	profileProperties := make(map[string]interface{})
	// for each profiles, create a union of the configuration
	for _, profile := range profiles {
		if applicationMetadata, err := appCtx.getConfigFileMetaByProfileAndContext(profile, context); err != nil {
			log.Errorf("Error loading %s profile, %v", profile, err)

		} else {
			// To store the keys in slice in sorted order
			var keys []int
			for k := range applicationMetadata {
				keys = append(keys, int(k))
			}
			sort.Ints(keys)
			log.Debugf("keys:%+v", keys)

			// To perform the opertion you want
			// THINK ABOUT THIS.  ESP the order.  this is doing default classpath, default external, profile classpath, profile external
			props := make(map[string]interface{}, 0)
			for _, k := range keys {
				log.Debugf("merging key:%d", k)
				log.Debugf("%+v", applicationMetadata[int8(k)])
				props = loadFromFile(applicationMetadata[int8(k)])
				log.Debugf("props : %+v", props)
				if len(profileProperties) == 0 {
					profileProperties = props
				} else {
					profileProperties = mergeMaps(profileProperties, props)
				}
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
	v := viper.New()
	v.SetConfigName(configName)                     // name of config file (without extension)
	v.SetConfigType(fileMetadata.ConfigurationType) // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath(path)                           // path to look for the config file in
	err := v.ReadInConfig()                         // Find and read the config file
	if err != nil {                                 // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}
	return v.AllSettings()
}

func (appCtx *Pruner) getConfigFileMetaByProfileAndContext(profile string, context string) (map[int8]model.JavaConfigFileMetadata, error) {
	profileConfigFiles := make(map[int8]model.JavaConfigFileMetadata)
	for loadOrder, fileList := range appCtx.ConfigFiles {
		log.Debugf("load order :%d, files : %+v", loadOrder, fileList)
		for _, file := range fileList {

			if file.ApplicationContext == context && file.Profile == profile {
				profileConfigFiles[loadOrder] = file
			}

		}
	}
	if len(profileConfigFiles) > 0 {
		log.Debugf("configFiles : %+v", profileConfigFiles)
		return profileConfigFiles, nil
	}
	return nil, fmt.Errorf("config not found for profile:%s and context:%s ", profile, context)
}

// mergeMaps will merge map m2 onto map m1
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
