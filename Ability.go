package main

import (
	"time"

	pb "./proto"
)

type Ability interface {
	use(Player) bool
}

type GunDef struct {
	// Rate of fire in milliseconds
	rateOfFire int64
	// Definition of the bullet to be fired
	bulletDef BulletDef
}

type Gun struct {
	// Definition of the gun
	GunDef
	// Last fire time
	lastFire int64
}

func (gun Gun) use(player Player) bool {
	timestamp := time.Now().UnixNano()
	if (timestamp-gun.lastFire)/1000000 > gun.rateOfFire {
		gun.lastFire = timestamp
		game := games[player.gameId]

		game.bullets[game.bulletId] = &Bullet{
			&pb.Bullet{
				Id:      game.bulletId,
				OwnerId: player.id,
				Type:    pb.BulletType_NORMAL,
				State:   pb.State_SPAWN,
				XPos:    player.ship.shipState.XPos,
				YPos:    player.ship.shipState.YPos,
				XVel:    player.ship.shipState.XVel,
				YVel:    player.ship.shipState.YVel},
			BulletDef{
				damage:     gun.bulletDef.damage,
				radius:     gun.bulletDef.radius,
				timeToLive: gun.bulletDef.timeToLive},
			timestamp}
		game.bulletId++
		return true
	} else {
		return false
	}
}
