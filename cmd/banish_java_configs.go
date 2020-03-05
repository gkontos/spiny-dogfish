package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/jeremywohl/flatten"
	"github.com/wolfeidau/unflatten"
	"gopkg.in/yaml.v2"
)

type MatchingKeys struct {
	profileMatches []string
	sharedValue    interface{}
}

type ProfileProperties struct {
	profile        string
	flatProperties map[string]interface{}
	keySet         mapset.Set
	changes        map[string]ChangeSet
}

type ChangeSet struct {
	oldValue interface{}
	newValue interface{}
	message  string
	delete   bool
	profile  string
}

func (env *Organizer) BanishProperties() {
	if runProfile, err := PromptString("Profiles to Consolidate (semi-colon separated list of profiles.  Ie: dev; prod)"); err != nil {
		fmt.Printf("Error: %v", err)
		fmt.Println("")
	} else {
		profiles := strings.Split(strings.ReplaceAll(runProfile, " ", ""), ";")
		for _, context := range fileNames {
			profileProperties, changes := env.intersectProfileAndContext(profiles, context)
			for _, newProperties := range profileProperties {
				newProperties.flatProperties = applyChanges(newProperties.flatProperties, newProperties.changes)

			}
			outputToFiles(profileProperties, context)
			outputChanges(changes, context)
		}
	}
}

func (env *Organizer) intersectProfileAndContext(profiles []string, context string) ([]ProfileProperties, []ChangeSet) {
	collectedProfiles := make(map[string]map[string]interface{})
	profileSlice := make([]map[string]interface{}, 0)
	for _, profile := range profiles {
		collectedProfiles[profile] = env.unionProfileAndContext(profile, context)
		if profile != DEFAULT_PROFILE {
			profileSlice = append(profileSlice, collectedProfiles[profile])
		}
	}

	profileProperties := getFlatProperties(collectedProfiles)

	// GET A SET OF ALL THE PROPERTIES WHICH ARE SET IN ALL PROFILES
	keysetIntersection := getPropertyIntersection(profileProperties, DEFAULT_PROFILE)

	profileProperties, changes := decorateWithChanges(profileProperties, keysetIntersection)

	// for the key intersections, if values match the default profile, they should be removed
	return profileProperties, changes

}

func mostMatches(matches []MatchingKeys) MatchingKeys {
	max := 0
	index := 0
	for i, match := range matches {
		if len(match.profileMatches) > max {
			max = len(match.profileMatches)
			index = i
		}
	}
	fmt.Printf("index w/ most matches is %d ", index)
	return matches[index]
}

// addToMatchingValuesSlice if there is already a matchingKay with value, add the profile to the exiting match; otherwise add a new match to the slice
func addToMatchingValuesSlice(value interface{}, profile string, matchingValues []MatchingKeys) []MatchingKeys {
	for i, match := range matchingValues {
		if match.sharedValue == value {
			match.profileMatches = append(match.profileMatches, profile)
			matchingValues[i] = match
			return matchingValues
		}
	}
	match := MatchingKeys{}
	match.profileMatches = []string{profile}
	match.sharedValue = value
	matchingValues = append(matchingValues, match)
	return matchingValues
}

// findSmallestSetProfile will find and return the profile with the smallest set of properties
func findSmallestSetProfile(profiles []ProfileProperties, exclude string) ProfileProperties {
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
	return ProfileProperties{}
}

func getFlatProperties(collectedProfiles map[string]map[string]interface{}) []ProfileProperties {
	profileProperties := make([]ProfileProperties, 0)
	for k, v := range collectedProfiles {

		if flatProfile, err := flatten.Flatten(v, "", flatten.DotStyle); err != nil {
			fmt.Printf("error flatteniing map %v", err)
		} else {
			propertyKeys := mapset.NewSet()
			for key := range flatProfile {
				propertyKeys.Add(key)
			}
			profileProperty := ProfileProperties{}
			profileProperty.keySet = propertyKeys
			profileProperty.profile = k
			profileProperty.flatProperties = flatProfile
			profileProperties = append(profileProperties, profileProperty)
		}

	}
	return profileProperties
}

func getPropertyIntersection(flatProfileProperties []ProfileProperties, excludeProfile string) mapset.Set {
	smallestSetProfile := findSmallestSetProfile(flatProfileProperties, DEFAULT_PROFILE)
	keysetIntersection := smallestSetProfile.keySet
	for _, v := range flatProfileProperties {
		if v.profile != smallestSetProfile.profile && v.profile != excludeProfile {
			keysetIntersection = keysetIntersection.Intersect(v.keySet)
		}
	}
	return keysetIntersection
}

