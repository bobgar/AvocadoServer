package main

import (
	"math"
	"time"

	pb "./proto"
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

	maxHealth int32

	ability1 Ability
	ability2 Ability
}

type Ship struct {
	ShipDef
	shipState  *pb.Ship
	shipUpdate *pb.ShipUpdate
	//needsUpdate flags if we need to resend to the client due to state or other things changing
	needsUpdate     bool
	lastUpdate      int64
	team            int32
	lastStateChange int64
}

func (ship Ship) update(game *Game, update *pb.GameState) {

	if ship.shipUpdate != nil {
		if ship.shipUpdate.RotLeft {
			ship.shipState.RotVel -= ship.rotAcceleration
			if ship.shipState.RotVel < -ship.maxRotVel {
				ship.shipState.RotVel = -ship.maxRotVel
			}
		}
		if ship.shipUpdate.RotRight {
			ship.shipState.RotVel += ship.rotAcceleration
			if ship.shipState.RotVel > ship.maxRotVel {
				ship.shipState.RotVel = ship.maxRotVel
			}
		}
		if ship.shipUpdate.Thrust {
			//First we apply controls
			ship.shipState.XVel += float32(math.Sin(float64(ship.shipState.Rot))) * ship.acceleration
			ship.shipState.YVel += -float32(math.Cos(float64(ship.shipState.Rot))) * ship.acceleration
			magnitude := float32(math.Sqrt(float64(ship.shipState.XVel*ship.shipState.XVel + ship.shipState.YVel*ship.shipState.YVel)))
			if magnitude > ship.maxVel {
				scale := ship.maxVel / magnitude
				ship.shipState.XVel *= scale
				ship.shipState.YVel *= scale
			}
		}

		ship.shipUpdate = nil
	}

	//then we handle movement
	ship.shipState.XPos += ship.shipState.XVel
	ship.shipState.YPos += ship.shipState.YVel
	if ship.shipState.XPos < 0 {
		ship.shipState.XPos += game.width
	} else if ship.shipState.XPos > game.width {
		ship.shipState.XPos -= game.width
	}
	if ship.shipState.YPos < 0 {
		ship.shipState.YPos += game.height
	} else if ship.shipState.YPos > game.height {
		ship.shipState.YPos -= game.height
	}

	//log.Printf("x: %v  ,  y: %v", ship.XPos, ship.YPos)

	ship.shipState.XVel *= velDampen
	ship.shipState.YVel *= velDampen

	ship.shipState.Rot += ship.shipState.RotVel
	if ship.shipState.Rot > math.Pi*2 {
		ship.shipState.Rot -= math.Pi * 2
	} else if ship.shipState.Rot < 0 {
		ship.shipState.Rot += math.Pi * 2
	}
	ship.shipState.RotVel *= rotDampen

	if ship.needsUpdate {
		update.Ships = append(update.Ships, ship.shipState)
		ship.needsUpdate = false
	}
}

func (ship Ship) takeDamage(damage int32, updates *pb.GameState) {
	ship.shipState.Health -= damage
	if ship.shipState.Health <= 0 {
		ship.shipState.State = pb.State_DEAD
		ship.lastStateChange = time.Now().UnixNano()
		ship.needsUpdate = true
	}
}
