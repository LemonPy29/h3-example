package main

import (
	"testing"

	"github.com/uber/h3-go/v4"
)

func BenchmarkComparison(b *testing.B) {
	value := h3.IndexFromString("8629a19afffffff")
	one := h3.Cell(value)
	if !one.IsValid() {
		panic("not a valid cell")
	}

	states := states()
	cells := make([][]h3.Cell, len(states.States))

	for i, state := range states.States {
		poly := state.Polygon()
		currentCells, _ := poly.Cells(3)
		cells[i] = currentCells
	}

	for b.Loop() {
		var matches uint8
		for _, cs := range cells {
			for _, cell := range cs {
				p, _ := one.Parent(cell.Resolution())
				if p == cell {
					matches++
					println(cell)
				}
			}
		}
	}
}
