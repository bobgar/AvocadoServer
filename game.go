package main

import (
	pb "./proto"
	"golang.org/x/net/websocket"
)

var rotAcceleration float32 = .01
var thrust float32 = .5

var worldWidth float32 = 1024
var worldHeight float32 = 900

var maxVel float32 = 20
var maxRotVel float32 = .3
var bulletTTL int64 = 1500
var rateOfFire int64 = 250

type Game struct {
	clients     map[int32]*websocket.Conn
	ships       map[int32]*pb.Ship
	shipUpdates map[int32]*pb.ShipUpdate
	state       pb.GameState
}
