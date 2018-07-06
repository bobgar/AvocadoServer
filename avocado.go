package main

import (
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	pb "./proto"
	"github.com/golang/protobuf/proto"

	"golang.org/x/net/websocket"
)

var clients = make(map[int32]*websocket.Conn)
var ships = make(map[int32]*pb.Ship)
var shipUpdates = make(map[int32]*pb.ShipUpdate)
var id int32 = 0
var bulletId int32 = 0

var state = &pb.GameState{Ships: []*pb.Ship{}}

var rotAcceleration float32 = .01
var thrust float32 = .5

var worldWidth float32 = 1200
var worldHeight float32 = 900

var maxVel float32 = 20
var maxRotVel float32 = .3
var bulletTTL int64 = 1500
var rateOfFire int64 = 250

//TODO this is actually really ship + bullet radius for now to simplify calculations
var bulletRadius float32 = 30

var rotDampen float32 = .98
var velDampen float32 = .98

// Echo the data received on the WebSocket.
func ReadClient(ws *websocket.Conn) {
	clients[id] = ws

	//A new user has connected, create a new ship and add it to our list of ships!
	ship := new(pb.Ship)
	ship.Id = id
	id++
	ship.Name = "ship " + strconv.Itoa(int(id))
	ship.Rot = 0
	ship.XPos = rand.Float32() * worldWidth
	ship.YPos = rand.Float32() * worldHeight
	state.Ships = append(state.Ships, ship)

	ships[ship.Id] = ship

	var lastFire int64 = 0

	for {
		data := make([]byte, 100)
		count, err := ws.Read(data)

		//log.Printf("%v %v %v", count, err, &data)
		if err != nil {
			delete(clients, ship.Id)
			delete(ships, ship.Id)
			delete(shipUpdates, ship.Id)
			for i, curShip := range state.Ships {
				if ship == curShip {
					state.Ships = append(state.Ships[:i], state.Ships[i+1:]...)
				}
			}
			//TODO remove ship
			log.Printf("%v", err)
			return
		} else if count > 0 {
			//TODO for some reason this seems to be the relevant bytes?
			readSize := data[4]
			//log.Printf("size : %v", readSize)
			relevantData := data[8 : readSize+8]
			//log.Printf("%v", relevantData)
			message, err := unwrapMessage(&relevantData)
			if err != nil {
				log.Printf("%v", err)
			} else {
				switch message.MessageType {
				case pb.GenericMessage_SHIP_UPDATE:
					//log.Printf("Got ship update")
					shipUpdate := new(pb.ShipUpdate)
					err := proto.Unmarshal(message.Data, shipUpdate)
					if err == nil {
						if shipUpdate.Fire {
							now := time.Now()
							timestamp := now.UnixNano()
							if (timestamp-lastFire)/1000000 > rateOfFire {
								lastFire = timestamp
							} else {
								shipUpdate.Fire = false
							}
						}

						shipUpdates[ship.Id] = shipUpdate
					}
				}
			}
		}
	}
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

// This example demonstrates a trivial echo server.
func main() {

	go WorldUpdates()

	http.Handle("/ws", websocket.Handler(ReadClient))
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func WorldUpdates() {
	for {
		time.Sleep(33000000) //In MS I assume?

		now := time.Now()
		timestamp := now.UnixNano()

		//Apply user controls
		for id, shipUpdate := range shipUpdates {
			if shipUpdate != nil {
				ship := ships[id]
				//log.Printf("%v", ship)
				if ship != nil {
					if shipUpdate.RotLeft {
						ship.RotVel -= rotAcceleration
						if ship.RotVel < -maxRotVel {
							ship.RotVel = -maxRotVel
						}
					}
					if shipUpdate.RotRight {
						ship.RotVel += rotAcceleration
						if ship.RotVel > maxRotVel {
							ship.RotVel = maxRotVel
						}
					}
					if shipUpdate.Thrust {
						ship.XVel += float32(math.Sin(float64(ship.Rot))) * thrust
						ship.YVel += -float32(math.Cos(float64(ship.Rot))) * thrust
						magnitude := float32(math.Sqrt(float64(ship.XVel*ship.XVel + ship.YVel*ship.YVel)))
						if magnitude > maxVel {
							scale := maxVel / magnitude
							ship.XVel *= scale
							ship.YVel *= scale
						}
					}
					if shipUpdate.Fire {
						bullet := new(pb.Ship_Bullet)
						bullet.Id = bulletId
						bulletId++
						bullet.XPos = ship.XPos
						bullet.XVel = ship.XVel + float32(math.Sin(float64(ship.Rot)))*maxVel
						bullet.YPos = ship.YPos

						bullet.YVel = ship.YVel - float32(math.Cos(float64(ship.Rot)))*maxVel
						bullet.Timestamp = timestamp
						ship.Bullets = append(ship.Bullets, bullet)
					}
				}

				//Once we've applied the controls we set to nil so we can reset them next update
				shipUpdates[id] = nil
			}
		}

		//Apply velocities and check collisions
		for _, ship := range ships {
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

		}

		//Send updated game state
		stateBytes, err := proto.Marshal(state)
		message := new(pb.GenericMessage)
		message.MessageType = pb.GenericMessage_GAME_STATE_UPDATE
		message.Data = stateBytes
		out, err := proto.Marshal(message)

		if err == nil {
			err := SendToClients(&out)
			if err != nil {
				log.Printf("%v", err)
			}
		}
	}
}

func respawnShip(enemyShip *pb.Ship) {
	enemyShip.XVel = 0
	enemyShip.YVel = 0
	enemyShip.Rot = 0
	enemyShip.XPos = rand.Float32() * worldWidth
	enemyShip.YPos = rand.Float32() * worldHeight
}

func SendToClients(data *[]byte) error {
	for _, client := range clients {
		//client.Write(data)
		frame, err := client.NewFrameWriter(websocket.BinaryFrame)
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

func CreateMessage(data []byte, messageType pb.GenericMessage_MessageTypeEnum) ([]byte, error) {
	message := new(pb.GenericMessage)
	message.MessageType = messageType
	message.Data = data
	out, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return out, nil
}
