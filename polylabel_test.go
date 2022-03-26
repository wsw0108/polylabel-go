package polylabel

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/tidwall/geojson/geometry"
)

func AssertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

type Point [2]float64
type Ring []Point
type Polygon []Ring

func loadData(filename string) (polygon *geometry.Poly) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		panic("failed to open json file")
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var poly Polygon
	err = json.Unmarshal(byteValue, &poly)
	if err != nil {
		panic("failed to parse json file")
	}

	var shell []geometry.Point
	var holes [][]geometry.Point
	for i, ring := range poly {
		var points []geometry.Point
		for _, p := range ring {
			points = append(points, geometry.Point{X: p[0], Y: p[1]})
		}
		if i == 0 {
			shell = points
		} else {
			holes = append(holes, points)
		}
	}
	polygon = geometry.NewPoly(shell, holes, nil)

	return
}

func TestPolylabelWater1(t *testing.T) {
	polygon := loadData("test_data/water1.json")
	var x, y float64

	x, y = Polylabel(polygon, 1.0)
	AssertEqual(t, x, 3865.85009765625)
	AssertEqual(t, y, 2124.87841796875)

	x, y = Polylabel(polygon, 50.0)
	AssertEqual(t, x, 3854.296875)
	AssertEqual(t, y, 2123.828125)
}

func TestPolylabelWater2(t *testing.T) {
	polygon := loadData("test_data/water2.json")

	x, y := Polylabel(polygon, 1.0)
	AssertEqual(t, x, 3263.5)
	AssertEqual(t, y, 3263.5)
}

func TestDegeneratePolygons(t *testing.T) {
	var x, y float64

	{
		polygon := geometry.NewPoly([]geometry.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 0, Y: 0}}, nil, nil)
		x, y = Polylabel(polygon, 1.0)
		AssertEqual(t, x, 0.0)
		AssertEqual(t, y, 0.0)
	}

	{
		polygon := geometry.NewPoly([]geometry.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0}}, nil, nil)
		x, y = Polylabel(polygon, 1.0)
		AssertEqual(t, x, 0.0)
		AssertEqual(t, y, 0.0)
	}
}
