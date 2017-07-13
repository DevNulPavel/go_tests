package gameserver

import (
	//"errors"
	"log"
	"net"
	"sync/atomic"
	"time"
)

var LAST_ID uint32 = 0

type ServerArena struct {
	arenaId uint32
	server  *Server
	clients []*ServerClient
	//arenaData            ArenaModel
	arenaData         []byte
	arenaState        GameArenaState
	isFull            uint32
	needSendAll       uint32
	addClientByConnCh chan *net.TCPConn
	deleteClientCh    chan *ServerClient
	exitLoopCh        chan bool
}

func NewServerArena(server *Server) (*ServerArena, error) {
	newArenaId := atomic.AddUint32(&LAST_ID, 1)

	// State
	state := NewServerArenaState(newArenaId)

	// Формируем список платформ для данной арены
	/*item, exists := GetApp().GetStaticInfo().Levels["egypt"]
	if exists == false {
		return nil, errors.New("No level with name")
	}
	platformsForArena := make([]*PlatformInfo, 0)
	for _, key := range item.Platforms {
		value, ok := GetApp().GetStaticInfo().Platforms[key]
		if ok {
			platformsForArena = append(platformsForArena, value)
		}
	}
	if len(platformsForArena) == 0 {
		return nil, errors.New("No platforms for arena")
	}
	arenaData := NewArenaModel(platformsForArena)*/

	// Server arena
	arena := &ServerArena{
		arenaId:           newArenaId,
		server:            server,
		clients:           make([]*ServerClient, 0),
		arenaData:         GetApp().GetStaticInfo().TestArenaData,
		arenaState:        state,
		isFull:            0,
		needSendAll:       0,
		addClientByConnCh: make(chan *net.TCPConn),
		deleteClientCh:    make(chan *ServerClient),
		exitLoopCh:        make(chan bool),
	}
	return arena, nil
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (arena *ServerArena) StartLoop() {
	go arena.mainLoop()
}

func (arena *ServerArena) Exit() {
	arena.exitLoopCh <- true
}

func (arena *ServerArena) AddClientForConnection(connection *net.TCPConn) {
	arena.addClientByConnCh <- connection
}

func (arena *ServerArena) DeleteClient(client *ServerClient) {
	arena.deleteClientCh <- client
}

func (arena *ServerArena) ClientStateUpdated(client *ServerClient) {
	atomic.StoreUint32(&arena.needSendAll, 1)
}

func (arena *ServerArena) GetIsFull() bool {
	return bool(atomic.LoadUint32(&arena.isFull) > 0)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (arena *ServerArena) sendAllNewState() {
	// Sync states
	arena.arenaState.Clients = make([]ServerClientState, 0)
	for _, client := range arena.clients {
		if client.IsValidState() {
			stateCopy := client.GetCurrentState()
			arena.arenaState.Clients = append(arena.arenaState.Clients, stateCopy)
		}
	}

	// State to data
	data, err := arena.arenaState.ToBytes()
	if err != nil {
		log.Printf("Failed arena state marshaling: %s\n", err)
		return
	}

	// Send all
	for _, client := range arena.clients {
		client.QueueSendData(data)
	}
}

func (arena *ServerArena) worldTick(delta float64) {
}

func (arena *ServerArena) mainLoop() {
	const updatePeriodMS = time.Millisecond * 50
	timer := time.NewTimer(updatePeriodMS)
	lastTickTime := time.Now()

	for {
		select {
		// Канал добавления нового юзера
		case connection := <-arena.addClientByConnCh:
			client := NewClient(connection, arena)
			arena.clients = append(arena.clients, client)
			client.StartLoop()

			client.QueueSendData(arena.arenaData)
			client.QueueSendCurrentClientState()

			/*arenaMapData, err := arena.arenaData.ToBytes()
			if err == nil {
				client.QueueSendData(arenaMapData)
			}*/
			// TODO: Send arena

		// Основной серверный таймер, который обновляет серверный мир
		case <-timer.C:
			timer.Reset(updatePeriodMS)
			delta := time.Now().Sub(lastTickTime).Seconds()
			lastTickTime = time.Now()

			arena.worldTick(delta)

			if atomic.LoadUint32(&arena.needSendAll) > 0 {
				atomic.StoreUint32(&arena.needSendAll, 0)
				arena.sendAllNewState()
			}

		// Канал удаления нового юзера
		case client := <-arena.deleteClientCh:
			deleteIndex := -1
			for i := range arena.clients {
				if arena.clients[i].id == client.id {
					deleteIndex = i
					break
				}
			}
			if deleteIndex >= 0 {
				arena.clients = append(arena.clients[:deleteIndex], arena.clients[deleteIndex+1:]...)
				arena.sendAllNewState()
			}

		// Выход из цикла обработки событий
		case <-arena.exitLoopCh:
			// Clients
			for _, client := range arena.clients {
				client.Close()
			}
			// Server
			arena.server.DeleteRoom(arena)
			return
		}
	}
}
