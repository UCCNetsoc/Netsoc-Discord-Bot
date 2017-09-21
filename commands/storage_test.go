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

	err := WriteToStorage("../storage/test.json", writeStructure)
	if err != nil {
		t.Errorf("Err: %#v", err)
	}

	err = LoadFromStorage("../storage/test.json", &readStructure)
	if err != nil {
		t.Errorf("Err: %#v", err)
	}

	assert.Equal(t, writeStructure, readStructure)
}
