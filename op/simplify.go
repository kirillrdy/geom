package geomop

import (
	"github.com/ctessum/geom"
)

func Simplify(g geom.T, tolerance float64) (geom.T, error) {

	switch g.(type) {
	case geom.Polygon:
		p := g.(geom.Polygon)
		var out geom.Polygon = make([][]geom.Point, len(p))
		for i, r := range p {
			out[i] = simplifyCurve(r, p, tolerance)
		}
		return out, nil
	case geom.MultiPolygon:
		mp := g.(geom.MultiPolygon)
		var out geom.MultiPolygon = make([]geom.Polygon, len(mp))
		for i, p := range mp {
			o, _ := Simplify(p, tolerance)
			out[i] = o.(geom.Polygon)
		}
		return out, nil
	case geom.LineString:
		l := g.(geom.LineString)
		out := geom.LineString(simplifyCurve(l, [][]geom.Point{}, tolerance))
		return out, nil
	case geom.MultiLineString:
		ml := g.(geom.MultiLineString)
		var out geom.MultiLineString = make([]geom.LineString, len(ml))
		for i, l := range ml {
			o, _ := Simplify(l, tolerance)
			out[i] = o.(geom.LineString)
		}
		return out, nil
	default:
		return nil, newUnsupportedGeometryError(g)
	}
}

func simplifyCurve(curve []geom.Point,
	otherCurves [][]geom.Point, tol float64) []geom.Point {
	out := make([]geom.Point, 0, len(curve))

	i := 0
	out = append(out, curve[i])
	for {
		breakTime := false
		for j := i + 2; j < len(curve); j++ {
			breakTime2 := false
			for k := i + 1; k < j; k++ {
				d := distPointToSegment(curve[k], curve[i], curve[j])
				if d > tol {
					// we have found a candidate point to keep
					for {
						// Make sure this simplifcation doesn't cause any self
						// intersections.
						if segMakesNotSimple(curve[i], curve[j-1],
							[][]geom.Point{out[0 : len(out)-1]}) ||
							segMakesNotSimple(curve[i], curve[j-1],
								[][]geom.Point{curve[j:]}) ||
							segMakesNotSimple(curve[i], curve[j-1],
								otherCurves) {
							j--
						} else {
							i = j - 1
							out = append(out, curve[i])
							breakTime2 = true
							break
						}
					}
				}
				if breakTime2 {
					break
				}
			}
			if j == len(curve)-1 {
				out = append(out, curve[j])
				breakTime = true
			}
		}
		if breakTime {
			break
		}
	}
	return out
}

func segMakesNotSimple(segStart, segEnd geom.Point, paths [][]geom.Point) bool {
	seg1 := segment{segStart, segEnd}
	for _, p := range paths {
		for i := 0; i < len(p)-1; i++ {
			seg2 := segment{p[i], p[i+1]}
			numIntersections, _, _ := findIntersection(seg1, seg2)
			if numIntersections > 0 {
				return true
			}
		}
	}
	return false
}