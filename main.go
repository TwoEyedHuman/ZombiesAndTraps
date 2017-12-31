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
	"math/rand"
)

const mapHeight int = 16 //map is mapHeight x mapHeight tiles
const pixelHeight int = 32 //every tile is pixelHeight x pixelHeight pixels

type mapObject struct { //object that holds the properties of the map itself
	Height int `json:"height"` //number of tiles in the map (across)
	Layers []mapLayer `json:"layers"` //Each map has multiple layers with tiles in each
	Nextobjectid int `json:"nextobjectid"` //Unknown at this time
	Orientation string `json:"orientation"` //How the tiles are structured (most common is orthogonal)
	Properties property `json:"properties"` //Properties of the map that are universal to lal layers
	Renderorder string `json:"renderorder"` //Unknown at this time
	Tiledversion string `json:"tiledversion"` //Version of the program Tiled that generated this map
	Tileheight int `json:"tileheight"` //Number of pixels that the tile is up
	Tilewidth int `json:"tilewidth"` //Number of pixels that the tile is across
	Type string `json:"type"` //Unknown at this time
	Version int `json:"version"` //Version of this map that is being used
	Width int `json:"width"` //number of tile in the map (up)
	sprite *pixel.Sprite //Rasterized image of all layers
	opponents []entity //Set of entities that represent the opposing force
	player entity //The entity that the player controls
	items []entity //Items that are interactable in the map, can be placed in players pack
}

type initializationData struct {
	Player entity `json:"player"` //Initial data for the player
	Zombies []entity `json:"zombies"` //Initial data for the opposing forces
}

type mapLayer struct { //each map has an associated layer
	Data []int `json:"data"` //Each layers matrix containing an index to the tile image to use
	Height int `json:"height"` //Number of tiles up
	Name string `json:"name"`  //Name given to the layer
	Opacity int `json:"opacity"` //How "see-through" is the layer
	Properties property `json:"properties"` //Properties of the layer (for example, whether collision is on)
	Type string `json:"type"` //Unknown at this time
	Visible bool `json:"visible"` //Boolean whether the layer is visible
	Width int	`json:"width"` //Number of tiles across
	X int `json:"x"` //Unknown at this time
	Y int `json:"y"` //Unknown at this time
}

type property struct { //each map and layer has a set of properties that it can hold
	Collision bool //Should the player be able to walk into the tile
	pickupable bool //is the player or entity able to pick up and store the item?
}

type entity struct { //players, items, and NPC
	Name string `json:"name"` //Name of the entity
	Sprite *pixel.Sprite //Image to use to represent the entity
	Pos intVec `json:"pos"` //integer based vector to determine its position in the grid
	Facing intVec `json:"facing"` //integer based vector to determine which direction the character is facing (used in determining which sprite to use)
	Health int `json:"health"` //health value of entity
	Pack []entity //set of items that are attached to character
	Properties property `json:"properties"` //Properties that relate to the entity, for example collision with opposing forces
	Displacement pixel.Vec `json:"displacement"` //Used in creating smooth movement, represents the amount of space left to move by entity between tiles
	Secondspertile float64 `json:"secondspertile"` //movement time between tiles
	Spritepath string `json:"spritepath"` //the filepath to the sprite image
}

type intVec struct {
	X int //horizontal displacement
	Y int //vertical displacement
}

func addIntVec(v1 intVec, v2 intVec) (v3 intVec) {
	v3.X = v1.X + v2.X //add the horizontal components
	v3.Y = v1.Y + v2.Y //add the vertical components
	return
}

func intVecEqual(v1 intVec, v2 intVec) bool { //compares the integer values of an intVec and determines if the components are equal
	var isEqual bool
	if v1.X == v2.X && v1.Y == v2.Y { //if both components are equal, then true
		isEqual = true
	} else {
		isEqual = false //at least one of the components did not match
	}
	return isEqual
}

func updateDisplacements(gameMap mapObject, dt float64) mapObject { //update the displacement of player and all enemies
	//update player displacement, if less than zero then return zero
	gameMap.player.Displacement.X = math.Max(float64(0), gameMap.player.Displacement.X - (dt/gameMap.player.Secondspertile))
	gameMap.player.Displacement.Y = math.Max(float64(0), gameMap.player.Displacement.Y - (dt/gameMap.player.Secondspertile))

	//update enemy displacements, if less than zero then return zero
	for _, opp := range gameMap.opponents {
		opp.Displacement.X = math.Max(float64(0), opp.Displacement.X - (dt/opp.Secondspertile))
		opp.Displacement.Y = math.Max(float64(0), opp.Displacement.Y - (dt/opp.Secondspertile))
	}

	return gameMap
}

func loadPicture(path string) (pixel.Picture, error) { //given a file path, create a picture object
	file, err := os.Open(path) //open the file
	if err != nil {
		return nil, err
	}
	defer file.Close() //delay closing the file until program exits
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil //return image object and no error signal
}

