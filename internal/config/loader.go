package config

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var values *MainConfig

func LoadConfig(configPath string) (*MainConfig, error) {
	// Check for config and create default
	_, err := os.Stat(configPath)
	exists := err == nil || !os.IsNotExist(err)
	if !exists {
		log.Println("Creating default config")
		configData, err := yaml.Marshal(NewMainConfig())
		if err != nil {
			return nil, err
		}
		configFile, err := os.Create(configPath)
		if err != nil {
			return nil, err
		}
		_, err = configFile.Write(configData)
		if err != nil {
			return nil, err
		}
		err = configFile.Close()
		if err != nil {
			return nil, err
		}
	}

	// Revalidate after writing
	stat, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	// Check if dir, then add all valid config files inside, otherwise add the single config file
	configFiles := []string{}
	if stat.IsDir() {
		log.Printf("Loading config from directory %v\n", configPath)
		potentialFiles, err := os.ReadDir(configPath)
		if err != nil {
			return nil, err
		}
		for _, file := range potentialFiles {
			if file.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext != ".yaml" && ext != ".yml" {
				continue
			}
			configFiles = append(configFiles, path.Join(configPath, file.Name()))
		}
		sort.Strings(configFiles)
	} else {
		log.Printf("Loading config from %v\n", configPath)
		configFiles = append(configFiles, stat.Name())
	}

	// Iterate over all found files and deep merge them in order
	fullMap := map[string]interface{}{}
	for _, f := range configFiles {
		newMap := map[string]interface{}{}
		data, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Error while reading a config file: %v\n", f)
			log.Println(err)
			continue
		}
		if err := yaml.Unmarshal(data, &newMap); err != nil {
			log.Printf("Error while parsing a config file: %v\n", f)
			log.Println(err)
			continue
		}
		fullMap = deepMerge(fullMap, newMap)
	}

	// Convert from map to struct
	mMap, err := yaml.Marshal(fullMap)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(mMap, &values); err != nil {
		return nil, err
	}

	// Check for environment variables to override config
	values.envInit()

	return values, nil
}

// From https://stackoverflow.com/a/70291996
func deepMerge(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = deepMerge(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
