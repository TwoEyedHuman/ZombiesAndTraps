package main

import (
	"os"
	"image"
	_"image/png"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"fmt"
	"math"
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
	Player entity `json:"player"`
	Zombies []entity `json:"zombies"`
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
	name string
	sprite *pixel.Sprite
	pos intVec
	facing intVec
	health int
	pack []item
	properties property
	displacement pixel.Vec
	secondsPerTile float64
}

type item struct {
	sprite *pixel.Sprite
	pos intVec
	health int
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
	gameMap.player.displacement.X = math.max(float64(0), gameMap.displacement.X - (dt/gameMap.player.secondsPerTile))
	gameMap.player.displacement.Y = max(float64(0), gameMap.displacement.Y - (dt/gameMap.player.secondsPerTile))

	//update opponents displacement
	for opp, _ := range gameMap.opponents {
		opp.displacement.X = max(float64(0), opp.displacement.X - (dt/opp.secondsPerTile))
		opp.displacement.Y = max(float64(0), opp.displacement.Y - (dt/opp.secondsPerTile))
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

func isValidMove(toPos intVec, gameMap mapObject) {
	for lyr := range gameMap.Layers { //iterate through each layer
		if lyr.Properties.Collision && lyr.Data[toPos.X * mapHeight + toPos.Y] > 0{ //check that this layer has collision enabled and that the to position has something there
			return false  //not a valid move
		}
	}

	for itm := range gameMap.items { //iterate through items in field and if collision, then it is not a valid move
		if !itm.properties.Collision && toPos == itm.pos {
			return false
		}
	}

	//collision with player
	if intVecEqual(toPos, gameMap.player.pos) {
		return false
	}

	//collision with enemies
	for opp := range gameMap.opponents {
		if toPos == opp.pos {
			return false
		}
	}

	return true //passed all tests, is a valid move
}

func posToVec(pos intVec) (v pixel.Vec) {
	v.X = float64(pixelPerGrid * pos.X)
	v.Y = float64(pixelPerGrid * pos.Y)
	return
}

func setEntityData(toEntity *entity, fromEntity entity) {
	if fromEntity.name != nil {
		toEntity.name = fromEntity.name
	}
	if fromEntity.pos != nil {
		toEntity.pos.X = fromEntity.pos.X
		toEntity.pos.Y = fromtEntity.pos.Y
	}
	if fromEntity.facing != nil {
		toEntity.facing.X = fromEntity.facing.X
		toEntity.facing.Y = fromEntity.facing.Y
	}
	if fromEntity.health != nil {
		toEntity.health = fromEntity.health
	}
}

func loadmap(mapImageFile string, mapStructureFile string) (returnMap mapObject) {
	xmlMapStructure, err := os.Open(mapStructureFile)
	if err != nil {
		fmt.Printf("Error loading map initialization data file: %s\n", err)
	}

	defer xmlMapStructure.Close()

	byteValue, _ := ioutil.ReadAll(xmlMapStructure)
	xml.Unmarshal(byteValue, &returnMap)

	mapImage, err := loadPicture(mapImageFile)
	if err!= nil {
		fmt.Printf("Error loading map image data file: %s\n", err)
	}
	returnMap.sprite = pixel.NewSprite(mapImage, mapImage.Bounds())
}

func initializeGame(gameFile string, inputMap mapObject) mapObject {
	//Build game initialization data structure
	xmlMapStructure, err := os.Open(gameFile)
	if err != nil {
		fmt.Printf("Error loading game initialization data file: %s\n", err)
	}
	defer xmlMapStructure.Close()
	byteValue, _ := ioutil.ReadAll(xmlMapStructure)
	var iniMap initializationData
	xml.Unmarshal(byteValue, &iniMap)
	
	//Load into the game map
	setEntityData(mapObject.player, iniMap.Player)
	for zom, _ := range iniMap.Zombies {
		inputMap.opponents = append(inputMap.opponents, zom)
	}
	
	return inputMap
}

func gameOverCondition(gameMap mapObject) (isGameOver bool) {
	isGameOver = false
	for opp, _ := range gameMap.opponents {
		if intVecEqual(gameMap.player.pos, opp.pos) {
			isGameOver = true
		}
	}
	return
}

func run() {
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
			       isValidMove(addIntVec(gameMap.player.pos, intVec{0,1})) {
				gameMap.player.pos = addIntVec(gameMap.player.pos, intVec{0,1})
			} else if win.Pressed(pixelgl.KeyDown) &&
				   isValidMove(addIntVec(gameMap.player.pos, intVec{0,-1})){
				gameMap.player.pos = addIntVec(gameMap.player.pos, intVec{0,-1})
			} else if win.Pressed(pixelgl.KeyLeft) &&
 				   isValidMove(addIntVec(gameMap.player.pos, intVec{1,0})){
				gameMap.player.pos = addIntVec(gameMap.player.pos, intVec{1,0})
			} else if win.Pressed(pixelgl.KeyRight) && 
				   isValidMove(addIntVec(gameMap.player.pos, intVec{1,0})){
				gameMap.player.pos = addIntVec(gameMap.player.pos, intVec{1,0})
			}
	
			//update time based objects or values
			dt = time.Since(last).Seconds()
			last = time.Now()
			gameMap = updateDisplacements(gameMap, dt)

			//check end-game conditions
			isGameOver = gameOverCondition(gameMap)
	
			//display everything
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center())) //display the map
			gameMap.player.sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(gameMap.player.pos))) //display the player
			for opp, _ := range gameMap.opponents { //iterate through and display the opponents
				opp.sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(opp.pos)))
			}
			if win.Pressed(pixel.KeySpace) { //check if the menu button is pressed and display everything in the backpack if so
				for itm := range gameMap.player.pack {
					itm.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(win.Bounds().Center()))
				}
			}
	
		} else {
			//display the end game graphic to window
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
			gameMap.player.sprite.Draw(win, pixel.IM.Scaled(pizel.ZV, 1).Moved(posToVec(gameMap.player.sprite.pos)))
		}
	}
}
	
func main() {
	pixelgl.Run(run)
}
