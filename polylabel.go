package polylabel

import (
	"container/heap"
	"math"

	"github.com/tidwall/geojson/geometry"
)

type Cell struct {
	x   float64
	y   float64
	h   float64
	d   float64
	max float64
}

func NewCell(x float64, y float64, h float64, polygon *geometry.Poly) *Cell {
	d := pointToPolygonDistance(x, y, polygon)
	cell := Cell{x, y, h, d, d + h*math.Sqrt2}
	return &cell
}

func NewCellItem(cell *Cell) *Item {
	return &Item{cell, cell.d, 0}
}

func Polylabel(polygon *geometry.Poly, precision float64) (float64, float64) {
	minX, minY, maxX, maxY := boundingBox(polygon)

	width := maxX - minX
	height := maxY - minY
	cellSize := math.Min(width, height)
	h := cellSize / 2

	if cellSize == 0 {
		return minX, minY
	}

	cellQueue := make(PriorityQueue, 0)

	// cover polygon with initial cells
	for x := minX; x < maxX; x += cellSize {
		for y := minY; y < maxY; y += cellSize {
			heap.Push(&cellQueue, NewCellItem(NewCell(x+h, y+h, h, polygon)))
		}
	}

	// take centroid as the first best guess
	bestCell := getCentroidCell(polygon)

	// special case for rectangular polygons
	bboxCell := NewCell(minX+width/2, minY+height/2, 0, polygon)
	if bboxCell.d > bestCell.d {
		bestCell = bboxCell
	}

	for cellQueue.Len() > 0 {
		// pick the most promising cell from the queue
		cellItem := heap.Pop(&cellQueue).(*Item)
		cell := cellItem.value

		// update the best cell if we found a better one
		if cell.d > bestCell.d {
			bestCell = cell
		}

		// do not drill down further if there's no chance of a better solution
		if (cell.max - bestCell.d) <= precision {
			continue
		}

		// split the cell into four cells
		h = cell.h / 2
		heap.Push(&cellQueue, NewCellItem(NewCell(cell.x-h, cell.y-h, h, polygon)))
		heap.Push(&cellQueue, NewCellItem(NewCell(cell.x+h, cell.y-h, h, polygon)))
		heap.Push(&cellQueue, NewCellItem(NewCell(cell.x-h, cell.y+h, h, polygon)))
		heap.Push(&cellQueue, NewCellItem(NewCell(cell.x+h, cell.y+h, h, polygon)))
	}

	return bestCell.x, bestCell.y
}

func boundingBox(polygon *geometry.Poly) (minX float64, minY float64, maxX float64, maxY float64) {
	rect := polygon.Rect()
	minX, minY, maxX, maxY = rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y
	return
}

// signed distance from point to polygon outline (negative if point is outside)
func pointToPolygonDistance(x float64, y float64, polygon *geometry.Poly) float64 {
	inside := false
	minDistSq := math.Inf(1)

	fn := func(ring geometry.Ring) {
		for n := 0; n < (ring.NumPoints() - 1); n++ {
			a := ring.PointAt(n)
			b := ring.PointAt(n + 1)
			if ((a.Y > y) != (b.Y > y)) && (x < ((b.X-a.X)*(y-a.Y)/(b.Y-a.Y) + a.X)) {
				inside = !inside
			}
			minDistSq = math.Min(minDistSq, segmentDistanceSquared(x, y, a, b))
		}
	}

	fn(polygon.Exterior)
	for _, hole := range polygon.Holes {
		fn(hole)
	}

	factor := 1.0
	if !inside {
		factor = -1.0
	}
	return factor * math.Sqrt(minDistSq)
}

// get polygon centroid
func getCentroidCell(polygon *geometry.Poly) *Cell {
	area := 0.0
	x := 0.0
	y := 0.0
	ring := polygon.Exterior
	for n := 0; n < ring.NumPoints()-1; n++ {
		a := ring.PointAt(n)
		b := ring.PointAt(n + 1)
		f := a.X*b.Y - b.X*a.Y
		x += (a.X + b.X) * f
		y += (a.Y + b.Y) * f
		area += f * 3
	}
	if area == 0 {
		p := ring.PointAt(0)
		return NewCell(p.X, p.Y, 0, polygon)
	}
	return NewCell(x/area, y/area, 0, polygon)
}

// get squared distance from a point to a segment
func segmentDistanceSquared(px float64, py float64, a, b geometry.Point) float64 {
	x := a.X
	y := a.Y
	dx := b.X - x
	dy := b.Y - y

	if dx != 0 || dy != 0 {
		t := ((px-x)*dx + (py-y)*dy) / (dx*dx + dy*dy)
		if t > 1 {
			x = b.X
			y = b.Y
		} else if t > 0 {
			x += dx * t
			y += dy * t
		}
	}

	dx = px - x
	dy = py - y

	return dx*dx + dy*dy
}
