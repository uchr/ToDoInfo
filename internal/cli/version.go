package cli

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var Version = "2.0.0" // CLI version

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display the current version of ToDo Info CLI",
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	// Create a beautiful version display
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(
		pterm.LightCyan("┌─────────────────────────────────────┐"),
		pterm.LightCyan("│")+" "+pterm.LightMagenta("ToDo Info CLI")+"                     "+pterm.LightCyan("│"),
		pterm.LightCyan("│")+" "+pterm.Gray("Beautiful Microsoft ToDo Analysis")+" "+pterm.LightCyan("│"),
		pterm.LightCyan("│")+"                                     "+pterm.LightCyan("│"),
		pterm.LightCyan("│")+" "+pterm.Green("Version: ")+pterm.White(Version)+"                    "+pterm.LightCyan("│"),
		pterm.LightCyan("│")+" "+pterm.Blue("Built with: Go + Cobra + Pterm")+"     "+pterm.LightCyan("│"),
		pterm.LightCyan("└─────────────────────────────────────┘"),
	)
	
	fmt.Println()
	pterm.Info.Println("🚀 Modern CLI transformation complete!")
	pterm.Info.Println("📊 Get started with: todoinfo stats --client-id YOUR_ID")
}