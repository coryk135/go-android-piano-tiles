// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// An app that draws a green triangle on a red background.
package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/app/debug"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/gl/glutil"
)

var (
	program  		gl.Program
	position 		gl.Attrib
	offset   		gl.Uniform
	color    		gl.Uniform
	buf      		gl.Buffer

	tick        	int
	tiles       	[4]int
	numTiles    	int
	key         	int
	tileWidth   	geom.Pt
	tileHeight  	geom.Pt

	green    		float32
	touchLoc 		geom.Point

	drawChan 		chan string
	quit     		chan int

	animating		bool
	animateUntil	time.Time
	now				time.Time
	timeDiff		time.Duration
)

// type Square struct {
//     offsetX float32
//     offsetY float32
// }

func main() {
	app.Run(app.Callbacks{
		Start: start,
		Stop:  stop,
		Draw:  draw,
		Touch: touch,
	})
}

func start() {
	key = 0
	tiles[0] = 0
	tiles[1] = 2
	tiles[2] = 1
	tiles[3] = 3
	tick = 0
	numTiles = 4
	tileWidth = geom.Width/geom.Pt(numTiles)
	tileHeight = geom.Height/geom.Pt(numTiles)
	// squares := []Square{Square{0.0, 0.0},
	//                     Square{0.5, 0.5},
	//                     Square{1.0, 1.0},
	//                     Square{1.5, 1.5}}

	var err error
	program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = gl.GenBuffer()
	calcTriangleData()
	// gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	// gl.BufferData(gl.ARRAY_BUFFER, gl.STATIC_DRAW, triangleData)

	position = gl.GetAttribLocation(program, "position")
	color = gl.GetUniformLocation(program, "color")
	offset = gl.GetUniformLocation(program, "offset")
	touchLoc = geom.Point{geom.Width / 2, geom.Height / 2}

	// TODO(crawshaw): the debug package needs to put GL state init here


	drawChan = make(chan string)
	// go func (){
	// 	for {
	// 		select {
	// 			case drawChan <- "draw":

	// 			case <-quit:
	// 				return
	// 			default:
	// 		}
	// 	}
	// }
}

func stop() {
	quit <- 0
	gl.DeleteProgram(program)
	gl.DeleteBuffer(buf)
}

func touch(t event.Touch) {
	touchLoc = t.Loc
	if t.Type == event.TouchStart {
		if inbounds(t){
			tick = (tick+1) % 4
			shiftTiles()
			calcTriangleData()
			animating = true
			animateUntil = time.Now().Add(time.Second)
		} //else {

		//}
	}
}

func shiftTiles(){
	for i := 0; i < len(tiles)-1; i++ {
		tiles[i] = tiles[i+1]
	}
	tiles[len(tiles)-1] = randomKey()
	key = tiles[0]
}

func calcTriangleData(){
	t := triangleFloats

	// i is a tile's row (y position)
	// v is a tile's col (x position)
	for i,v := range tiles {
		b := i*6*3
		u := float32(v)
		j := float32(i)
		t[b+0]  = 0.0 + 0.5*u; t[b+1]  = 0.5 + 0.5*j // top left
		t[b+3]  = 0.0 + 0.5*u; t[b+4]  = 0.0 + 0.5*j // bottom left
		t[b+6]  = 0.5 + 0.5*u; t[b+7]  = 0.0 + 0.5*j // bottom right
		t[b+9]  = 0.0 + 0.5*u; t[b+10] = 0.5 + 0.5*j // top left
		t[b+12] = 0.5 + 0.5*u; t[b+13] = 0.5 + 0.5*j // top right
		t[b+15] = 0.5 + 0.5*u; t[b+16] = 0.0 + 0.5*j // bottom right
	}
	triangleData = f32.Bytes(binary.LittleEndian, triangleFloats...)
	// buf = gl.GenBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, gl.STATIC_DRAW, triangleData)
}

func inbounds(t event.Touch) bool{
	return geom.Pt(key)*tileWidth < t.Loc.X && t.Loc.X < geom.Pt((key+1))*tileWidth// &&
		//geom.Height > t.Loc.Y && t.Loc.Y > geom.Height - (geom.Pt((0+1))*tileHeight)

}

func randomKey() int{
	return rand.Int() % len(tiles)
}

func draw() {
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(program)

	// green += 0.01
	// if green > 1 {
	//  green = 0
	// }
	gl.Uniform4f(color, 0, 0, 0, 1)

	now = time.Now()
	timeDiff = animateUntil.Sub(now)
	if timeDiff < 0 {
		animating = false
	}
	if animating {
		gl.Uniform2f(offset, float32(-1), float32(timeDiff.Seconds()))
		// gl.Uniform2f(offset, float32(touchLoc.X/geom.Width), float32(touchLoc.Y/geom.Height))
	} else {
		gl.Uniform2f(offset, float32(-1), float32(0a))
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.EnableVertexAttribArray(position)
	gl.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	gl.DisableVertexAttribArray(position)

	debug.DrawFPS()
}

var triangleFloats = []float32{
	0.0, 0.5, 0.0, // top left
	0.0, 0.0, 0.0, // bottom left
	0.5, 0.0, 0.0, // bottom right
	0.0, 0.5, 0.0, // top left
	0.5, 0.5, 0.0, // top right
	0.5, 0.0, 0.0, // bottom right

	0.5, 1.0, 0.0, // top left
	0.5, 0.5, 0.0, // bottom left
	1.0, 0.5, 0.0, // bottom right
	0.5, 1.0, 0.0, // top left
	1.0, 1.0, 0.0, // top right
	1.0, 0.5, 0.0, // bottom right

	1.0, 1.5, 0.0, // top left
	1.0, 1.0, 0.0, // bottom left
	1.5, 1.0, 0.0, // bottom right
	1.0, 1.5, 0.0, // top left
	1.5, 1.5, 0.0, // top right
	1.5, 1.0, 0.0, // bottom right

	1.5, 1.0, 0.0, // top left
	1.5, 1.5, 0.0, // bottom left
	2.0, 1.5, 0.0, // bottom right
	1.5, 1.0, 0.0, // top left
	2.0, 1.0, 0.0, // top right
	2.0, 1.5, 0.0, // bottom right
}
var triangleData = f32.Bytes(binary.LittleEndian, triangleFloats...)

const (
	coordsPerVertex = 3
	vertexCount     = 6 * 4
)

const vertexShader = `#version 100
uniform vec2 offset;

attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	// vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	vec4 offset4 = vec4(-1, 0.5*offset.y-1.0, 0, 0); // vec4(-1.0, -1.0, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
