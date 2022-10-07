package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zchase/jacuik/pkg/jacuik_config"
	"github.com/zchase/jacuik/pkg/terminal"
	"github.com/zchase/jacuik/pkg/utils"
)

var newServiceCmd = &cobra.Command{
	Use:   "new-service",
	Short: "Create a new service.",
	Long:  `Create a new service in a Jacuik project.`,
	Run:   createNewService,
}

func createNewService(cmd *cobra.Command, args []string) {
	appConfig, configType, err := jacuik_config.ParseJacuikConfig()
	utils.IfErrorExit(err, "couldn't successfully parse config")

	serviceName, err := terminal.NewTextPrompt("What is the name of your new service?", "")
	utils.IfErrorExit(err, "couldn't set service name")

	public, err := terminal.NewChoicePrompt("Is this a public service?", []string{"true", "false"})
	utils.IfErrorExit(err, "couldn't set service public setting")

	// Create the service directory and Dockerfile
	wd, err := os.Getwd()
	utils.IfErrorExit(err, "couldn't get current working directory")

	serviceDirPath := fmt.Sprintf("%s/%s", wd, serviceName)
	err = utils.CreateDirectory(serviceDirPath)
	utils.IfErrorExit(err, "couldn't create service directory")

	dockerfilePath := fmt.Sprintf("%s/Dockerfile", serviceDirPath)
	err = utils.WriteFile(dockerfilePath, "")
	utils.IfErrorExit(err, "couldn't create service Dockerfile")

	isServicePublic := true
	if public == "false" {
		isServicePublic = false
	}

	newService := jacuik_config.ServiceConfig{
		Name:             serviceName,
		PathToDockerfile: ".",
		Public:           isServicePublic,
	}

	appConfig.AddService(newService)

	err = appConfig.WriteOutConfigFile(configType)
	utils.IfErrorExit(err, "couldn't update config file")

	fmt.Println("✅ Service scaffolding created.")
}

func init() {
	RootCmd.AddCommand(newServiceCmd)
}
