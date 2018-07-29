package main

import (
	"log"
	"time"

	pb "./proto"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/websocket"
)

/*var worldWidth float32 = 1024
var worldHeight float32 = 900*/

//TODO REMOVE these are handled by ship / ship type now
/*var rotAcceleration float32 = .01
var thrust float32 = .5
var maxVel float32 = 20
var maxRotVel float32 = .3
var bulletTTL int64 = 1500
var rateOfFire int64 = 250*/

var defaultGameDef = GameDef{
	gameType:    TEAM_DEATHMATCH,
	maxTeams:    8,
	maxTeamSize: 16,
	respawnTime: 5000,
	width:       1024,
	height:      1024}

type Game struct {
	id      int32
	players map[int32]*Player
	bullets map[int32]*Bullet
	*GameDef
	bulletId int32
	/*clients     map[int32]*websocket.Conn
	ships       map[int32]*pb.Ship
	shipUpdates map[int32]*pb.ShipUpdate
	state       pb.GameState*/
}

//Enum for various game modes
type GameType int

const (
	//Default (original) game mode
	DEATHMATCH GameType = iota
	//basically just deathmatch
	TEAM_DEATHMATCH
	//TODO add more!
)

type GameDef struct {
	gameType    GameType
	maxTeams    int32
	maxTeamSize int32
	respawnTime int64
	//TODO should these be ints?
	width  float32
	height float32
}

func (game Game) update() {

	update := new(pb.GameState)
	for {
		//TODO I should take into account time to do update and subtract that from sleep time.

		state := new(pb.GameState)

		timestamp := time.Now().UnixNano()

		for _, bullet := range game.bullets {
			bullet.update(&game, timestamp, update)
		}

		for _, player := range game.players {
			if player.ship != nil {
				player.ship.update(&game, update)
			}
		}

		stateBytes, err := proto.Marshal(state)
		message := new(pb.GenericMessage)
		message.MessageType = pb.GenericMessage_GAME_STATE_UPDATE
		message.Data = stateBytes
		out, err := proto.Marshal(message)

		if err == nil {
			err := game.SendToClients(&out)
			if err != nil {
				log.Printf("%v", err)
			}
		}

		timestampEnd := time.Now().UnixNano()

		time.Sleep(time.Duration(33000000 - (timestampEnd - timestamp))) //In nano seconds (33ms)

		update = new(pb.GameState)

		/*if shipUpdate.Fire {
			bullet := new(pb.Ship_Bullet)
			bullet.Id = bulletId
			bulletId++
			bullet.XPos = ship.XPos
			bullet.XVel = ship.XVel + float32(math.Sin(float64(ship.Rot)))*maxVel
			bullet.YPos = ship.YPos

			bullet.YVel = ship.YVel - float32(math.Cos(float64(ship.Rot)))*maxVel
			bullet.Timestamp = timestamp
			ship.Bullets = append(ship.Bullets, bullet)
		}*/

		//Apply velocities and check collisions
		/*for _, ship := range ships {
			ship.XPos += ship.XVel
			ship.YPos += ship.YVel
			if ship.XPos < 0 {
				ship.XPos += worldWidth
			} else if ship.XPos > worldWidth {
				ship.XPos -= worldWidth
			}
			if ship.YPos < 0 {
				ship.YPos += worldHeight
			} else if ship.YPos > worldHeight {
				ship.YPos -= worldHeight
			}

			log.Printf("x: %v  ,  y: %v", ship.XPos, ship.YPos)

			ship.XVel *= velDampen
			ship.YVel *= velDampen

			ship.Rot += ship.RotVel
			if ship.Rot > math.Pi*2 {
				ship.Rot -= math.Pi * 2
			} else if ship.Rot < 0 {
				ship.Rot += math.Pi * 2
			}
			ship.RotVel *= rotDampen

			//for i, bullet := range ship.Bullets {
			for i := len(ship.Bullets) - 1; i >= 0; i-- {
				bullet := ship.Bullets[i]
				//Delete timed out bullets
				if (timestamp-bullet.Timestamp)/1000000 > bulletTTL {
					log.Printf("removing bullet: %v ", ship.Bullets[i])
					ship.Bullets = append(ship.Bullets[:i], ship.Bullets[i+1:]...)
				} else {
					bullet.XPos += bullet.XVel
					bullet.YPos += bullet.YVel

					if bullet.XPos < 0 {
						bullet.XPos += worldWidth
					} else if bullet.XPos > worldWidth {
						bullet.XPos -= worldWidth
					}
					if bullet.YPos < 0 {
						bullet.YPos += worldHeight
					} else if bullet.YPos > worldHeight {
						bullet.YPos -= worldHeight
					}
					//TODO Test colision
					for _, enemyShip := range ships {
						if enemyShip == ship {
							continue
						}
						deltaX := bullet.XPos - enemyShip.XPos
						deltaY := bullet.YPos - enemyShip.YPos
						if bulletRadius*bulletRadius > (deltaX*deltaX + deltaY*deltaY) {
							//COLLISION!
							ship.Bullets = append(ship.Bullets[:i], ship.Bullets[i+1:]...)
							respawnShip(enemyShip)
						}
					}
				}
			}
		}*/
	}
}

func (game Game) SendToClients(data *[]byte) error {
	for _, player := range game.players {
		//client.Write(data)
		frame, err := player.websocket.NewFrameWriter(websocket.BinaryFrame)
		if err != nil {
			return err
		}
		_, err = frame.Write(*data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (game Game) addPlayer(player *Player) {
	game.players[player.id] = player
	//TODO a lot more logic needed here
}

func (game Game) removePlayer(player *Player) {
	delete(game.players, player.id)
}

func (game Game) cleanup() {
}
