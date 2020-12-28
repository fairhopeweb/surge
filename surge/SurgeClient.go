package surge

import (
	"fmt"
	"time"

	"log"

	nkn "github.com/nknorg/nkn-sdk-go"
	"github.com/rule110-io/surge-ui/surge/constants"
	"github.com/rule110-io/surge-ui/surge/platform"
	"github.com/rule110-io/surge-ui/surge/sessionmanager"
)

//ChunkSize is size of chunk in bytes (256 kB)
const ChunkSize = 1024 * 256

//NumClients is the number of NKN clients
const NumClients = 8

//NumWorkers is the total number of concurrent chunk fetches allowed
const NumWorkers = 32

//whether the nkn client is initialized
var clientInitialized = false

//The nkn client
var client *nkn.MultiClient

//duration of a subscription blocktime is ~20sec
const subscriptionDuration = 180 // 180 is approximately one hour

//InitializeClient initializes connection with nkn
func InitializeClient(args []string) bool {
	var err error

	account := InitializeAccount()
	client, err = nkn.NewMultiClient(account, "", NumClients, false, &nkn.ClientConfig{
		ConnectRetries: 1000,
	})
	if err != nil {
		pushError(err.Error(), "do you have an active internet connection?")
	}

	<-client.OnConnect.C
	clientInitialized = true
	sessionmanager.Initialize(client, onClientConnected, onClientDisconnected)

	pushNotification("Client Connected", "Successfully connected to the NKN network")

	client.Listen(nil)
	go Listen()

	dbFiles := dbGetAllFiles()
	var filesOnDisk []File

	for i := 0; i < len(dbFiles); i++ {
		if FileExists(dbFiles[i].Path) {
			filesOnDisk = append(filesOnDisk, dbFiles[i])
		} else {
			dbFiles[i].IsMissing = true
			dbFiles[i].IsDownloading = false
			dbFiles[i].IsUploading = false
			dbInsertFile(dbFiles[i])
		}
	}

	go BuildSeedString(filesOnDisk)
	for i := 0; i < len(filesOnDisk); i++ {
		if filesOnDisk[i].IsDownloading {
			go restartDownload(filesOnDisk[i].FileHash)
		}
	}

	go autoSubscribeWorker()

	go platform.WatchOSXHandler()

	//Insert new file from arguments and start download
	if args != nil && len(args) > 0 && len(args[0]) > 0 {
		platform.AskUser("startDownloadMagnetLinks", "{files : ["+args[0]+"]}")
	}

	return true
}

//Stop cleanup for surge
func Stop() {
	client.Close()
	client = nil
}

//Function that automatically grabs subscriptions for nkn topic
func rescanPeers() {

	for true {
		time.Sleep(constants.RescanPeerInterval)
		GetSubscriptions()
	}
}

func autoSubscribeWorker() {

	//As long as the client is running subscribe
	for true {
		//Only subscribe when this client is hosting anything
		hosting := false

		files := dbGetAllFiles()
		for i := 0; i < len(files); i++ {
			if files[i].IsUploading {
				hosting = true
				break
			}
		}

		if hosting {
			subscribeToSurgeTopic()
		}

		time.Sleep(time.Second * 20 * subscriptionDuration)
	}
}

func subscribeToSurgeTopic() {
	Topic := TopicEncode(TestTopic)
	txnHash, err := client.Subscribe("", Topic, subscriptionDuration, "Surge Beta Client", nil)
	if err != nil {
		log.Println("Probably already subscribed", err)
	} else {
		log.Println("Subscribed: ", txnHash)
	}
}

//GetSubscriptions .
func GetSubscriptions() {

	Topic := TopicEncode(TestTopic)

	subResponse, err := client.GetSubscribers(Topic, 0, 100, true, true)
	if err != nil {
		pushError(err.Error(), "do you have an active internet connection?")
		return
	}

	for k, v := range subResponse.SubscribersInTxPool.Map {
		subResponse.Subscribers.Map[k] = v
	}

	subscribers = []string{}
	for k, v := range subResponse.Subscribers.Map {
		if len(v) > 0 {
			if k != client.Addr().String() {
				subscribers = append(subscribers, k)
			}
		}
	}

	fmt.Println(string("\033[36m"), "Get Subscriptions", len(subscribers), string("\033[0m"))

	for _, sub := range subscribers {
		connectAndQueryJob := func(addr string) {
			_, err := sessionmanager.GetSession(addr, constants.GetSessionDialTimeout)
			if err == nil {
				fmt.Println(string("\033[36m"), "Sending file query to subscriber", addr, string("\033[0m"))
				go SendQueryRequest(addr, "Testing query functionality.")
			}
		}
		go connectAndQueryJob(sub)
	}
}