func decorateWithChanges(profileProperties []ProfileProperties, keysetIntersection mapset.Set) ([]ProfileProperties, []ChangeSet) {
	changes := make([]ChangeSet, 0)
	it := keysetIntersection.Iterator()
	for elem := range it.C {
		fmt.Printf("DECORATING FOR KEY %s", elem.(string))
		fmt.Println("")
		matchingValues := make([]MatchingKeys, 0)
		for _, profileProperty := range profileProperties {
			value, ok := profileProperty.flatProperties[elem.(string)]
			if !ok {
				value = ""
			}
			fmt.Printf("adding matching value %v to slice for %s", value, profileProperty.profile)
			fmt.Println("")

			matchingValues = addToMatchingValuesSlice(value, profileProperty.profile, matchingValues)

		}

		// THE RULES BELOW APPLY B/C we are looking only at properties which intersect across all files
		if len(matchingValues) < len(profileProperties) {
			// if all properties are equal ; set the default property to the found value and mark the property for removal from all profiles

			// if some properties are equal; mark the default property for update and mark the property for removal in equivalent profiles
			matchingValue := mostMatches(matchingValues)

			allFileMessage := fmt.Sprintf("The property %s is equivalent across profiles %s."+
				"The shared value of %v is being added to the default file.",
				elem.(string), strings.Join(matchingValue.profileMatches, ","), matchingValue.sharedValue)

			change := ChangeSet{}
			change.delete = false
			change.newValue = matchingValues[0].sharedValue
			change.profile = DEFAULT_PROFILE
			change.message = allFileMessage
			fmt.Printf("BEFORE: %+v", profileProperties)
			profileProperties = setProfilePropertyChange(profileProperties, elem.(string), change, DEFAULT_PROFILE)
			fmt.Printf("AFTER: %+v", profileProperties)
			changes = append(changes, change)

			for _, profile := range matchingValue.profileMatches {
				change := ChangeSet{}
				change.delete = true
				change.oldValue = matchingValues[0].sharedValue
				change.profile = profile
				change.message = allFileMessage
				profileProperties = setProfilePropertyChange(profileProperties, elem.(string), change, profile)
				fmt.Printf("%s %+v", profile, profileProperties)
				changes = append(changes, change)
			}

		} else {
			// add a comment to the default propertyfile?  this property is set in all files, but a different value exists in all files.  This might be a mistake?
			var msg strings.Builder
			msg.WriteString(fmt.Sprintf("The property %s is set with different values on all profiles", elem.(string)))
			for _, v := range matchingValues {
				msg.WriteString(fmt.Sprintf(" {Profile : %s => %v} ", strings.Join(v.profileMatches, ","), v.sharedValue))
			}
			change := ChangeSet{}
			change.message = msg.String()
			change.delete = false
			profileProperties = setProfilePropertyChange(profileProperties, elem.(string), change, DEFAULT_PROFILE)
			fmt.Printf("DEFAULT %+v", profileProperties)

			changes = append(changes, change)

		}
	}
	return profileProperties, changes
}

func setProfilePropertyChange(profileProperties []ProfileProperties, propertyKey string, change ChangeSet, profile string) []ProfileProperties {
	updatedProperties := profileProperties[:0]
	for _, profileProperty := range profileProperties {
		if profileProperty.profile == profile {
			if profileProperty.changes == nil {
				profileProperty.changes = make(map[string]ChangeSet)
			}
			profileProperty.changes[propertyKey] = change

		}
		updatedProperties = append(updatedProperties, profileProperty)
	}
	return updatedProperties
}

func applyChanges(flatProperties map[string]interface{}, changes map[string]ChangeSet) map[string]interface{} {
	for k, v := range changes {
		if v.delete {
			delete(flatProperties, k)
		} else {
			if v.newValue != nil {
				flatProperties[k] = v.newValue
			}
		}
	}
	return flatProperties
}

func outputToFiles(profileProperties []ProfileProperties, context string) {
	for _, properties := range profileProperties {
		// properties.profile
		propertiesFileName := fmt.Sprintf("%s-%s-banished.yml", context, properties.profile)
		changesFileName := fmt.Sprintf("%s-%s-banished-changes.txt", context, properties.profile)
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

func outputChanges(changes []ChangeSet, context string) {

	f, _ := os.Create(fmt.Sprintf("change-%s-set.txt", context))
	defer f.Close()
	for _, line := range changes {
		_, _ = f.WriteString(line.message + "\n")
	}

	f.Sync()

}
