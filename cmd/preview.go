package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zchase/jacuik/pkg/infrastructure"
	"github.com/zchase/jacuik/pkg/jacuik_config"
	"github.com/zchase/jacuik/pkg/terminal"
	"github.com/zchase/jacuik/pkg/utils"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview a deployment.",
	Long:  `Preview a deployment.`,
	Run:   preview,
}

func preview(cmd *cobra.Command, args []string) {
	// fmt.Println("preview")

	// pbSteps := []terminal.ProgressBarStep{
	// 	{
	// 		Label: "Configuring preview...",
	// 	},
	// 	{
	// 		Label: "Running preview...",
	// 	},
	// }

	// pb, err := terminal.NewProgressBar(terminal.ProgressBarArgs{
	// 	Steps: pbSteps,
	// })
	// utils.IfErrorExit(err, "could not create progress bar")

	// time.Sleep(time.Second * 2)
	// pb.IncrementStep()

	// time.Sleep(time.Second * 2)
	// pb.Done()

	// //time.Sleep(time.Second * 5)

	// fmt.Println("done")

	fmt.Print("Preview of application updates:\n\n")

	config, _, err := jacuik_config.ParseJacuikConfig()
	utils.IfErrorExit(err, "couldn't parse config")

	infra := infrastructure.NewInfrastructureHandler("jacuik-demo", config)

	view := terminal.NewView(infra.Preview)
	err = view.Start()
	utils.IfErrorExit(err, "error running preview")
}

func init() {
	RootCmd.AddCommand(previewCmd)
}
