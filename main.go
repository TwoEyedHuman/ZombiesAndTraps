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

func run() {
	//Load the map into a structure
	gameMap := loadMap(
	//Initialize items, zombies, player

	//Launch program
	for !win.Closed() { //close the program when the user hits the X
		//read and reach to inputs

		//update time based objects or values

		//display map, items, zombies, player
	}
}

func main() {
	pixelgl.Run(run)
}
