package jacuik_config

import (
	"fmt"
	"os"

	"github.com/zchase/jacuik/pkg/utils"
)

type ServiceConfig struct {
	Name             string `yaml:"name" json:"name"`
	PathToDockerfile string `yaml:"path" json:"path"`
	Public           bool   `yaml:"public" json:"public"`
}

type AppConfig struct {
	Name        string          `yaml:"name" json:"name"`
	Description string          `yaml:"description" json:"description"`
	Services    []ServiceConfig `yaml:"services" json:"services"`
}

func (a *AppConfig) WriteOutConfigFile(typ string) error {
	switch typ {
	case "yaml":
		return utils.WriteYAMLFile("schema.yaml", a)
	case "json":
		return utils.WriteJSONFile("schema.json", a)
	default:
		return fmt.Errorf("Unknown file type supplied [%s].", typ)
	}
}

func (a *AppConfig) AddService(service ServiceConfig) {
	a.Services = append(a.Services, service)
}

func ParseJacuikConfig() (*AppConfig, string, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	directoryContents, err := utils.ReadDirectoryContents(currentDirectory)
	if err != nil {
		return nil, "", err
	}

	var config *AppConfig
	var configType string
	for _, directoryItem := range directoryContents {
		switch directoryItem {
		case "schema.yaml":
			configType = "yaml"
			filePath := fmt.Sprintf("%s/schema.yaml", currentDirectory)
			config, err = utils.ParseYAMLFile[AppConfig](filePath)
			if err != nil {
				return nil, "", err
			}
		case "schema.yml":
			configType = "yaml"
			filePath := fmt.Sprintf("%s/schema.yml", currentDirectory)
			config, err = utils.ParseYAMLFile[AppConfig](filePath)
			if err != nil {
				return nil, "", err
			}
		case "schema.json":
			configType = "json"
			filePath := fmt.Sprintf("%s/schema.json", currentDirectory)
			config, err = utils.ParseJSONFile[AppConfig](filePath)
			if err != nil {
				return nil, "", err
			}
		}
	}

	if config == nil {
		return nil, "", fmt.Errorf("The schema file is empty or was not found. Please ensure you are in a Jacuik project.")
	}

	return config, configType, nil
}
