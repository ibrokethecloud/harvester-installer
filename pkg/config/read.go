package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rancher/mapper/convert"

	"github.com/harvester/harvester-installer/pkg/util"
)

const (
	kernelParamPrefix       = "harvester"
	defaultConfigFileEnvVar = "HARVESTER_INSTALL_CONFIG"
)

// ReadConfig constructs a config by reading various sources
func ReadConfig() (HarvesterConfig, error) {
	result := NewHarvesterConfig()

	hcFromFile, err := ReadConfigFromFile()
	if err != nil {
		return *result, err
	}

	// Reading from file overrides kernel arguments
	if hcFromFile != nil {
		return *hcFromFile, nil
	}

	data, err := util.ReadCmdline(kernelParamPrefix)
	if err != nil {
		return *result, err
	}
	schema.Mapper.ToInternal(data)
	return *result, convert.ToObj(data, result)
}

//ReadConfigFromFile reads the config from an override file specified
//by the environment variable HARVESTER_INSTALL_CONFIG
func ReadConfigFromFile() (*HarvesterConfig, error) {
	val, ok := os.LookupEnv(defaultConfigFileEnvVar)
	if !ok {
		return nil, nil
	}

	f, err := os.Stat(val)
	if err != nil {
		return nil, err
	}

	if f.IsDir() {
		return nil, fmt.Errorf("file path %s is a directory, expected a file", val)
	}

	fileContent, err := ioutil.ReadFile(val)
	if err != nil {
		return nil, err
	}

	result := NewHarvesterConfig()
	err = yaml.Unmarshal(fileContent, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ToEnv(prefix string, obj interface{}) ([]string, error) {
	data, err := convert.EncodeToMap(obj)
	if err != nil {
		return nil, err
	}

	return mapToEnv(prefix, data), nil
}

func mapToEnv(prefix string, data map[string]interface{}) []string {
	var result []string
	for k, v := range data {
		keyName := strings.ToUpper(prefix + convert.ToYAMLKey(k))
		if data, ok := v.(map[string]interface{}); ok {
			subResult := mapToEnv(keyName+"_", data)
			result = append(result, subResult...)
		} else {
			result = append(result, fmt.Sprintf("%s=%v", keyName, v))
		}
	}
	return result
}