func isValidMove(toPos intVec, gameMap mapObject) (isValid bool) {
	isValid = true //initial value prior to testing invalid conditions

	if toPos.X >= mapHeight || toPos.Y >= mapHeight { //check that the to position is inside the square
		isValid = false
	}

	for _, lyr := range gameMap.Layers { //iterate through each layer
		if lyr.Properties.Collision &&
				lyr.Data[256 - toPos.Y * mapHeight + toPos.X] > 0{ //check that this layer has collision enabled and that the to position has something there
			isValid = false  //not a valid move
		}
	}

	for _, itm := range gameMap.items { //iterate through items in field and if collision, then it is not a valid move
		if !itm.Properties.Collision && intVecEqual(toPos, itm.Pos) {
			isValid = false
		}
	}

	//collision with player
	if intVecEqual(toPos, gameMap.player.Pos) {
		isValid = false
	}

	//collision with enemies
	for _, opp := range gameMap.opponents {
		if intVecEqual(toPos, opp.Pos) {
			isValid = false
		}
	}

	return //return isValid
}

func posToVec(pos intVec) (v pixel.Vec) { //convert the intVec to a pixel vector for display purposes
	v.X = float64(pixelHeight * pos.X + pixelHeight/2) //adjust horizontal position to display
	v.Y = float64(pixelHeight * pos.Y - pixelHeight/2) //adjsut vertical position to display
	return //return v
}

func loadMap(mapImageFile string, mapStructureFile string) (returnMap mapObject) { //load the map initializtion file (structure and image) into the map object
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

	playerImage, err := loadPicture(iniData.Player.Spritepath) //load the players sprite
	if err!= nil {
		fmt.Printf("Error loading player image data file (%s): %s\n", iniData.Player.Spritepath, err)
		panic(err)
	}
	//transfer all player details to the map
	outMap.player.Health = iniData.Player.Health
	outMap.player.Pos.X = iniData.Player.Pos.X
	outMap.player.Pos.Y = iniData.Player.Pos.Y
	outMap.player.Facing.X = iniData.Player.Facing.X
	outMap.player.Facing.Y = iniData.Player.Facing.Y
	outMap.player.Sprite = pixel.NewSprite(playerImage, playerImage.Bounds())

	for _, zom := range iniData.Zombies { //add the zombies to the opponents array
		fmt.Printf("Adding a zombie!")
		zomImage, err := loadPicture(zom.Spritepath)
		if err != nil {
			fmt.Printf("Error loading zombie image data file (%s): %s\n", zom.Spritepath, err)
			panic(err)
		}
		zom.Sprite = pixel.NewSprite(zomImage, zomImage.Bounds())
		outMap.opponents = append(outMap.opponents, zom)
	}

	return outMap
}

func intAbs (x int) (y int) {
	if x >0 {
		y = x
	} else {
		y = -1*x
	}
	return
}

func oppChase(plr intVec, opp intVec) (retVec intVec) {
	moveHorizontal := false
	//if the player and opponent are on an angle, randomly choose a direction toward player
	if plr.X - opp.X != 0 && plr.Y - opp.Y != 0 && rand.Intn(2) == 0 {
		moveHorizontal = true
	}

	if (plr.X == opp.X && plr.Y != opp.Y) || moveHorizontal { //if player and opponent are horizontal (X == X), move toward player
		retVec.X = 0
		retVec.Y = int((plr.Y - opp.Y)/intAbs(plr.Y - opp.Y))
	} else { //if player and opponent are vertical (Y == Y), move toward player
		retVec.Y = 0
		retVec.X = int((plr.X - opp.X)/intAbs(plr.X - opp.X))
	}
	return
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
			if win.JustPressed(pixelgl.KeyUp) && 
			       isValidMove(addIntVec(gameMap.player.Pos, intVec{0,1}), gameMap) {
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{0,1})
			} else if win.JustPressed(pixelgl.KeyDown) &&
				   isValidMove(addIntVec(gameMap.player.Pos, intVec{0,-1}), gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{0,-1})
			} else if win.JustPressed(pixelgl.KeyLeft) &&
 				   isValidMove(addIntVec(gameMap.player.Pos, intVec{-1,0}), gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{-1,0})
			} else if win.JustPressed(pixelgl.KeyRight) && 
				   isValidMove(addIntVec(gameMap.player.Pos, intVec{1,0}), gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{1,0})
			}
	
			if win.JustPressed(pixelgl.KeySpace) {
				oppMove := oppChase(gameMap.player.Pos, gameMap.opponents[0].Pos)
				if isValidMove(addIntVec(gameMap.opponents[0].Pos, oppMove), gameMap) {
					gameMap.opponents[0].Pos = addIntVec(gameMap.opponents[0].Pos, oppMove)
				}
			}

			//update time based objects or values
			dt = time.Since(last).Seconds()
			last = time.Now()
			gameMap = updateDisplacements(gameMap, dt)

			//check end-game conditions
			isGameOver = gameOverCondition(gameMap)
			//display everything
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center())) //display the map
			for _, opp := range gameMap.opponents {
				fmt.Printf("Drawing an opponent using %s at (%d,%d).\n", opp.Spritepath, opp.Pos.X, opp.Pos.Y)
				opp.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1). Moved(posToVec(opp.Pos)))
			}
			gameMap.player.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(gameMap.player.Pos))) //display the player

			if win.Pressed(pixelgl.KeySpace) { //check if the menu button is pressed and display everything in the backpack if so
				for _, itm := range gameMap.player.Pack {
					itm.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(win.Bounds().Center()))
				}
			}
			win.Update()
		} else {
			//display the end game graphic to window
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
			for _, opp := range gameMap.opponents {
				fmt.Println("Drawing an opponent.")
				opp.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1). Moved(posToVec(opp.Pos)))
			}
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
