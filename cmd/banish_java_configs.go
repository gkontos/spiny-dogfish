package cmd

import (
	"container/list"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

type MatchingKeys struct {
	profileMatches []string
	sharedValue    interface{}
	mapPath        *list.List
}

type valueCount struct {
	value interface{}
	count int
}

func (env *Organizer) BanishProperties() {
	if runProfile, err := PromptString("Profiles to Consolidate (semi-colon separated list of profiles.  Ie: dev; prod)"); err != nil {
		fmt.Printf("Error: %v", err)
		fmt.Println("")
	} else {
		profiles := strings.Split(strings.ReplaceAll(runProfile, " ", ""), ";")
		for _, context := range fileNames {
			env.intersectProfileAndContext(profiles, context)
		}
	}
}

func (env *Organizer) intersectProfileAndContext(profiles []string, context string) map[string]interface{} {
	collectedProfiles := make(map[string]map[string]interface{})
	profileSlice := make([]map[string]interface{}, 0)
	for _, profile := range profiles {
		collectedProfiles[profile] = env.unionProfileAndContext(profile, context)
		if profile != DEFAULT_PROFILE {
			profileSlice = append(profileSlice, collectedProfiles[profile])
		}
	}
	intersectingKeys := make([]*list.List, 0)
	// NOTE: remove the default profile before running..
	intersectingKeys, _ = mapKeyIntersection(profileSlice, intersectingKeys, list.New())

	propertyMatches := make(map[string][]MatchingKeys)
	// for the key intersections, determine if there are any duplicate values
	for _, keyPath := range intersectingKeys {
		profileValueMap := make(map[string]interface{})
		for k, v := range collectedProfiles {
			profileValueMap[k] = getMapValueAtPath(v, keyPath)
		}

		values := make([]MatchingKeys, 0)
		for k, v := range profileValueMap {
			if k != DEFAULT_PROFILE {
				if index := valueSliceContains(v, values); index >= 0 {
					valueCounter := values[index]
					valueCounter.profileMatches = append(valueCounter.profileMatches, k)
					values[index] = valueCounter
				} else {
					valueCounter := MatchingKeys{}
					valueCounter.profileMatches = []string{k}
					valueCounter.sharedValue = v
					valueCounter.mapPath = keyPath
				}
			}
		}

		// there is at least one match
		if len(values) < len(profileSlice) {
			propertyMatches[RandStringBytesMaskImprSrcSB(25)] = values
		}

	}
	handleIntersection(collectedProfiles, propertyMatches)

	// for the key intersections, if values match the default profile, they should be removed

	return nil
}

func valueSliceContains(value interface{}, values []MatchingKeys) int {
	return -1
}

func getMapValueAtPath(haystack map[string]interface{}, keyPath *list.List) interface{} {
	itStart := keyPath.Front()
	var found interface{}
	found = haystack[itStart.Value.(string)]
	if isMap(found) {
		for {
			if el := itStart.Next(); el == nil {
				break
			} else {
				mapValue := found.(map[string]interface{})
				found = mapValue[el.Value.(string)]
				if !isMap(found) {
					break
				}
			}
		}
	}
	return found
}

// if a give key exists within all maps in the slice, return the value from all maps; otherwise return empty slice
func keyExistsInAllMaps(key string, otherMaps []map[string]interface{}) []map[string]interface{} {

	return nil
}

func mapKeyIntersection(maps []map[string]interface{}, keyPaths []*list.List, thisPath *list.List) ([]*list.List, *list.List) {
	smallestMapSize := 0
	var smallestMapIndex int
	baseComparisonMap := make(map[string]interface{})
	for index, value := range maps {
		if smallestMapSize == 0 || len(value) < smallestMapSize {
			smallestMapIndex = index
			smallestMapSize = len(value)
		}
	}
	baseComparisonMap = maps[smallestMapIndex]

	otherComparisonMaps := make([]map[string]interface{}, 0)
	for i, v := range maps {
		if i != smallestMapIndex {
			otherComparisonMaps = append(otherComparisonMaps, v)
		}
	}

	for k := range baseComparisonMap {
		if checkKey := keyExistsInAllMaps(k, otherComparisonMaps); checkKey != nil {
			// need something recursive here...?
			if isMap(baseComparisonMap[k]) {
				// we want to check the values across all maps and consolidate
				keyPaths = append(keyPaths, thisPath)

			} else {
				thisPath.PushBack(k)
				return mapKeyIntersection(checkKey, keyPaths, thisPath)

			}
		}
	}
	return keyPaths, nil
}

func isMap(value interface{}) bool {
	return (reflect.TypeOf(value).Kind() == reflect.Map)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits
)

var src = rand.NewSource(time.Now().UnixNano())

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
