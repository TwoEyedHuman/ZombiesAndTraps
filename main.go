package main

import (
	"os"
	"encoding/json"
	"image"
	_"image/png"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"fmt"
	"math"
	"io/ioutil"
	"time"
)

const mapHeight int = 16 //map is mapHeight x mapHeight tiles
const pixelHeight int = 32 //every tile is pixelHeight x pixelHeight pixels

type mapObject struct { //object that holds the properties of the map itself
	Height int `json:"height"`
	Layers []mapLayer `json:"layers"`
	Nextobjectid int `json:"nextobjectid"`
	Orientation string `json:"orientation"`
	Properties property `json:"properties"`
	Renderorder string `json:"renderorder"`
	Tiledversion string `json:"tiledversion"`
	Tileheight int `json:"tileheight"`
	Tilewidth int `json:"tilewidth"`
	Type string `json:"type"`
	Version int `json:"version"`
	Width int `json:"width"`
	sprite *pixel.Sprite
	opponents []entity
	player entity
	items []entity
}

type initializationData struct {
	Loadtype string `json:"loadtype"`
	Player entity `json:"player"`
//	Zombies []entity `json:"zombies"`
}

type mapLayer struct { //each map has an associated layer
	Data []int `json:"data"`
	Height int `json:"height"`
	Name string `json:"name"`
	Opacity int `json:"opacity"`
	Properties property `json:"properties"`
	Type string `json:"type"`
	Visible bool `json:"visible"`
	Width int	`json:"width"`
	X int `json:"x"`
	Y int `json:"y"`
	sprite *pixel.Sprite
}

type property struct { //each map and layer has a set of properties that it can hold
	Collision bool
	pickupable bool //is the player or entity able to pick up and store the item?
}

type entity struct { //players and NPC
	Name string `json:"name"`
	Sprite *pixel.Sprite
	Pos intVec `json:"pos"`
	Facing intVec `json:"facing"`
	Health int `json:"health"`
	Pack []entity
	Properties property `json:"properties"`
	Displacement pixel.Vec `json:"displacement"`
	Secondspertile float64 `json:"secondspertile"`
	Spritepath string `json:"spritepath"`
}

type intVec struct {
	X int
	Y int
}

func addIntVec(v1 intVec, v2 intVec) (v3 intVec) {
	v3.X = v1.X + v2.X
	v3.Y = v1.Y + v2.Y
	return
}

func intVecEqual(v1 intVec, v2 intVec) bool {
	var isEqual bool
	if v1.X == v2.X && v1.Y == v2.Y {
		isEqual = true
	} else {
		isEqual = false
	}
	return isEqual
}

