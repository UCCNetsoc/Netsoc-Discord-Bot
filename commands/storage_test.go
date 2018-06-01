package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteToStorage(t *testing.T) {
	var readStructure = map[string]string{}
	writeStructure := map[string]string{
		"test": "test1",
		"key":  "value",
	}

	if err := WriteToStorage("../storage/test.json", writeStructure); err != nil {
		t.Errorf("WriteToStorage error: %s", err)
	}

	if err = LoadFromStorage("../storage/test.json", &readStructure); err != nil {
		t.Errorf("LoadFromStorage error: %s", err)
	}

	assert.Equal(t, writeStructure, readStructure)
}
