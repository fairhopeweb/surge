package main

import (
	"os"

	"log"

	"github.com/leaanthony/mewn"
	"github.com/rule110-io/surge-ui/surge"
	"github.com/rule110-io/surge-ui/surge/platform"
	"github.com/wailsapp/wails"
)

var wailsRuntime *wails.Runtime
var arguments []string

func getLocalFiles(Query string, OrderBy string, IsDesc bool, Skip int, Take int) surge.LocalFilePageResult {
	return surge.SearchLocalFile(Query, OrderBy, IsDesc, Skip, Take)
}

func getRemoteFiles(Query string, OrderBy string, IsDesc bool, Skip int, Take int) surge.SearchQueryResult {
	return surge.SearchRemoteFile(Query, OrderBy, IsDesc, Skip, Take)
}

func getPublicKey() string {
	return surge.GetMyAddress()
}

func getFileChunkMap(Hash string, Size int) string {
	if Size == 0 {
		Size = 400
	}
	return surge.GetFileChunkMapStringByHash(Hash, Size)
}

func downloadFile(Hash string) bool {
	return surge.DownloadFile(Hash)
}

func setDownloadPause(Hash string, State bool) {
	surge.SetFilePause(Hash, State)
}

func openFile(Hash string) {
	surge.OpenFileByHash(Hash)
}

func openLink(Link string) {
	surge.OpenOSPath(Link)
}

func openLog() {
	surge.OpenLogFile()
}

func openFolder(Hash string) {
	surge.OpenFolderByHash(Hash)
}

//RemoteClientOnlineModel holds info of remote clients
type RemoteClientOnlineModel struct {
	NumKnown  int
	NumOnline int
}

func seedFile() bool {
	path := platform.OpenFileDialog()
	if path == "" {
		return false
	}
	return surge.SeedFile(path)
}

func removeFile(Hash string, FromDisk bool) bool {
	return surge.RemoveFile(Hash, FromDisk)
}

func writeSetting(Key string, Value string) bool {
	err := surge.DbWriteSetting(Key, Value)
	return err != nil
}

func readSetting(Key string) string {
	val, _ := surge.DbReadSetting(Key)
	return val
}

func startDownloadMagnetLinks(Magnetlinks string) bool {
	//need to parse Magnetlinks array and download all of them
	go surge.ParsePayloadString(Magnetlinks)
	return true
}

// Stats .
type Stats struct {
	log *wails.CustomLogger
}

// WailsInit .
func (s *Stats) WailsInit(runtime *wails.Runtime) error {
	s.log = runtime.Log.New("Stats")
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

	js := mewn.String("./frontend/dist/app.js")
	css := mewn.String("./frontend/dist/app.css")

	app := wails.CreateApp(&wails.AppConfig{
		Width:     1144,
		Height:    790,
		Resizable: true,
		Title:     "Surge",
		JS:        js,
		CSS:       css,
		Colour:    "#131313",
	})
	app.Bind(stats)
	app.Bind(getLocalFiles)
	app.Bind(getRemoteFiles)
	app.Bind(downloadFile)
	app.Bind(setDownloadPause)
	app.Bind(openFile)
	app.Bind(openLink)
	app.Bind(openLog)
	app.Bind(openFolder)
	app.Bind(getFileChunkMap)
	app.Bind(seedFile)
	app.Bind(removeFile)
	app.Bind(writeSetting)
	app.Bind(readSetting)
	app.Bind(startDownloadMagnetLinks)
	app.Bind(getPublicKey)

	app.Run()

}
