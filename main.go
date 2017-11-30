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

type mapObject struct {
	height int
	layers []mapLayer
	nextobjectid int
	orientation string
	renderorder string
	tiledversion string
	tileheight int
	
}

func run() {
	//Load the map into a structure

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
