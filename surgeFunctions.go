package main

import (
	"github.com/rule110-io/surge-ui/surge"
	"github.com/rule110-io/surge-ui/surge/platform"
	"github.com/wailsapp/wails/v2"
)

type SurgeFunctions struct {
	r *wails.Runtime
}

func (s *SurgeFunctions) GetLocalFiles(Query string, OrderBy string, IsDesc bool, Skip int, Take int) surge.LocalFilePageResult {
	return surge.SearchLocalFile(Query, OrderBy, IsDesc, Skip, Take)
}

func (s *SurgeFunctions) GetRemoteFiles(Query string, OrderBy string, IsDesc bool, Skip int, Take int) surge.SearchQueryResult {
	return surge.SearchRemoteFile(Query, OrderBy, IsDesc, Skip, Take)
}

func (s *SurgeFunctions) GetPublicKey() string {
	return surge.GetMyAddress()
}

func (s *SurgeFunctions) GetFileChunkMap(Hash string, Size int) string {
	if Size == 0 {
		Size = 400
	}
	return surge.GetFileChunkMapStringByHash(Hash, Size)
}

func (s *SurgeFunctions) DownloadFile(Hash string) bool {
	return surge.DownloadFile(Hash)
}

func (s *SurgeFunctions) SetDownloadPause(Hash string, State bool) {
	surge.SetFilePause(Hash, State)
}

func (s *SurgeFunctions) OpenFile(Hash string) {
	surge.OpenFileByHash(Hash)
}

func (s *SurgeFunctions) OpenLink(Link string) {
	surge.OpenOSPath(Link)
}

func (s *SurgeFunctions) OpenLog() {
	surge.OpenLogFile()
}

func (s *SurgeFunctions) OpenFolder(Hash string) {
	surge.OpenFolderByHash(Hash)
}

func (s *SurgeFunctions) seedFile() bool {
	path := platform.OpenFileDialog()
	if path == "" {
		return false
	}
	return surge.SeedFile(path)
}

func (s *SurgeFunctions) RemoveFile(Hash string, FromDisk bool) bool {
	return surge.RemoveFile(Hash, FromDisk)
}

func (s *SurgeFunctions) WriteSetting(Key string, Value string) bool {
	err := surge.DbWriteSetting(Key, Value)
	return err != nil
}

func (s *SurgeFunctions) ReadSetting(Key string) string {
	val, _ := surge.DbReadSetting(Key)
	return val
}

func (s *SurgeFunctions) StartDownloadMagnetLinks(Magnetlinks string) bool {
	//need to parse Magnetlinks array and download all of them
	go surge.ParsePayloadString(Magnetlinks)
	return true
}
