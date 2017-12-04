package main

import (
	"os"
	"image"
	_"image/png"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"fmt"
	"encoding/json"
	"bytes"
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
	pickupable bool
}

type entity struct { //players and NPC
	name string
	sprite *pixel.Sprite
	pos intVec
	facing intVec
	health int
	pack []item
	properties property
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
	if toPos == gameMap.player.pos {
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

func loadmap(mapImageFile string, mapStructureFile string) (returnMap mapObject) {
	xmlMapStructure, err := os.Open(mapStructureFile)
	if err != nul {
		fmt.Println(err)
	}

	defer xmlMapStructure.Close()

	byteValue, _ := ioutil.ReadAll(xmlMapStructure)
	xml.Unmarshal(byteValue, &returnMap)

	mapImage, err := loadPicture(mapImageFile)
	if err!= nil {
		panic(err)
	}
	returnMap.sprite = pixel.NewSprite(mapImage, mapImage.Bounds())
}

func run() {
	//Load the map into a structure
	gameMap := loadMap("map/bar_Image.png", "map/bar.json")

	//Initialize items, zombies, player

	isGameOver := false //condition on if player won/lost
	//Launch program
	for !win.Closed() { //close the program when the user hits the X

	if !isGameOver {
		//read and react to inputs
		//check positional movements
		if win.Pressed(pixelgl.KeyUp) && isValidMove(addIntVec(gameMap.player.pos, intVec{0,1})) {
		} else if win.Pressed(pixelgl.KeyDown) && isValidMove(addIntVec(gameMap.player.pos, intVec{0,-1})){
		} else if win.Pressed(pixelgl.KeyLeft) && {isValidMove(addIntVec(gameMap.player.pos, intVec{1,0}))
		} else if win.Pressed(pixelgl.KeyRight) && {isValidMove(addIntVec(gameMap.player.pos, intVec{1,0}))
		}

		//update time based objects or values

		//display everything
		gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		gameMap.player.sprite.Draw(win, pixel.IM.Scaled(pizel.ZV, 1).Moved(posToVec(gameMap.player.sprite.pos)))
		if win.Pressed(pixel.KeySpace) {
			for itm := range gameMap.player.pack {
				itm.Draw(win, pixel.IM.Scaled(pixel.ZV, 1).Moved(win.Bounds().Center()))
			}
		}

	}
		//display map, items, zombies, player
	if isGameOver {
		//display the end game graphic to window
		gameMap.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		gameMap.player.sprite.Draw(win, pixel.IM.Scaled(pizel.ZV, 1).Moved(posToVec(gameMap.player.sprite.pos)))
	}
	}

}

func main() {
	pixelgl.Run(run)
}
