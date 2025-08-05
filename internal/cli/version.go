package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var Version = "2.0.0" // CLI version

var (
	magentaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Bold(true)
	grayStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	greenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	whiteStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	blueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	infoStyleVer = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
)

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
	// Create a beautiful version display using lipgloss
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8BE9FD")).
		Padding(1, 2).
		Margin(1, 0)

	content := magentaStyle.Render("ToDo Info CLI") + "\n" +
		grayStyle.Render("Beautiful Microsoft ToDo Analysis") + "\n\n" +
		greenStyle.Render("Version: ") + whiteStyle.Render(Version) + "\n" +
		blueStyle.Render("Built with: Go + Cobra + Lipgloss")

	fmt.Println(lipgloss.Place(80, 1, lipgloss.Center, lipgloss.Center, box.Render(content)))

	fmt.Println()
	fmt.Println(infoStyleVer.Render("🚀 Modern CLI transformation complete!"))
	fmt.Println(infoStyleVer.Render("📊 Get started with: todoinfo stats --client-id YOUR_ID"))
}
