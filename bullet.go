package main

import (
	pb "./proto"
)

type Bullet struct {
	bulletState *pb.Bullet
	BulletDef
	timestamp int64
}

type BulletDef struct {
	damage     int32
	radius     float32
	timeToLive int64
}

func (bullet Bullet) update(game *Game, timestamp int64, updates *pb.GameState) {
	//Delete timed out bullets
	if (timestamp-bullet.timestamp)/1000000 > bullet.timeToLive {
		//ship.Bullets = append(ship.Bullets[:i], ship.Bullets[i+1:]...)
		delete(game.bullets, bullet.bulletState.Id)
	} else {
		bullet.bulletState.XPos += bullet.bulletState.XVel
		bullet.bulletState.YPos += bullet.bulletState.YVel

		if bullet.bulletState.XPos < 0 {
			bullet.bulletState.XPos += game.width
		} else if bullet.bulletState.XPos > game.width {
			bullet.bulletState.XPos -= game.width
		}
		if bullet.bulletState.YPos < 0 {
			bullet.bulletState.YPos += game.height
		} else if bullet.bulletState.YPos > game.height {
			bullet.bulletState.YPos -= game.height
		}

		if bullet.bulletState.State == pb.State_SPAWN {

		} else {
			bulletOwner := game.players[bullet.bulletState.OwnerId]

			//TODO Test colision
			for _, player := range game.players {
				if player.ship != nil && player.ship.team != bulletOwner.ship.team {
					deltaX := bullet.bulletState.XPos - player.ship.shipState.XPos
					deltaY := bullet.bulletState.YPos - player.ship.shipState.YPos
					if bulletRadius*bulletRadius > (deltaX*deltaX + deltaY*deltaY) {
						delete(game.bullets, bullet.bulletState.Id)

						player.ship.takeDamage(bullet.damage, updates)

						bullet.bulletState.State = pb.State_DESTROY
					}
				}
			}
		}
	}
}
