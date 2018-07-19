package main

import "time"

type Ability interface {
	use() bool
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

func (gun Gun) use() bool {
	now := time.Now()
	timestamp := now.UnixNano()
	if (timestamp-gun.lastFire)/1000000 > gun.rateOfFire {
		gun.lastFire = timestamp
		return true
	} else {
		return false
	}
}

type BulletDef struct {
	radius     float32
	timeToLive float64
}
