package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zchase/jacuik/pkg/jacuik_config"
	"github.com/zchase/jacuik/pkg/terminal"
	"github.com/zchase/jacuik/pkg/utils"
)

var newProjectCmd = &cobra.Command{
	Use:   "new-project",
	Short: "Create a new project.",
	Long:  `Create a new Jacuik project in an empty directory.`,
	Run:   createNewProject,
}

func createNewProject(cmd *cobra.Command, args []string) {
	// Check the working directory is empty and if it isn't throw
	// an error.
	isWorkingDirectoryEmpty, err := utils.IsCurrentDirectoryEmpty()
	utils.IfErrorExit(err, "could determine if directory is empty")
	if !isWorkingDirectoryEmpty {
		utils.ThrowError("Current working directory is not empty.\n\nPlease switch to an empty directory to create a project.\n")
	}

	// Project Name
	defaultProjectName, err := utils.GetWorkingDirectoryName()
	utils.IfErrorExit(err, "couldn't read directroy name")

	projectName, err := terminal.NewTextPrompt("What is the name of your project?", defaultProjectName)
	utils.IfErrorExit(err, "couldn't set project name")

	// Project Description
	defaultProjectDescription := "A simple jacuik application."
	projectDescription, err := terminal.NewTextPrompt("How would you describe your project?", defaultProjectDescription)
	utils.IfErrorExit(err, "couldn't set project description")

	appConfig := &jacuik_config.AppConfig{
		Name:        projectName,
		Description: projectDescription,
	}

	// Schema file type
	schemaFileType, err := terminal.NewChoicePrompt("How would you like to author your config?", []string{"yaml", "json"})
	utils.IfErrorExit(err, "couldn't set config language")

	err = appConfig.WriteOutConfigFile(schemaFileType)
	utils.IfErrorExit(err, "couldn't write out config file")

	fmt.Println("✅ Project created.")
}

func init() {
	RootCmd.AddCommand(newProjectCmd)
}
