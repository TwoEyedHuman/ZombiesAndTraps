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
}

type entity struct { //players and NPC
	name string
	sprite *pixel.Sprite
	pos intVec
	facing intVec
	health int
	pack []item
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
	return true //passed all tests, is a valid move
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
		if win.Pressed(pixelgl.KeyUp) {
		} else if win.Pressed(pixelgl.KeyDown) {
		} else if win.Pressed(pixelgl.KeyLeft) {
		} else if win.Pressed(pixelgl.KeyRight) {
		}

		//update time based objects or values
	}
		//display map, items, zombies, player
		if isGameOver {
			//display the end game graphic to window
		}
	}

}

func main() {
	pixelgl.Run(run)
}
