package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// WriteToStorage stores given data in a json file
func WriteToStorage(filePath string, s interface{}) error {
	structure, err := json.Marshal(s)

	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %v", err)
	}

	if err = ioutil.WriteFile(filePath, structure, 0744); err != nil {
		return fmt.Errorf("Failed to write file: %v", err)
	}

	return nil
}

// LoadFromStorage loads a JSON file into a given interface
func LoadFromStorage(filePath string, s interface{}) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		WriteToStorage(filePath, s)
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %v", err)
	}

	if err := json.Unmarshal(file, &s); err != nil {
		return fmt.Errorf("failed to unmarshal configuration JSON: %s", err)
	}

	return nil
}
