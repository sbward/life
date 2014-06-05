package life

import (
	"encoding/json"
	//"log"
)

type Simulator interface {
	Cells() [][]bool
	SetCell(int, int, bool)
	Width() int
	Height() int
	Step()
	json.Marshaler

	// Stream returns a channel of World.Cells world generations and a
	// response channel to ask the simulation to stop and delete itself.
	Stream() (<-chan [][]bool, chan<- bool)
}

type World struct {
	Cells         [][]bool // [y][x] to match HTML tables.
	Width, Height int
	step          int
}

// Create an empty simulation from scratch
func New(width, height int) *World {
	cells := make([][]bool, height)
	for i := range cells {
		cells[i] = make([]bool, width)
	}
	return &World{cells, width, height, 0}
}

// NumAlive returns the number of alive cells.
func (w *World) NumAlive() int {
	var n int
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			if w.Cells[y][x] {
				n++
			}
		}
	}
	return n
}

func (w *World) findCell(x, y int) bool {
	// Allow overflow by wrapping around with remainder (modulus).
	x = x % w.Width
	y = y % w.Height
	// Allow negative indices by adding one width and/or height if necessary.
	if x < 0 {
		x += w.Width
	}
	if y < 0 {
		y += w.Height
	}
	return w.Cells[y][x]
}

/*
	Game of Life rules:
	Any live cell with fewer than two live neighbours dies (change)
	Any live cell with two or three live neighbours lives (no change)
	Any live cell with more than three live neighbours dies (change)
	Any dead cell with three live neighbours becomes a live cell (change)
*/

// Calculates the next step of a cell without changing it.
func (w *World) stepCell(x, y int) bool {
	oldCell := w.findCell(x, y)
	newCell := oldCell
	numAlive := 0

	// Loop through neighbors, counting the living ones.
	for i := (x - 1); i <= (x + 1); i++ {
		for j := (y - 1); j <= (y + 1); j++ {
			// Skip the center cell.
			if (i == x) && (j == y) {
				continue
			}
			alive := w.findCell(i, j)
			// Count living neighbors.
			if alive {
				numAlive++
			}
		}
	}

	//if oldCell {
	//log.Printf("Cell %d,%d (%d neighbors)\n", x, y, numAlive)
	//}

	// Enforce the game rules.
	if oldCell && (numAlive < 2 || numAlive > 3) {
		newCell = false
	} else if !oldCell && numAlive == 3 {
		newCell = true
	}

	return newCell
}

// Calculate next world generation.
func (w *World) Step() {
	var numAlive int
	var c bool
	next := New(w.Width, w.Height)
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			c = w.stepCell(x, y)
			next.Cells[y][x] = c
			if c {
				numAlive++
			}
		}
	}
	w.step++
	w.Cells = next.Cells
	return
}

// Produces a stream of World generations.
func (w *World) Stream() (<-chan World, chan<- bool) {
	send := make(chan World)
	stop := make(chan bool)

	go func() {
		defer close(send)
		defer close(stop)

	loop:
		for {
			// First check for a stop signal, then attempt to send a step.

			select {
			case <-stop:
				break loop
			case send <- *w:
				w.Step()
			}
		}
	}()

	return (<-chan World)(send), (chan<- bool)(stop)
}

func (w *World) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Cells)
}
