package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddToMatchingValuesSlice(t *testing.T) {
	matchingValues := make([]matchingKeys, 0)
	profile := "profile1"
	value := 4
	matchingValues = addToMatchingValuesSlice(value, profile, matchingValues)
	assert.EqualValues(t, 1, len(matchingValues))
	assert.EqualValues(t, 1, len(matchingValues[0].profileMatches))
	assert.EqualValues(t, value, matchingValues[0].sharedValue)

	profile = "profile2"
	matchingValues = addToMatchingValuesSlice(value, profile, matchingValues)
	assert.EqualValues(t, 1, len(matchingValues))
	assert.EqualValues(t, 2, len(matchingValues[0].profileMatches))
	assert.EqualValues(t, value, matchingValues[0].sharedValue)

	profile = "profile3"
	matchingValues = addToMatchingValuesSlice(value, profile, matchingValues)
	assert.EqualValues(t, 1, len(matchingValues))
	assert.EqualValues(t, 3, len(matchingValues[0].profileMatches))
	assert.EqualValues(t, value, matchingValues[0].sharedValue)

	profile = "profile4"
	value2 := "anotherValue"
	matchingValues = addToMatchingValuesSlice(value2, profile, matchingValues)
	assert.EqualValues(t, 2, len(matchingValues))
	assert.EqualValues(t, 3, len(matchingValues[0].profileMatches))
	assert.EqualValues(t, 1, len(matchingValues[1].profileMatches))
	assert.EqualValues(t, value2, matchingValues[1].sharedValue)

}

func TestSetProfilePropertyChange(t *testing.T) {
	profile := profilePropertyPruner{}
	profile.profile = "profile1"
	change := changeSet{}
	change.delete = false
	change.message = "message of change"
	change.profile = profile.profile
	profileProperties := []profilePropertyPruner{profile}
	key := "changed.propery"
	updatedProperties := setProfilePropertyChange(profileProperties, key, change, profile.profile)

	assert.EqualValues(t, 1, len(updatedProperties))
	updatedProfile := updatedProperties[0]
	assert.EqualValues(t, 1, len(updatedProfile.changes))

	updatedProperties = setProfilePropertyChange(profileProperties, key, change, "other profile")

	assert.EqualValues(t, 1, len(updatedProperties))
	updatedProfile = updatedProperties[0]
	assert.EqualValues(t, 1, len(updatedProfile.changes))

}

func TestUnionProfileAndContext(t *testing.T) {
	profile := "dev"
	context := "application"

}
