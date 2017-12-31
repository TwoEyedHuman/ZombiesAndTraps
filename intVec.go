package main

import (
	"github.com/faiface/pixel"
)

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

func posToVec(pos intVec) (v pixel.Vec) { //convert the intVec to a pixel vector for display purposes
	v.X = float64(pixelHeight * pos.X + pixelHeight/2) //adjust horizontal position to display
	v.Y = float64(pixelHeight * pos.Y - pixelHeight/2) //adjsut vertical position to display
	return //return v
}
