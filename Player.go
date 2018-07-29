package main

import (
	"log"

	pb "./proto"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/websocket"
)

//Enum for the state players can be in
type PlayerState int

const (
	//State when player initially connects
	INITIAL_CONNECT PlayerState = iota
	//State when a player is logged in but not in a game
	LOGGED_IN
	//State while player is joined to a game
	IN_GAME
)

type Player struct {
	id        int32
	name      string
	state     PlayerState
	gameId    int32
	ship      *Ship
	websocket *websocket.Conn
}

func (player Player) listen() {
	for {
		data := make([]byte, 1000)
		_, err := player.websocket.Read(data)

		//If we ever get an error reading from the socket, assume the socket was closed and delete it.
		if err != nil {
			log.Printf("%v", err)
			player.cleanup()
			return
		}

		readSize := data[4]
		payload := data[8 : readSize+8]

		//If we're in "INITIAL_CONNECTION" the only type of packet we can accept is login.
		//Any other packet type will result in connection termination
		if player.state == INITIAL_CONNECT {
			login := new(pb.Login)
			err := proto.Unmarshal(data, login)
			//If we have an error on login abort the connection
			if err != nil {
				log.Printf("%v", err)
				player.cleanup()
			} else {
				//TODO in the future this will check a database and do an actual user login.
				//For now this just reads the username field.
				player.name = login.UserName
				player.state = LOGGED_IN
			}

		} else {
			message, err := unwrapMessage(&payload)

			if err != nil {
				log.Printf("%v", err)
			} else {
				switch message.MessageType {
				case pb.GenericMessage_JOIN_GAME:
					player.joinGame(message.Data)
				case pb.GenericMessage_SET_TEAM_AND_SHIP:
					if player.state == IN_GAME {
						player.setShipAndTeam(message.Data)
					}
				case pb.GenericMessage_SHIP_UPDATE:
					if player.state == IN_GAME && player.ship != nil {
						player.handleInput(message.Data)
					}
				}
			}
		}
	}
}

func (player Player) handleInput(data []byte) {
	shipUpdate := new(pb.ShipUpdate)
	err := proto.Unmarshal(data, shipUpdate)
	if err == nil {
		if shipUpdate.Ability1 {
			used := player.ship.ability1.use(player)
			if used {
				player.ship.needsUpdate = true
			}
		}

		player.ship.shipUpdate = shipUpdate
	}
}

func (player Player) joinGame(data []byte) {
	joinGame := new(pb.JoinGame)
	err := proto.Unmarshal(data, joinGame)
	if err == nil {
		if player.state == IN_GAME {
			games[player.gameId].removePlayer(&player)
		}
		if val, ok := games[joinGame.GameId]; ok {
			//If the game exists join it
			val.addPlayer(&player)
		} else {
			//If it doesn't exist create the new game
			games[joinGame.GameId] = &Game{
				players: make(map[int32]*Player),
				id:      joinGame.GameId,
				GameDef: &defaultGameDef}
			games[joinGame.GameId].addPlayer(&player)
			//then start the update loop thread
			go games[joinGame.GameId].update()
		}

		player.state = IN_GAME
	}
}

func (player Player) setShipAndTeam(data []byte) {

}

func unwrapMessage(data *[]byte) (pb.GenericMessage, error) {
	message := new(pb.GenericMessage)
	err := proto.Unmarshal(*data, message)
	if err != nil {
		return *message, err
	} else {
		return *message, nil
	}
}

func (player Player) cleanup() {
	delete(clients, player.id)
	delete(players, player.id)
	if player.state == IN_GAME {
		games[player.gameId].removePlayer(&player)
	}
}
