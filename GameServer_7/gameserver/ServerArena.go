package gameserver

import (
	"errors"
	"log"
	//"math"
	"net"
	"sync/atomic"
	"time"
    "math/rand"
)

var LAST_ID uint32 = 0
var LAST_MONSTER_ID uint32 = 0

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
	forceSendAll      chan bool
	exitLoopCh        chan bool
}

func NewServerArena(server *Server) (*ServerArena, error) {
	newArenaId := atomic.AddUint32(&LAST_ID, 1)

	// State
	state := NewServerArenaState(newArenaId)

	// Формируем список платформ для данной арены
	item, exists := GetApp().GetStaticInfo().Levels["egypt"]
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
    arenaModel := NewArenaModel(platformsForArena)
	arenaData, err := arenaModel.ToBytes()
    if err != nil {
        return nil, err
    }

	//arenaData := GetApp().GetStaticInfo().TestArenaData

	// Server arena
	arena := &ServerArena{
		arenaId:           newArenaId,
		server:            server,
		clients:           make([]*ServerClient, 0),
		arenaData:         arenaData,
		arenaState:        state,
		isFull:            0,
		needSendAll:       0,
		addClientByConnCh: make(chan *net.TCPConn),
		deleteClientCh:    make(chan *ServerClient),
		forceSendAll:      make(chan bool),
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

func (arena *ServerArena) ClientStateUpdated(client *ServerClient, force bool) {
	if force {
		arena.forceSendAll <- true
	} else {
		atomic.StoreUint32(&arena.needSendAll, 1)
	}
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
			stateCopy := client.GetCurrentState(true)
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
	if len(arena.arenaState.Monsters) > 0 {
		hits := []ClientCommandHitInfo{}
		for _, client := range arena.clients {
			hits = append(hits, client.GetCurrentHitsWithReset()...)
		}

		// TODO: Optimize
		validMonsters := make([]ServerMonsterState, 0)
		haveUpdates := false
		for i, _ := range arena.arenaState.Monsters {
			for _, hit := range hits {
				if arena.arenaState.Monsters[i].ID == hit.ID {
					arena.arenaState.Monsters[i].Health -= hit.Damage / 10

					log.Printf("Hit monster %d: damage = %d, health = %d\n", hit.ID, hit.Damage, arena.arenaState.Monsters[i].Health)

					haveUpdates = true
				}
			}

			if arena.arenaState.Monsters[i].Health > 0 {
				//arena.arenaState.Monsters[i].Health = int16(math.Max(float64(arena.arenaState.Monsters[i].Health), 0.0))
				validMonsters = append(validMonsters, arena.arenaState.Monsters[i])
			}
		}
		arena.arenaState.Monsters = validMonsters

		if haveUpdates == true {
			atomic.StoreUint32(&arena.needSendAll, 1)
		}
	}
}

func (arena *ServerArena) createMonster() {
	if len(arena.arenaState.Monsters) == 0 {
		newMonsterId := atomic.AddUint32(&LAST_MONSTER_ID, 1)

        points := [5]Point16{
            NewPoint16(10, 2),
            NewPoint16(2, 2),
            NewPoint16(2, 10),
            NewPoint16(10, 10),
            NewPoint16(4, 5),
        }

        point := points[rand.Int() % len(points)]

		monsterState := NewServerMonsterState(newMonsterId)
		monsterState.Name = "angry_cat"
		monsterState.Health = 1000
		monsterState.X = float64(point.X)
		monsterState.Y = float64(point.Y)

		arena.arenaState.Monsters = append(arena.arenaState.Monsters, monsterState)

		log.Printf("Generated monster %d", newMonsterId)

		atomic.StoreUint32(&arena.needSendAll, 1)
	}
}

func (arena *ServerArena) mainLoop() {
	updatePeriodMS := time.Millisecond * 50
	updateTimer := time.NewTimer(updatePeriodMS)
	lastTickTime := time.Now()

	monsterGeneratePeriod := time.Second * 3
	newMonsterTimer := time.NewTimer(monsterGeneratePeriod)

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
		case <-updateTimer.C:
			updateTimer.Reset(updatePeriodMS)
			delta := time.Now().Sub(lastTickTime).Seconds()
			lastTickTime = time.Now()

			arena.worldTick(delta)

			if atomic.LoadUint32(&arena.needSendAll) > 0 {
				atomic.StoreUint32(&arena.needSendAll, 0)
				arena.sendAllNewState()
			}

		case <-newMonsterTimer.C:
			newMonsterTimer.Reset(time.Second * 20)
			arena.createMonster()

		case <-arena.forceSendAll:
			atomic.StoreUint32(&arena.needSendAll, 0)
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
			updateTimer.Stop()
			newMonsterTimer.Stop()
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
