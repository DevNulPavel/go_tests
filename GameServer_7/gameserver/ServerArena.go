package gameserver

import (
	//"errors"
	"net"
	"sync/atomic"
)

var LAST_ID uint32 = 0

type ServerArena struct {
	arenaId              uint32
	server               *Server
	clients              []*ServerClient
	//arenaData            ArenaModel
	arenaData            []byte
	arenaState           GameArenaState
	isFull               uint32
	addClientByConnCh    chan *net.TCPConn
	deleteClientCh       chan *ServerClient
	clientStateUpdatedCh chan bool
	exitLoopCh           chan bool
}

func NewServerArena(server *Server) (*ServerArena, error) {
	newArenaId := atomic.AddUint32(&LAST_ID, 1)

	// State
	state := NewServerArenaState()
	state.ID = newArenaId

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
		arenaId:              newArenaId,
		server:               server,
		clients:              make([]*ServerClient, 0),
		arenaData:            GetApp().GetStaticInfo().TestArenaData,
		arenaState:           state,
		isFull:               0,
		addClientByConnCh:    make(chan *net.TCPConn),
		deleteClientCh:       make(chan *ServerClient),
		clientStateUpdatedCh: make(chan bool),
		exitLoopCh:           make(chan bool),
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
	arena.clientStateUpdatedCh <- true
}

func (arena *ServerArena) GetIsFull() bool {
	return bool(atomic.LoadUint32(&arena.isFull) > 0)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (arena *ServerArena) sendAllNewState() {
	// TODO: Send all
}

func (arena *ServerArena) mainLoop() {
	for {
		select {
		// Канал добавления нового юзера
		case connection := <-arena.addClientByConnCh:
			client := NewClient(connection, arena)
			arena.clients = append(arena.clients, client)
			client.StartLoop()


            client.QueueSendData(arena.arenaData)

			/*arenaMapData, err := arena.arenaData.ToBytes()
			if err == nil {
				client.QueueSendData(arenaMapData)
			}*/
			//client.QueueSendCurrentClientState()
			// TODO: Send arena

		// Канал обновления состояния юзера
		case <-arena.clientStateUpdatedCh:
			arena.sendAllNewState()

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
