package main

import (
	"os"

	"log"

	"github.com/rule110-io/surge-ui/surge"
	"github.com/rule110-io/surge-ui/surge/platform"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

var wailsRuntime *wails.Runtime
var arguments []string

//RemoteClientOnlineModel holds info of remote clients
type RemoteClientOnlineModel struct {
	NumKnown  int
	NumOnline int
}

// Stats .
type Stats struct {
}

// WailsInit .
func (s *Stats) WailsInit(runtime *wails.Runtime) error {
	go surge.WailsBind(runtime)

	return nil
}

//WailsRuntime .
type WailsRuntime struct {
	runtime *wails.Runtime
}

//WailsShutdown .
func (s *WailsRuntime) WailsShutdown() {
	surge.Stop()
}

func main() {
	defer surge.RecoverAndLog()

	keepRunning := platform.ProcessStartupArgs(os.Args, &surge.FrontendReady)
	if !keepRunning {
		return
	}

	//surge.HashFile("C:\\Users\\mitch\\Downloads\\surge_remote\\surge-0.2.0-beta.windows.zip")

	stats := &Stats{}

	argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]
	log.Println(argsWithProg)
	log.Println(argsWithoutProg)

	//invoked with a download
	if len(argsWithoutProg) > 0 {
		arguments = os.Args[1:]
	}

	//Initialize folder structures on os filesystem
	newlyCreated, err := platform.InitializeFolders()
	if err != nil {
		log.Panic("Error on startup", err.Error())
	}
	surge.InitializeDb()
	//surge.InitializeLog()
	defer surge.CloseDb()
	if newlyCreated {
		// seems like this is the first time starting the app
		//set tour to active
		surge.DbWriteSetting("Tour", "true")
		//set default mode to light
		surge.DbWriteSetting("DarkMode", "false")
	}

	surge.Start(arguments)

	// Create application with options
	app, err := wails.CreateAppWithOptions(&options.App{
		Title:     "Surge",
		Width:     1144,
		Height:    790,
		MinWidth:  1144,
		MinHeight: 790,
		//Tray:      menu.NewMenuFromItems(menu.AppMenu()),
		//Menu:      menu.NewMenuFromItems(menu.AppMenu()),
		//StartHidden:  true,
		Mac: &mac.Options{
			WebviewIsTransparent:          true,
			WindowBackgroundIsTranslucent: true,
			// Comment out line below to see Window.SetTitle() work
			TitleBar: mac.TitleBarHiddenInset(),
			//Menu:     createApplicationMenu(),
			//Tray: &menu.TrayOptions{
			//	Icon: "light",
			//	Menu: createApplicationTray(),
			//},
		},

		LogLevel: logger.TRACE,
	})

	app.Bind(stats)

	app.Bind(&SurgeFunctions{})

	app.Run()

}
