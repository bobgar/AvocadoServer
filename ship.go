package main

import (
	pb "./proto"
	"github.com/golang/protobuf/proto"
)

//Struct that defines ships
type ShipDef struct {
	radius float32

	acceleration float32
	maxVel       float32
	velDampen    float32

	rotAcceleration float32
	maxRotVel       float32
	rotDampen       float32

	ability1 Ability
	ability2 Ability
}

type Ship struct {
	ShipDef
	*pb.Ship
}

func (ship Ship) handleInput(data []byte) {
	shipUpdate := new(pb.ShipUpdate)
	err := proto.Unmarshal(data, shipUpdate)
	if err == nil {
		if shipUpdate.Fire {
			ship.ability1.use()

			/*now := time.Now()
			timestamp := now.UnixNano()
			if (timestamp-ship.lastFire)/1000000 > rateOfFire {
				ship.lastFire = timestamp
			} else {
				shipUpdate.Fire = false
			}*/
		}

		shipUpdates[ship.Id] = shipUpdate
	}
}