func updateNumClientStore() {
	numConnections := sessionmanager.GetSessionLength()
	if clientInitialized {
		numConnections++
	}
	numClientsStore.Update(func(data NumClientsStruct) NumClientsStruct {
		return NumClientsStruct{
			Online: numConnections,
		}
	})
}

func listenForIncomingSessions() {

	for !client.IsClosed() {
		listenSession, err := client.Accept()
		if err != nil {
			pushError("Error on client accept", err.Error())
			continue
		}

		sessionmanager.AcceptSession(listenSession)
	}
}

// Listen will listen to incoming requests for chunks
func Listen() {
	go listenForIncomingSessions()
}

func onClientConnected(session *sessionmanager.Session, isDialIn bool) {
	go updateNumClientStore()
	addr := session.Session.RemoteAddr().String()

	fmt.Println(string("\033[36m"), "Client Connected", addr, string("\033[0m"))

	go listenToSession(session)

	if isDialIn {
		fmt.Println(string("\033[36m"), "Sending file query to accepted client", addr, string("\033[0m"))
		go SendQueryRequest(addr, "Testing query functionality.")
	}
}

func onClientDisconnected(addr string) {
	go updateNumClientStore()

	//Remove this address from remote file seeders
	ListedFilesLock.Lock()
	for i := 0; i < len(ListedFiles); i++ {
		ListedFiles[i].seeders = removeStringFromSlice(ListedFiles[i].seeders, addr)
		ListedFiles[i].seederCount = len(ListedFiles[i].seeders)
		fmt.Println(string("\033[31m"), "onClientDisconnected", ListedFiles[i].FileName, "seeders remaining:", ListedFiles[i].seederCount, string("\033[0m"))
	}

	//Remove empty seeders listings
	for i := 0; i < len(ListedFiles); i++ {
		if len(ListedFiles[i].seeders) == 0 {
			// Remove the element at index i from a.
			ListedFiles[i] = ListedFiles[len(ListedFiles)-1] // Copy last element to index i.
			ListedFiles[len(ListedFiles)-1] = File{}         // Erase last element (write zero value).
			ListedFiles = ListedFiles[:len(ListedFiles)-1]   // Truncate slice.
			i--
		}
	}

	ListedFilesLock.Unlock()
}

func listenToSession(Session *sessionmanager.Session) {
	defer RecoverAndLog()

	addr := Session.Session.RemoteAddr().String()

	fmt.Println(string("\033[31m"), "Initiate Session", addr, string("\033[0m"))

	for Session.Session != nil {
		data, chunkType, err := SessionRead(Session)
		fmt.Println(string("\033[31m"), "Read data from session", addr, string("\033[0m"))

		if err != nil {
			log.Println("Session read failed, closing session error:", err)
			break
		}

		sessionmanager.UpdateActivity(Session.Session.RemoteAddr().String())

		switch chunkType {
		case surgeChunkID:
			//Write add to download internally after parsing data
			processChunk(Session, data)
			break
		case surgeQueryRequestID:
			processQueryRequest(Session, data)
			//Write add to download
			bandwidthAccumulatorMapLock.Lock()
			downloadBandwidthAccumulator["DISCOVERY"] += len(data)
			bandwidthAccumulatorMapLock.Unlock()
			break
		case surgeQueryResponseID:
			processQueryResponse(Session, data)
			//Write add to download
			bandwidthAccumulatorMapLock.Lock()
			downloadBandwidthAccumulator["DISCOVERY"] += len(data)
			bandwidthAccumulatorMapLock.Unlock()
			break
		}
	}
}
