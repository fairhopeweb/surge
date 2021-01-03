package platform

import (
	"github.com/wailsapp/wails/v2/pkg/options"
)

//OpenFileDialog uses platform agnostic package for a file dialog
func OpenFileDialog() string {
	selectedFile := wailsRuntime.Dialog.Open(&options.OpenDialog{
		Title:                      "Choose a file to add to surge",
		AllowFiles:                 true,
		AllowDirectories:           false,
		AllowMultiple:              false,
		ShowHiddenFiles:            false,
		CanCreateDirectories:       false,
		ResolvesAliases:            false,
		TreatPackagesAsDirectories: false,
	})
	// selectedFiles is a string slice. Get the first selection
	if len(selectedFile) == 0 {
		// Cancelled
		return ""
	}

	return string(selectedFile[0])
}
