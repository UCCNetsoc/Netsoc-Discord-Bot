package commands

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createSampleAliases(testAliasMap map[string]*alias) (func(), error) {
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
	aliasMapWant := map[string]*alias{
		"test": &alias{Value: "test1", Kind: kindOTHER},
		"key":  &alias{Value: "value", Kind: kindOTHER},
	}
	cleanup, err := createSampleAliases(aliasMapWant)
	if err != nil {
		t.Fatalf("Failed to create sample alias file: %s", err)
	}
	defer cleanup()

	aliasMap = make(map[string]*alias)
	if err := loadFromStorage(); err != nil {
		t.Errorf("LoadFromStorage error: %s", err)
	}

	assert.Equal(t, aliasMap, aliasMapWant)
}

func TestWithAliasCommands(t *testing.T) {
	testAliases := map[string]*alias{
		"testKey":  &alias{Value: "testValue", Kind: kindOTHER},
		"testKey1": &alias{Value: "testValue1", Kind: kindOTHER},
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

	for aliasKey := range testAliases {
		c, ok := commMap[aliasKey]
		if !ok {
			t.Fatal("Command map doesn't contain a command named with the alias key")
		}

		aliasCommandFunc := c.(*textCommand).commandFunc()
		got, err := aliasCommandFunc(context.Background(), []string{aliasKey})
		if err != nil {
			t.Fatalf("Failed to run internal alias command function: %s", err)
		}

		if want := testAliases[aliasKey].Value; got != want {
			t.Errorf("Wrong alias value: got %q; want %q", got, want)
		}
	}
}

func TestAliasCommand_SetNewAlias(t *testing.T) {
	testAliases := make(map[string]*alias)
	cleanup, err := createSampleAliases(testAliases)
	if err != nil {
		t.Fatalf("Failed to create sample alias file: %s", err)
	}
	defer cleanup()

	tests := []struct {
		key, value string
		kind       aliasKind
	}{
		{"test1", "test1value", kindOTHER},
		{"test2", "https://media.giphy.com/media/3oz8xODcLLAxb8Qyju/giphy.gif", kindIMAGE},
	}
	for i, ts := range tests {
		if _, err := aliasCommand(context.Background(), []string{"alias", ts.key, ts.value}); err != nil {
			t.Fatalf("aliasCommand error: %s", err)
		}

		got, inAliasMap := aliasMap[ts.key]
		if !inAliasMap {
			t.Errorf("%d) New alias not in aliasMap", i)
		}
		if got.Value != ts.value {
			t.Errorf("%d) New alias has wrong value: got %q; want %q", i, got.Value, ts.value)
		}
		if got.Kind != ts.kind {
			t.Errorf("%d) New alias has wrong kind: got %q; want %q", i, got.Kind, ts.kind)
		}

		if _, inCommMap := commMap[ts.key]; !inCommMap {
			t.Errorf("%d) New alias not in commMap", i)
		}

		aliasMap = make(map[string]*alias)
		if err := loadFromStorage(); err != nil {
			t.Errorf("%d) Failed to load alias storage file: %s", i, err)
			continue
		}
		got, inAliasMap = aliasMap[ts.key]
		if !inAliasMap {
			t.Errorf("%d) New alias not written to storage", i)
		}
		if got.Value != ts.value {
			t.Errorf("%d) New alias not written to storage correctly. Got %q; want %q", i, got.Value, ts.value)
		}
	}
}

func TestUnAliasCommand(t *testing.T) {
	var (
		removeAlias = "unaliasCommandTest"
		removeValue = &alias{Value: "some val"}
	)
	testAliases := map[string]*alias{removeAlias: removeValue}
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

	aliasMap = make(map[string]*alias)
	if err := loadFromStorage(); err != nil {
		t.Fatalf("Failed to load alias storage file: %s", err)
	}
	if _, inAliasMap := aliasMap[removeAlias]; inAliasMap {
		t.Error("Alias not removed from storage file")
	}
}
