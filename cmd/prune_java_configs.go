package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set"
	log "github.com/gkontos/bivalve-chronicles"
	"github.com/jeremywohl/flatten"
	"github.com/wolfeidau/unflatten"
	"gopkg.in/yaml.v2"
)

type matchingKeys struct {
	profileMatches []string
	sharedValue    interface{}
}

type profilePropertyPruner struct {
	profile        string
	flatProperties map[string]interface{}
	keySet         mapset.Set
	changes        map[string]changeSet
}

type changeSet struct {
	oldValue interface{}
	newValue interface{}
	message  string
	delete   bool
	profile  string
}

// PruneProperties will load config files, compact duplicate values, and output updated configuration files
func (env *Pruner) PruneProperties() {
	if runProfile, err := promptString("Profiles to Consolidate (semi-colon separated list of profiles.  Ie: dev; prod)"); err != nil {
		log.Errorf("Error: %v", err)
	} else {
		profiles := strings.Split(strings.ReplaceAll(runProfile, " ", ""), ";")
		for _, context := range fileNames {
			profileProperties, changes := env.intersectProfileAndContext(profiles, context)

			for _, newProperties := range profileProperties {
				log.Debugf("APPLYING CHANGES TO PROFILE: %s", newProperties.profile)
				newProperties.flatProperties = applyChanges(newProperties.flatProperties, newProperties.changes)
			}

			outputToFiles(profileProperties, context)
			outputChanges(changes, context)
		}
	}
}

func (env *Pruner) intersectProfileAndContext(profiles []string, context string) ([]profilePropertyPruner, []changeSet) {
	collectedProfiles := make(map[string]map[string]interface{})
	profiles = append(profiles, defaultProfileKey)
	for _, profile := range profiles {
		collectedProfiles[profile] = env.unionProfileAndContext(profile, context)
	}

	profileProperties := getFlatProperties(collectedProfiles)

	// GET A SET OF ALL THE PROPERTIES WHICH ARE SET IN ALL PROFILES
	keysetIntersection := getPropertyIntersection(profileProperties, defaultProfileKey)

	profileProperties, changes := decorateWithChanges(profileProperties, keysetIntersection)
	// for the key intersections, if values match the default profile, they should be removed
	return profileProperties, changes

}

func mostMatches(matches []matchingKeys) matchingKeys {
	max := 0
	index := 0
	for i, match := range matches {
		if len(match.profileMatches) > max {
			max = len(match.profileMatches)
			index = i
		}
	}
	return matches[index]
}

// addToMatchingValuesSlice if there is already a matchingKay with value, add the profile to the exiting match; otherwise add a new match to the slice
func addToMatchingValuesSlice(value interface{}, profile string, matchingValues []matchingKeys) []matchingKeys {
	for i, match := range matchingValues {
		if match.sharedValue == value {
			match.profileMatches = append(match.profileMatches, profile)
			matchingValues[i] = match
			return matchingValues
		}
	}
	match := matchingKeys{}
	match.profileMatches = []string{profile}
	match.sharedValue = value
	matchingValues = append(matchingValues, match)
	return matchingValues
}

// findSmallestSetProfile will find and return the profile with the smallest set of properties
func findSmallestSetProfile(profiles []profilePropertyPruner, exclude string) profilePropertyPruner {
	minsize := -1
	minindex := -1
	for i, profile := range profiles {
		if profile.profile != exclude {
			if minsize == -1 || len(profile.flatProperties) < minsize {
				minsize = len(profile.flatProperties)
				minindex = i
			}
		}
	}
	if minindex > -1 {
		return profiles[minindex]
	}
	return profilePropertyPruner{}
}

func getFlatProperties(collectedProfiles map[string]map[string]interface{}) []profilePropertyPruner {
	profileProperties := make([]profilePropertyPruner, 0)
	for k, v := range collectedProfiles {

		if flatProfile, err := flatten.Flatten(v, "", flatten.DotStyle); err != nil {
			log.Errorf("error flatteniing map %v", err)
		} else {
			propertyKeys := mapset.NewSet()
			for key := range flatProfile {
				propertyKeys.Add(key)
			}
			profileProperty := profilePropertyPruner{}
			profileProperty.keySet = propertyKeys
			profileProperty.profile = k
			profileProperty.flatProperties = flatProfile
			profileProperties = append(profileProperties, profileProperty)
		}

	}
	return profileProperties
}

func getPropertyIntersection(flatProfileProperties []profilePropertyPruner, excludeProfile string) mapset.Set {
	smallestSetProfile := findSmallestSetProfile(flatProfileProperties, defaultProfileKey)
	keysetIntersection := smallestSetProfile.keySet
	for _, v := range flatProfileProperties {
		if v.profile != smallestSetProfile.profile && v.profile != excludeProfile {
			keysetIntersection = keysetIntersection.Intersect(v.keySet)
		}
	}
	return keysetIntersection
}

