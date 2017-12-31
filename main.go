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
	"btlmath"
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
	Fielditems []entity `json:"fielditems"` //Initial items laying in field
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
	Pickupable bool //is the player or entity able to pick up and store the item?
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
	DisplacementTime float64
	Secondspertile float64 `json:"secondspertile"` //movement time between tiles
	Spritepath string `json:"spritepath"` //the filepath to the sprite image
}

func updateDisplacements(gameMap mapObject, dt float64) mapObject { //update the displacement of player and all enemies
	returnMap := mapObject(gameMap)
	//update player displacement, if less than zero then return zero
	if gameMap.player.DisplacementTime > 0 {
		returnMap.player.Displacement.X = returnMap.player.Displacement.X - (dt/returnMap.player.Secondspertile)
		returnMap.player.Displacement.Y = returnMap.player.Displacement.Y - (dt/returnMap.player.Secondspertile)
	} else {
		returnMap.player.DisplacementTime = 0
		returnMap.player.Displacement = pixel.ZV
	}

	//update enemy displacements, if less than zero then return zero
	for i, opp := range returnMap.opponents {
		if opp.DisplacementTime > 0 {
			returnMap.opponents[i].Displacement.X = math.Max(float64(0), opp.Displacement.X - (dt/opp.Secondspertile))
			returnMap.opponents[i].Displacement.Y = math.Max(float64(0), opp.Displacement.Y - (dt/opp.Secondspertile))
		} else {
			returnMap.opponents[i].DisplacementTime = 0
			returnMap.opponents[i].Displacement = pixel.ZV
		}
	}

	return returnMap
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

func isValidMove(toPos intVec, itemCollision bool, gameMap mapObject) (isValid bool) { //checks that a position is able to be occupied by a player or opponent
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

	if itemCollision {
		for _, itm := range gameMap.items { //iterate through items in field and if collision, then it is not a valid move
			if !itm.Properties.Collision && intVecEqual(toPos, itm.Pos) {
				isValid = false
			}
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

func loadMap(mapImageFile string, mapStructureFile string) (returnMap mapObject) { //load the map initializtion file (structure and image) into the map object
	jsonMapStructure, err := os.Open(mapStructureFile) //open the json file containing map structure
	if err != nil {
		fmt.Printf("Error loading map initialization data file: %s\n", err)
		panic(err)
	}

	defer jsonMapStructure.Close() //delay closing the file

	byteValue, _ := ioutil.ReadAll(jsonMapStructure)
	json.Unmarshal(byteValue, &returnMap) //load the json data into a mapObject struture

	mapImage, err := loadPicture(mapImageFile) //load the map sprite into a usable object
	if err!= nil {
		fmt.Printf("Error loading map image data file: %s\n", err)
		panic(err)
	}

	returnMap.sprite = pixel.NewSprite(mapImage, mapImage.Bounds()) //apply the map picture as the sprite for the map
	return
}

func initializeGame(gameFile string, inMap mapObject) mapObject { //load the initial setup of the game into the map
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
		zomImage, err := loadPicture(zom.Spritepath)
		if err != nil {
			fmt.Printf("Error loading zombie image data file (%s): %s\n", zom.Spritepath, err)
			panic(err)
		}
		zom.Sprite = pixel.NewSprite(zomImage, zomImage.Bounds()) //set zombie sprite
		outMap.opponents = append(outMap.opponents, zom)
	}

	for _, itm := range iniData.Fielditems {
		itmImage, err := loadPicture(itm.Spritepath)
		if err != nil {
			fmt.Printf("Error loading an item image data file (%s): %s\n", itm.Spritepath, err)
			panic(err)
		}
		itm.Sprite = pixel.NewSprite(itmImage, itmImage.Bounds())
		outMap.items = append(outMap.items, itm)
	}

	return outMap
}

func oppChase(plr intVec, opp intVec) (retVec intVec) { //function that runs the next move for the opponent to make when chasing the player
	moveHorizontal := false
	//if the player and opponent are on an angle, randomly choose a direction toward player
	if plr.X - opp.X != 0 && plr.Y - opp.Y != 0 && rand.Intn(2) == 0 {
		moveHorizontal = true
	}

	if (plr.X == opp.X && plr.Y != opp.Y) || moveHorizontal { //move horizontally toward player
		retVec.X = 0
		retVec.Y = int((plr.Y - opp.Y)/btlmath.IntAbs(plr.Y - opp.Y))
	} else { //move vertically toward player
		retVec.Y = 0
		retVec.X = int((plr.X - opp.X)/btlmath.IntAbs(plr.X - opp.X))
	}
	return
}

func gameOverCondition(gameMap mapObject) (isGameOver bool) { //determines if the game is over or not
	isGameOver = false
	for _, opp := range gameMap.opponents {
		if intVecEqual(gameMap.player.Pos, opp.Pos) { //if an opponent occupies same space as player, then game is over
			isGameOver = true
		}
	}
	return
}
func playerPickup (gameMap mapObject) (returnMap mapObject) { //determine what to do when a player presses pickup
	itmIndex := -1
	if len(gameMap.items) > 0 {
		var itmPicked entity
		for i, itm := range gameMap.items {
			if intVecEqual(gameMap.player.Pos, itm.Pos) {
				itmPicked = itm	
				itmIndex = i
			}
		}
		if itmIndex >= 0 {
			gameMap.items[itmIndex] = gameMap.items[len(gameMap.items)-1]
			gameMap.items = gameMap.items[:len(gameMap.items)-1]
			gameMap.player.Pack = append(gameMap.player.Pack, itmPicked)
		}
	}
	if itmIndex == -1 && len(gameMap.player.Pack) > 0 {
		gameMap.player.Pack[0].Pos.X = gameMap.player.Pos.X
		gameMap.player.Pack[0].Pos.Y = gameMap.player.Pos.Y
		gameMap.items = append(gameMap.items, gameMap.player.Pack[0])
		gameMap.player.Pack[0] = gameMap.player.Pack[len(gameMap.player.Pack)-1]
		gameMap.player.Pack = gameMap.player.Pack[:len(gameMap.player.Pack)-1]
	}
	returnMap = gameMap
	return
}

func updateOppPos(gameMap mapObject) (returnMap mapObject) { //iterate over all opponents and update their position
	for i, opp := range gameMap.opponents {
		oppMove := oppChase(gameMap.player.Pos, opp.Pos)
		if isValidMove(addIntVec(opp.Pos, oppMove), true, gameMap) {
			gameMap.opponents[i].Pos = addIntVec(opp.Pos, oppMove)
		}
	}
	returnMap = gameMap
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
			//check positional movements and update as necessary
			if win.JustPressed(pixelgl.KeyUp) && 
			       isValidMove(addIntVec(gameMap.player.Pos, intVec{0,1}), false, gameMap) {
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{0,1})
			} else if win.JustPressed(pixelgl.KeyDown) &&
				   isValidMove(addIntVec(gameMap.player.Pos, intVec{0,-1}), false, gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{0,-1})
			} else if win.JustPressed(pixelgl.KeyLeft) &&
 				   isValidMove(addIntVec(gameMap.player.Pos, intVec{-1,0}), false, gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{-1,0})
			} else if win.JustPressed(pixelgl.KeyRight) && 
				   isValidMove(addIntVec(gameMap.player.Pos, intVec{1,0}), false, gameMap){
				gameMap.player.Pos = addIntVec(gameMap.player.Pos, intVec{1,0})
			}
	
			if win.JustPressed(pixelgl.KeyP) {
				gameMap = playerPickup(gameMap)
			}

			if win.JustPressed(pixelgl.KeyM) { //iterate over opponents to update their position
				gameMap = updateOppPos(gameMap)
			}

			//update time based objects or values
			dt = time.Since(last).Seconds()
			last = time.Now()
			gameMap = updateDisplacements(gameMap, dt)

			//check end-game conditions
			isGameOver = gameOverCondition(gameMap)
			//display everything
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center())) //display the map
			for _, opp := range gameMap.opponents { //display the opponents
				opp.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(opp.Pos)))
			}
			for _, itm := range gameMap.items { //display the field items
				itm.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(itm.Pos)))
			}
			gameMap.player.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(gameMap.player.Pos))) //display the player

			if win.Pressed(pixelgl.KeyN) { //check if the menu button is pressed and display everything in the backpack if so
				spaceDiff := pixel.Vec{float64(0), float64(0)}
				for _, itm := range gameMap.player.Pack {
					itm.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(win.Bounds().Center().Add(spaceDiff)))
					spaceDiff.X = spaceDiff.X + float64(itm.Sprite.Frame().Max.X - itm.Sprite.Frame().Min.X)
				}
			}
			win.Update()
		} else {
			//display the end game graphic to window
			gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
			for _, opp := range gameMap.opponents {
				opp.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1). Moved(posToVec(opp.Pos)))
			}
			gameMap.player.Sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(posToVec(gameMap.player.Pos)))
			win.Update()
		}
	}
}
	
func main() { //pixel uses the run function to perform all functions
	pixelgl.Run(run)
}
