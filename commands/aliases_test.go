package commands

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createSampleAliases(testAliasMap map[string]string) (func(), error) {
	old := aliasStorageFilepath
	aliasStorageFilepath = "test.json"
	aliasMap = testAliasMap

	if err := writeToStorage(); err != nil {
		return nil, fmt.Errorf("WriteToStorage error: %s", err)
	}
	return func() {
		os.Remove(aliasStorageFilepath)
		aliasStorageFilepath = old
	}, nil

}

func TestLoadFromStorage(t *testing.T) {
	aliasMapWant := map[string]string{
		"test": "test1",
		"key":  "value",
	}
	cleanup, err := createSampleAliases(aliasMapWant)
	if err != nil {
		t.Fatalf("Failed to create sample alias file: %s", err)
	}
	defer cleanup()

	aliasMap = make(map[string]string)
	if err := loadFromStorage(); err != nil {
		t.Errorf("LoadFromStorage error: %s", err)
	}

	assert.Equal(t, aliasMap, aliasMapWant)
}

func TestWithAliasCommands(t *testing.T) {
	testAliases := map[string]string{
		"testKey": "testValue",
	}
	cleanup, err := createSampleAliases(testAliases)
	if err != nil {
		t.Fatalf("Failed to create sample alias file: %s", err)
	}
	defer cleanup()

	commMap, err := withAliasCommands(commMap)
	if err != nil {
		t.Fatalf("withAliasCommands: %s", err)
	}

	c, ok := commMap["testKey"]
	if !ok {
		t.Fatal("Command map doesn't contain a command named witht he alias key")
	}

	aliasCommandFunc := c.(*textCommand).commandFunc()
	got, err := aliasCommandFunc(context.Background(), []string{"test1"})
	if err != nil {
		t.Fatalf("Failed to run internal alias command function: %s", err)
	}

	if got != "testValue" {
		t.Errorf("Want 'testValue' from alias function; got %q", got)
	}
}

func TestAliasCommand_SetNewAlias(t *testing.T) {
	testAliases := make(map[string]string)
	cleanup, err := createSampleAliases(testAliases)
	if err != nil {
		t.Fatalf("Failed to create sample alias file: %s", err)
	}
	defer cleanup()

	var (
		aliasKey   = "aliasCommandTest"
		aliasValue = "test2"
	)
	if _, err := aliasCommand(context.Background(), []string{"alias", aliasKey, aliasValue}); err != nil {
		t.Fatalf("aliasCommand error: %s", err)
	}

	gotValue, inAliasMap := aliasMap[aliasKey]
	if !inAliasMap {
		t.Error("New alias not in aliasMap")
	}
	if gotValue != aliasValue {
		t.Errorf("New alias has wrong value. Got %q; want %q", gotValue, aliasValue)
	}

	if _, inCommMap := commMap[aliasKey]; !inCommMap {
		t.Error("New alias not in commMap")
	}

	aliasMap = make(map[string]string)
	if err := loadFromStorage(); err != nil {
		t.Fatalf("Failed to load alias storage file: %s", err)
	}
	gotValue, inAliasMap = aliasMap[aliasKey]
	if !inAliasMap {
		t.Error("New alias not written to storage")
	}
	if gotValue != aliasValue {
		t.Errorf("New alias not written to storage correctly. Got %q; want %q", gotValue, aliasValue)
	}

}

func TestUnAliasCommand(t *testing.T) {
	var (
		removeAlias = "unaliasCommandTest"
		removeValue = "some val"
	)
	testAliases := map[string]string{removeAlias: removeValue}
	cleanup, err := createSampleAliases(testAliases)
	if err != nil {
		t.Fatalf("Failed to create sample alias file: %s", err)
	}
	defer cleanup()

	if _, err := unAliasCommand(context.Background(), []string{"unalias", removeAlias}); err != nil {
		t.Fatalf("aliasCommand error: %s", err)
	}

	if _, inAliasMap := aliasMap[removeAlias]; inAliasMap {
		t.Error("Alias not removed from aliasMap")
	}

	if _, inCommMap := commMap[removeAlias]; inCommMap {
		t.Error("Alias command not removed from commMap")
	}

	aliasMap = make(map[string]string)
	if err := loadFromStorage(); err != nil {
		t.Fatalf("Failed to load alias storage file: %s", err)
	}
	if _, inAliasMap := aliasMap[removeAlias]; inAliasMap {
		t.Error("Alias not removed from storage file")
	}
}