func decorateWithChanges(profileProperties []profilePropertyPruner, keysetIntersection mapset.Set) ([]profilePropertyPruner, []changeSet) {
	changes := make([]changeSet, 0)
	it := keysetIntersection.Iterator()
	for elem := range it.C {
		log.Debugf("DECORATING FOR KEY %s", elem.(string))

		matchingValues := make([]matchingKeys, 0)
		for _, profileProperty := range profileProperties {
			if profileProperty.profile != defaultProfileKey {
				value, ok := profileProperty.flatProperties[elem.(string)]
				if !ok {
					value = ""
				}

				matchingValues = addToMatchingValuesSlice(value, profileProperty.profile, matchingValues)
			}
		}

		// THE RULES BELOW APPLY B/C we are looking only at properties which intersect across all files
		if len(matchingValues) < len(profileProperties) {
			// if all properties are equal ; set the default property to the found value and mark the property for removal from all profiles

			// if some properties are equal; mark the default property for update and mark the property for removal in equivalent profiles
			matchingValue := mostMatches(matchingValues)

			allFileMessage := fmt.Sprintf("The property %s is equivalent across profiles %s."+
				"The shared value of %v is being added to the default file.",
				elem.(string), strings.Join(matchingValue.profileMatches, ","), matchingValue.sharedValue)

			change := changeSet{}
			change.delete = false
			change.newValue = matchingValues[0].sharedValue
			change.profile = defaultProfileKey
			change.message = allFileMessage
			profileProperties = setProfilePropertyChange(profileProperties, elem.(string), change, defaultProfileKey)
			changes = append(changes, change)

			for _, profile := range matchingValue.profileMatches {
				change := changeSet{}
				change.delete = true
				change.oldValue = matchingValues[0].sharedValue
				change.profile = profile
				change.message = allFileMessage
				profileProperties = setProfilePropertyChange(profileProperties, elem.(string), change, profile)
				changes = append(changes, change)
			}

		} else {
			// add a comment to the default propertyfile?  this property is set in all files, but a different value exists in all files.  This might be a mistake?
			var msg strings.Builder
			msg.WriteString(fmt.Sprintf("The property %s is set with different values on all profiles", elem.(string)))
			for _, v := range matchingValues {
				msg.WriteString(fmt.Sprintf(" {Profile : %s => %v} ", strings.Join(v.profileMatches, ","), v.sharedValue))
			}
			change := changeSet{}
			change.message = msg.String()
			change.delete = false
			profileProperties = setProfilePropertyChange(profileProperties, elem.(string), change, defaultProfileKey)

			changes = append(changes, change)

		}
	}
	return profileProperties, changes
}

func setProfilePropertyChange(profileProperties []profilePropertyPruner, propertyKey string, change changeSet, profile string) []profilePropertyPruner {
	updatedProperties := profileProperties[:0]
	for _, profileProperty := range profileProperties {
		if profileProperty.profile == profile {
			if profileProperty.changes == nil {
				profileProperty.changes = make(map[string]changeSet)
			}
			profileProperty.changes[propertyKey] = change

		}
		updatedProperties = append(updatedProperties, profileProperty)
	}
	return updatedProperties
}

func applyChanges(flatProperties map[string]interface{}, changes map[string]changeSet) map[string]interface{} {
	log.Debugf("Start count %d", len(flatProperties))
	expectedcount := len(flatProperties)
	deletecount := 0
	addcount := 0
	for k, v := range changes {
		if v.delete {
			delete(flatProperties, k)
			expectedcount--
			deletecount++
		} else if v.newValue != nil {
			if _, ok := flatProperties[k]; !ok {
				// if the element does not already exist in the map, this is a new value which will be added
				expectedcount++
				addcount++
			}
			flatProperties[k] = v.newValue

		}
	}
	log.Debugf("deleted: %d, added: %d", deletecount, addcount)
	log.Debugf("End count %d", len(flatProperties))
	log.Debugf("Expected %d, Saw %d", expectedcount, len(flatProperties))
	return flatProperties
}

func outputToFiles(profileProperties []profilePropertyPruner, context string) {
	for _, properties := range profileProperties {
		propertiesFileName := fmt.Sprintf("%s-%s-pruned.yml", context, properties.profile)
		changesFileName := fmt.Sprintf("%s-%s-pruned-changes.txt", context, properties.profile)
		expandedProperties := unflatten.Unflatten(properties.flatProperties, func(k string) []string { return strings.Split(k, ".") })
		ymlString, _ := yaml.Marshal(expandedProperties)
		ioutil.WriteFile(propertiesFileName, ymlString, 0644)

		f, _ := os.Create(changesFileName)
		defer f.Close()
		for _, line := range properties.changes {
			_, _ = f.WriteString(line.message + "\n")
		}

		f.Sync()
	}
}

func outputChanges(changes []changeSet, context string) {

	f, _ := os.Create(fmt.Sprintf("change-set-%s.txt", context))
	defer f.Close()
	for _, line := range changes {
		_, _ = f.WriteString(line.message + "\n")
	}

	f.Sync()

}
