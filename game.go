package main

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
	gameWidth:   1024,
	gameHeight:  1024}

type Game struct {
	id      int32
	players map[int32]*Player
	*GameDef
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
	gameWidth  float32
	gameHeight float32
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