func updateDisplacements(gameMap mapObject, dt float64) mapObject {
	//update player displacement
	gameMap.player.Displacement.X = math.Max(float64(0), gameMap.player.Displacement.X - (dt/gameMap.player.Secondspertile))
	gameMap.player.Displacement.Y = math.Max(float64(0), gameMap.player.Displacement.Y - (dt/gameMap.player.Secondspertile))

	//update opponents displacement
	for _, opp := range gameMap.opponents {
		opp.Displacement.X = math.Max(float64(0), opp.Displacement.X - (dt/opp.Secondspertile))
		opp.Displacement.Y = math.Max(float64(0), opp.Displacement.Y - (dt/opp.Secondspertile))
	}
	
	return gameMap
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func isValidMove(toPos intVec, gameMap mapObject) bool {
	for _, lyr := range gameMap.Layers { //iterate through each layer
		if lyr.Properties.Collision && lyr.Data[toPos.X * mapHeight + toPos.Y] > 0{ //check that this layer has collision enabled and that the to position has something there
			return false  //not a valid move
		}
	}

	for _, itm := range gameMap.items { //iterate through items in field and if collision, then it is not a valid move
		if !itm.Properties.Collision && toPos == itm.Pos {
			return false
		}
	}

	//collision with player
	if intVecEqual(toPos, gameMap.player.Pos) {
		return false
	}

	//collision with enemies
	for _, opp := range gameMap.opponents {
		if toPos == opp.Pos {
			return false
		}
	}

	return true //passed all tests, is a valid move
}

func posToVec(pos intVec) (v pixel.Vec) {
	v.X = float64(pixelHeight * pos.X)
	v.Y = float64(pixelHeight * pos.Y)
	return
}



func loadMap(mapImageFile string, mapStructureFile string) (returnMap mapObject) {
	jsonMapStructure, err := os.Open(mapStructureFile)
	if err != nil {
		fmt.Printf("Error loading map initialization data file: %s\n", err)
		panic(err)
	}

	defer jsonMapStructure.Close()

	byteValue, _ := ioutil.ReadAll(jsonMapStructure)
	json.Unmarshal(byteValue, &returnMap)

	mapImage, err := loadPicture(mapImageFile)
	if err!= nil {
		fmt.Printf("Error loading map image data file: %s\n", err)
		panic(err)
	}

	returnMap.sprite = pixel.NewSprite(mapImage, mapImage.Bounds())
	return
}

func initializeGame(gameFile string, inMap mapObject) mapObject {
	outMap := mapObject(inMap)
	//Build game initialization data structure
	iDFile, err := os.Open(gameFile)
	if err != nil {
		fmt.Printf("Error loading game initialization data file: %s\n", err)
	}
	defer iDFile.Close()
	byteValue, _ := ioutil.ReadAll(iDFile)
	var iniData initializationData
	json.Unmarshal(byteValue, &iniData)

	//Load into the game map
/*	for _, zom := range iniMap.Zombies {
		inputMap.opponents = append(inputMap.opponents, zom)
	}*/

	playerImage, err := loadPicture(iniData.Player.Spritepath)
	if err!= nil {
		fmt.Printf("Error loading player image data file (%s): %s\n", iniData.Player.Spritepath, err)
		panic(err)
	}
	outMap.player.Health = iniData.Player.Health
	outMap.player.Pos.X = iniData.Player.Pos.X
	outMap.player.Pos.Y = iniData.Player.Pos.Y
	outMap.player.Facing.X = iniData.Player.Facing.X
	outMap.player.Facing.Y = iniData.Player.Facing.Y

	outMap.player.Sprite = pixel.NewSprite(playerImage, playerImage.Bounds())

	return outMap
}

func gameOverCondition(gameMap mapObject) (isGameOver bool) {
	isGameOver = false
	for _, opp := range gameMap.opponents {
		if intVecEqual(gameMap.player.Pos, opp.Pos) {
			isGameOver = true
		}
	}
	return
}

func run() {
	//Create the window to display the game
	cfg := pixelgl.WindowConfig {
		Title: "Zombies and Traps",
		Bounds: pixel.R(0,0,512,512),
		VSync: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	//Load the map into a structure
	gameMap := loadMap("map/bar_Image.png", "map/bar.json")
	gameMap = initializeGame("initializationData.json", gameMap)

	//Initialize items, zombies, player
	isGameOver := false //condition on if player won/lost

	last := time.Now()
	dt := time.Since(last).Seconds()
	//Launch program
	for !win.Closed() { //close the program when the user hits the X
		//display map, items, zombies, player
		if !isGameOver {
			//read and react to inputs
			//check positional movements and update as necessary
			if win.Pressed(pixelgl.KeyUp) && 
			       isValidMove(addIntVec(gameMap.player.Pos, intVec{0,1}), gameMap) {
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{0,1})
			} else if win.Pressed(pixelgl.KeyDown) &&
				   isValidMove(addIntVec(gameMap.player.Pos, intVec{0,-1}), gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{0,-1})
			} else if win.Pressed(pixelgl.KeyLeft) &&
 				   isValidMove(addIntVec(gameMap.player.Pos, intVec{1,0}), gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{1,0})
			} else if win.Pressed(pixelgl.KeyRight) && 
				   isValidMove(addIntVec(gameMap.player.Pos, intVec{1,0}), gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{1,0})
			}
	
			//update time based objects or values
			dt = time.Since(last).Seconds()
			last = time.Now()
			gameMap = updateDisplacements(gameMap, dt)

			//check end-game conditions
			isGameOver = gameOverCondition(gameMap)
			//display everything
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center())) //display the map
//			gameMap.player.sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(gameMap.player.pos))) //display the player
//			for _, opp := range gameMap.opponents { //iterate through and display the opponents
//				opp.sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(opp.pos)))
//			}

			if win.Pressed(pixelgl.KeySpace) { //check if the menu button is pressed and display everything in the backpack if so
				for _, itm := range gameMap.player.Pack {
					itm.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(win.Bounds().Center()))
				}
			}
			win.Update()
		} else {
			//display the end game graphic to window
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
			gameMap.player.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(gameMap.player.Pos)))
			win.Update()
		}
	}
}
	
func main() {
	pixelgl.Run(run)
}

//Functions that will not be used
/*
func setPlayerData(toEntity mapObject, fromEntity entity) mapObject {
	if fromEntity.name != "" {
		toEntity.player.name = fromEntity.name
	}
	if !intVecEqual(fromEntity.pos, intVec{0,0}) {
		toEntity.player.pos.X = fromEntity.pos.X
		toEntity.player.pos.Y = fromEntity.pos.Y
	}
	if !intVecEqual(fromEntity.facing, intVec{0,0}) {
		toEntity.player.facing.X = fromEntity.facing.X
		toEntity.player.facing.Y = fromEntity.facing.Y
	}
	if fromEntity.health == -1 {
		toEntity.player.health = fromEntity.health
	}

	return toEntity
}


*/
