package platform

import (
	"io/ioutil"

	"github.com/wailsapp/wails/v2/pkg/options"
)

//OpenFileDialog uses platform agnostic package for a file dialog
func OpenFileDialog() string {
	selectedFile := wailsRuntime.Dialog.Open(&options.OpenDialog{
		DefaultDirectory:           "",
		DefaultFilename:            "",
		Title:                      "",
		Filters:                    "",
		AllowFiles:                 false,
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

	// Load notes
	noteData, err := ioutil.ReadFile(selectedFile[0])
	if err != nil {
		return ""
	}

	return string(noteData)
}
