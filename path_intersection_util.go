package canvas

import (
	"fmt"
	"math"
)

// see https://github.com/signavio/svg-intersections
// see https://github.com/w8r/bezier-intersect
// see https://cs.nyu.edu/exact/doc/subdiv1.pdf

// intersect for path segments a and b, starting at a0 and b0. Note that all intersection functions
// return upto two intersections.
func intersectionSegment(zs Intersections, a0 Point, a []float64, b0 Point, b []float64) Intersections {
	n := len(zs)
	swapCurves := false
	if a[0] == LineToCmd || a[0] == CloseCmd {
		if b[0] == LineToCmd || b[0] == CloseCmd {
			zs = intersectionLineLine(zs, a0, Point{a[1], a[2]}, b0, Point{b[1], b[2]})
		} else if b[0] == QuadToCmd {
			zs = intersectionLineQuad(zs, a0, Point{a[1], a[2]}, b0, Point{b[1], b[2]}, Point{b[3], b[4]})
		} else if b[0] == CubeToCmd {
			zs = intersectionLineCube(zs, a0, Point{a[1], a[2]}, b0, Point{b[1], b[2]}, Point{b[3], b[4]}, Point{b[5], b[6]})
		} else if b[0] == ArcToCmd {
			rx := b[1]
			ry := b[2]
			phi := b[3] * math.Pi / 180.0
			large, sweep := toArcFlags(b[4])
			cx, cy, theta0, theta1 := ellipseToCenter(b0.X, b0.Y, rx, ry, phi, large, sweep, b[5], b[6])
			zs = intersectionLineEllipse(zs, a0, Point{a[1], a[2]}, Point{cx, cy}, Point{rx, ry}, phi, theta0, theta1)
		}
	} else if a[0] == QuadToCmd {
		if b[0] == LineToCmd || b[0] == CloseCmd {
			zs = intersectionLineQuad(zs, b0, Point{b[1], b[2]}, a0, Point{a[1], a[2]}, Point{a[3], a[4]})
			swapCurves = true
		} else if b[0] == QuadToCmd {
			panic("unsupported intersection for quad-quad")
		} else if b[0] == CubeToCmd {
			panic("unsupported intersection for quad-cube")
		} else if b[0] == ArcToCmd {
			panic("unsupported intersection for quad-arc")
		}
	} else if a[0] == CubeToCmd {
		if b[0] == LineToCmd || b[0] == CloseCmd {
			zs = intersectionLineCube(zs, b0, Point{b[1], b[2]}, a0, Point{a[1], a[2]}, Point{a[3], a[4]}, Point{a[5], a[6]})
			swapCurves = true
		} else if b[0] == QuadToCmd {
			panic("unsupported intersection for cube-quad")
		} else if b[0] == CubeToCmd {
			panic("unsupported intersection for cube-cube")
		} else if b[0] == ArcToCmd {
			panic("unsupported intersection for cube-arc")
		}
	} else if a[0] == ArcToCmd {
		rx := a[1]
		ry := a[2]
		phi := a[3] * math.Pi / 180.0
		large, sweep := toArcFlags(a[4])
		cx, cy, theta0, theta1 := ellipseToCenter(a0.X, a0.Y, rx, ry, phi, large, sweep, a[5], a[6])
		if b[0] == LineToCmd || b[0] == CloseCmd {
			zs = intersectionLineEllipse(zs, b0, Point{b[1], b[2]}, Point{cx, cy}, Point{rx, ry}, phi, theta0, theta1)
			swapCurves = true
		} else if b[0] == QuadToCmd {
			panic("unsupported intersection for arc-quad")
		} else if b[0] == CubeToCmd {
			panic("unsupported intersection for arc-cube")
		} else if b[0] == ArcToCmd {
			rx2 := b[1]
			ry2 := b[2]
			phi2 := b[3] * math.Pi / 180.0
			large2, sweep2 := toArcFlags(b[4])
			cx2, cy2, theta20, theta21 := ellipseToCenter(b0.X, b0.Y, rx2, ry2, phi2, large2, sweep2, b[5], b[6])
			zs = intersectionEllipseEllipse(zs, Point{cx, cy}, Point{rx, ry}, phi, theta0, theta1, Point{cx2, cy2}, Point{rx2, ry2}, phi2, theta20, theta21)
		}
	}

	// swap A and B in the intersection found to match segments A and B of this function
	if swapCurves {
		for i := n; i < len(zs); i++ {
			zs[i].T[0], zs[i].T[1] = zs[i].T[1], zs[i].T[0]
			zs[i].Dir[0], zs[i].Dir[1] = zs[i].Dir[1], zs[i].Dir[0]
		}
	}
	return zs
}

// Intersection is an intersection between two path segments, e.g. Line x Line. Note that an
// intersection is tangent also when it is at one of the endpoints, in which case it may be tangent
// for this segment but may or may not cross the path depending on the adjacent segment.
// Notabene: for quad/cube/ellipse aligned angles at the endpoint for non-overlapping curves are deviated slightly to correctly calculate the value for Into, and will thus not be aligned
type Intersection struct {
	Point              // coordinate of intersection
	T       [2]float64 // position along segment [0,1]
	Dir     [2]float64 // direction at intersection [0,2*pi)
	Tangent bool       // intersection is tangent (touches) instead of secant (crosses)
	Same    bool       // intersection is of two overlapping segments (tangent is also true)
}

// Into returns true if first path goes into the left-hand side of the second path,
// i.e. the second path goes to the right-hand side of the first path.
func (z Intersection) Into() bool {
	return angleBetweenExclusive(z.Dir[1]-z.Dir[0], math.Pi, 2.0*math.Pi)
}

func (z Intersection) Equals(o Intersection) bool {
	return z.Point.Equals(o.Point) && Equal(z.T[0], o.T[0]) && Equal(z.T[1], o.T[1]) && angleEqual(z.Dir[0], o.Dir[0]) && angleEqual(z.Dir[1], o.Dir[1]) && z.Tangent == o.Tangent && z.Same == o.Same
}

func (z Intersection) String() string {
	var extra string
	if z.Tangent {
		extra = " Tangent"
	}
	if z.Same {
		extra = " Same"
	}
	return fmt.Sprintf("({%v,%v} t={%v,%v} dir={%v°,%v°}%v)", numEps(z.Point.X), numEps(z.Point.Y), numEps(z.T[0]), numEps(z.T[1]), numEps(angleNorm(z.Dir[0])*180.0/math.Pi), numEps(angleNorm(z.Dir[1])*180.0/math.Pi), extra)
}

type Intersections []Intersection

// Has returns true if there are secant/tangent intersections.
func (zs Intersections) Has() bool {
	return 0 < len(zs)
}

// HasSecant returns true when there are secant intersections, i.e. the curves intersect and cross (they cut).
func (zs Intersections) HasSecant() bool {
	for _, z := range zs {
		if !z.Tangent {
			return true
		}
	}
	return false
}

// HasTangent returns true when there are tangent intersections, i.e. the curves intersect but don't cross (they touch).
func (zs Intersections) HasTangent() bool {
	for _, z := range zs {
		if z.Tangent {
			return true
		}
	}
	return false
}

func (zs Intersections) add(pos Point, ta, tb, dira, dirb float64, tangent, same bool) Intersections {
	// normalise T values between [0,1]
	if ta < 0.0 { // || Equal(ta, 0.0) {
		ta = 0.0
	} else if 1.0 <= ta { // || Equal(ta, 1.0) {
		ta = 1.0
	}
	if tb < 0.0 { // || Equal(tb, 0.0) {
		tb = 0.0
	} else if 1.0 < tb { // || Equal(tb, 1.0) {
		tb = 1.0
	}
	return append(zs, Intersection{pos, [2]float64{ta, tb}, [2]float64{dira, dirb}, tangent, same})
}

func correctIntersection(z, aMin, aMax, bMin, bMax Point) Point {
	if z.X < aMin.X {
		//fmt.Println("CORRECT 1:", a0, a1, "--", b0, b1)
		z.X = aMin.X
	} else if aMax.X < z.X {
		//fmt.Println("CORRECT 2:", a0, a1, "--", b0, b1)
		z.X = aMax.X
	}
	if z.X < bMin.X {
		//fmt.Println("CORRECT 3:", a0, a1, "--", b0, b1)
		z.X = bMin.X
	} else if bMax.X < z.X {
		//fmt.Println("CORRECT 4:", a0, a1, "--", b0, b1)
		z.X = bMax.X
	}
	if z.Y < aMin.Y {
		//fmt.Println("CORRECT 5:", a0, a1, "--", b0, b1)
		z.Y = aMin.Y
	} else if aMax.Y < z.Y {
		//fmt.Println("CORRECT 6:", a0, a1, "--", b0, b1)
		z.Y = aMax.Y
	}
	if z.Y < bMin.Y {
		//fmt.Println("CORRECT 7:", a0, a1, "--", b0, b1)
		z.Y = bMin.Y
	} else if bMax.Y < z.Y {
		//fmt.Println("CORRECT 8:", a0, a1, "--", b0, b1)
		z.Y = bMax.Y
	}
	return z
}

// F. Antonio, "Faster Line Segment Intersection", Graphics Gems III, 1992
func intersectionLineLineBentleyOttmann(zs []Point, a0, a1, b0, b1 Point) []Point {
	// fast line-line intersection code, with additional constraints for the BentleyOttmann code:
	// - a0 is to the left and/or bottom of a1, same for b0 and b1
	// - an intersection z must keep the above property between (a0,z), (z,a1), (b0,z), and (z,b1)
	// note that an exception is made for (z,a1) and (z,b1) to allow them to become vertical, this
	// is because there isn't always "space" between a0.X and a1.X, eg. when a1.X = nextafter(a0.X)
	if a1.X < b0.X || b1.X < a0.X {
		return zs
	}

	aMin, aMax, bMin, bMax := a0, a1, b0, b1
	if a1.Y < a0.Y {
		aMin.Y, aMax.Y = aMax.Y, aMin.Y
	}
	if b1.Y < b0.Y {
		bMin.Y, bMax.Y = bMax.Y, bMin.Y
	}
	if aMax.Y < bMin.Y || bMax.Y < aMin.Y {
		return zs
	} else if (aMax.X == bMin.X || bMax.X == aMin.X) && (aMax.Y == bMin.Y || bMax.Y == aMin.Y) {
		return zs
	}

	// only the position and T values are valid for each intersection
	A := a1.Sub(a0)
	B := b0.Sub(b1)
	C := a0.Sub(b0)
	denom := B.PerpDot(A)
	// divide by length^2 since the perpdot between very small segments may be below Epsilon
	if denom == 0.0 {
		// colinear
		if C.PerpDot(B) == 0.0 {
			// overlap, rotate to x-axis
			a, b, c, d := a0.X, a1.X, b0.X, b1.X
			if math.Abs(A.X) < math.Abs(A.Y) {
				// mostly vertical
				a, b, c, d = a0.Y, a1.Y, b0.Y, b1.Y
			}
			if c < b && a < d {
				if a < c {
					zs = append(zs, b0)
				} else if c < a {
					zs = append(zs, a0)
				}
				if d < b {
					zs = append(zs, b1)
				} else if b < d {
					zs = append(zs, a1)
				}
			}
		}
		return zs
	}

	// find intersections within +-Epsilon to avoid missing near intersections
	ta := C.PerpDot(B) / denom
	if ta < -Epsilon || 1.0+Epsilon < ta {
		return zs
	}

	tb := A.PerpDot(C) / denom
	if tb < -Epsilon || 1.0+Epsilon < tb {
		return zs
	}

	// ta is snapped to 0.0 or 1.0 if very close
	if ta <= Epsilon {
		ta = 0.0
	} else if 1.0-Epsilon <= ta {
		ta = 1.0
	}

	z := a0.Interpolate(a1, ta)
	z = correctIntersection(z, aMin, aMax, bMin, bMax)
	if z != a0 && z != a1 || z != b0 && z != b1 {
		// not at endpoints for both
		if a0 != b0 && z != a0 && z != b0 && b0.Sub(z).PerpDot(z.Sub(a0)) == 0.0 {
			a, c, m := a0.X, b0.X, z.X
			if math.Abs(z.Sub(a0).X) < math.Abs(z.Sub(a0).Y) {
				// mostly vertical
				a, c, m = a0.Y, b0.Y, z.Y
			}

			if a != c && (a < m) == (c < m) {
				if a < m && a < c || m < a && c < a {
					zs = append(zs, b0)
				} else {
					zs = append(zs, a0)
				}
			}
			zs = append(zs, z)
		} else if a1 != b1 && z != a1 && z != b1 && z.Sub(b1).PerpDot(a1.Sub(z)) == 0.0 {
			b, d, m := a1.X, b1.X, z.X
			if math.Abs(z.Sub(a1).X) < math.Abs(z.Sub(a1).Y) {
				// mostly vertical
				b, d, m = a1.Y, b1.Y, z.Y
			}

			if b != d && (b < m) == (d < m) {
				if b < m && b < d || m < b && d < b {
					zs = append(zs, b1)
				} else {
					zs = append(zs, a1)
				}
			}
		} else {
			zs = append(zs, z)
		}
	}
	return zs
}

func intersectionLineLine(zs Intersections, a0, a1, b0, b1 Point) Intersections {
	if a0.Equals(a1) || b0.Equals(b1) {
		return zs // zero-length Close
	}

	da := a1.Sub(a0)
	db := b1.Sub(b0)
	anglea := da.Angle()
	angleb := db.Angle()
	div := da.PerpDot(db)

	// divide by length^2 since otherwise the perpdot between very small segments may be
	// below Epsilon
	if length := da.Length() * db.Length(); Equal(div/length, 0.0) {
		// parallel
		if Equal(b0.Sub(a0).PerpDot(db), 0.0) {
			// overlap, rotate to x-axis
			a := a0.Rot(-anglea, Point{}).X
			b := a1.Rot(-anglea, Point{}).X
			c := b0.Rot(-anglea, Point{}).X
			d := b1.Rot(-anglea, Point{}).X
			if Interval(a, c, d) && Interval(b, c, d) {
				// a-b in c-d or a-b == c-d
				zs = zs.add(a0, 0.0, (a-c)/(d-c), anglea, angleb, true, true)
				zs = zs.add(a1, 1.0, (b-c)/(d-c), anglea, angleb, true, true)
			} else if Interval(c, a, b) && Interval(d, a, b) {
				// c-d in a-b
				zs = zs.add(b0, (c-a)/(b-a), 0.0, anglea, angleb, true, true)
				zs = zs.add(b1, (d-a)/(b-a), 1.0, anglea, angleb, true, true)
			} else if Interval(a, c, d) {
				// a in c-d
				same := a < d-Epsilon || a < c-Epsilon
				zs = zs.add(a0, 0.0, (a-c)/(d-c), anglea, angleb, true, same)
				if a < d-Epsilon {
					zs = zs.add(b1, (d-a)/(b-a), 1.0, anglea, angleb, true, true)
				} else if a < c-Epsilon {
					zs = zs.add(b0, (c-a)/(b-a), 0.0, anglea, angleb, true, true)
				}
			} else if Interval(b, c, d) {
				// b in c-d
				same := c < b-Epsilon || d < b-Epsilon
				if c < b-Epsilon {
					zs = zs.add(b0, (c-a)/(b-a), 0.0, anglea, angleb, true, true)
				} else if d < b-Epsilon {
					zs = zs.add(b1, (d-a)/(b-a), 1.0, anglea, angleb, true, true)
				}
				zs = zs.add(a1, 1.0, (b-c)/(d-c), anglea, angleb, true, same)
			}
		}
		return zs
	} else if a1.Equals(b0) {
		// handle common cases with endpoints to avoid numerical issues
		zs = zs.add(a1, 1.0, 0.0, anglea, angleb, true, false)
		return zs
	} else if a0.Equals(b1) {
		// handle common cases with endpoints to avoid numerical issues
		zs = zs.add(a0, 0.0, 1.0, anglea, angleb, true, false)
		return zs
	}

	ta := db.PerpDot(a0.Sub(b0)) / div
	tb := da.PerpDot(a0.Sub(b0)) / div
	if Interval(ta, 0.0, 1.0) && Interval(tb, 0.0, 1.0) {
		tangent := Equal(ta, 0.0) || Equal(ta, 1.0) || Equal(tb, 0.0) || Equal(tb, 1.0)
		zs = zs.add(a0.Interpolate(a1, ta), ta, tb, anglea, angleb, tangent, false)
	}
	return zs
}

// https://www.particleincell.com/2013/cubic-line-intersection/
func intersectionLineQuad(zs Intersections, l0, l1, p0, p1, p2 Point) Intersections {
	if l0.Equals(l1) {
		return zs // zero-length Close
	}

	// write line as A.X = bias
	A := Point{l1.Y - l0.Y, l0.X - l1.X}
	bias := l0.Dot(A)

	a := A.Dot(p0.Sub(p1.Mul(2.0)).Add(p2))
	b := A.Dot(p1.Sub(p0).Mul(2.0))
	c := A.Dot(p0) - bias

	roots := []float64{}
	r0, r1 := solveQuadraticFormula(a, b, c)
	if !math.IsNaN(r0) {
		roots = append(roots, r0)
		if !math.IsNaN(r1) {
			roots = append(roots, r1)
		}
	}

	dira := l1.Sub(l0).Angle()
	horizontal := math.Abs(l1.Y-l0.Y) <= math.Abs(l1.X-l0.X)
	for _, root := range roots {
		if Interval(root, 0.0, 1.0) {
			var s float64
			pos := quadraticBezierPos(p0, p1, p2, root)
			if horizontal {
				s = (pos.X - l0.X) / (l1.X - l0.X)
			} else {
				s = (pos.Y - l0.Y) / (l1.Y - l0.Y)
			}
			if Interval(s, 0.0, 1.0) {
				deriv := quadraticBezierDeriv(p0, p1, p2, root)
				dirb := deriv.Angle()
				endpoint := Equal(root, 0.0) || Equal(root, 1.0) || Equal(s, 0.0) || Equal(s, 1.0)
				if endpoint {
					// deviate angle slightly at endpoint when aligned to properly set Into
					deriv2 := quadraticBezierDeriv2(p0, p1, p2)
					if (0.0 <= deriv.PerpDot(deriv2)) == (Equal(root, 0.0) || !Equal(root, 1.0) && Equal(s, 0.0)) {
						dirb += Epsilon * 2.0 // t=0 and CCW, or t=1 and CW
					} else {
						dirb -= Epsilon * 2.0 // t=0 and CW, or t=1 and CCW
					}
					dirb = angleNorm(dirb)
				}
				zs = zs.add(pos, s, root, dira, dirb, endpoint || Equal(A.Dot(deriv), 0.0), false)
			}
		}
	}
	return zs
}

// https://www.particleincell.com/2013/cubic-line-intersection/
func intersectionLineCube(zs Intersections, l0, l1, p0, p1, p2, p3 Point) Intersections {
	if l0.Equals(l1) {
		return zs // zero-length Close
	}

	// write line as A.X = bias
	A := Point{l1.Y - l0.Y, l0.X - l1.X}
	bias := l0.Dot(A)

	a := A.Dot(p3.Sub(p0).Add(p1.Mul(3.0)).Sub(p2.Mul(3.0)))
	b := A.Dot(p0.Mul(3.0).Sub(p1.Mul(6.0)).Add(p2.Mul(3.0)))
	c := A.Dot(p1.Mul(3.0).Sub(p0.Mul(3.0)))
	d := A.Dot(p0) - bias

	roots := []float64{}
	r0, r1, r2 := solveCubicFormula(a, b, c, d)
	if !math.IsNaN(r0) {
		roots = append(roots, r0)
		if !math.IsNaN(r1) {
			roots = append(roots, r1)
			if !math.IsNaN(r2) {
				roots = append(roots, r2)
			}
		}
	}

	dira := l1.Sub(l0).Angle()
	horizontal := math.Abs(l1.Y-l0.Y) <= math.Abs(l1.X-l0.X)
	for _, root := range roots {
		if Interval(root, 0.0, 1.0) {
			var s float64
			pos := cubicBezierPos(p0, p1, p2, p3, root)
			if horizontal {
				s = (pos.X - l0.X) / (l1.X - l0.X)
			} else {
				s = (pos.Y - l0.Y) / (l1.Y - l0.Y)
			}
			if Interval(s, 0.0, 1.0) {
				deriv := cubicBezierDeriv(p0, p1, p2, p3, root)
				dirb := deriv.Angle()
				tangent := Equal(A.Dot(deriv), 0.0)
				endpoint := Equal(root, 0.0) || Equal(root, 1.0) || Equal(s, 0.0) || Equal(s, 1.0)
				if endpoint {
					// deviate angle slightly at endpoint when aligned to properly set Into
					deriv2 := cubicBezierDeriv2(p0, p1, p2, p3, root)
					if (0.0 <= deriv.PerpDot(deriv2)) == (Equal(root, 0.0) || !Equal(root, 1.0) && Equal(s, 0.0)) {
						dirb += Epsilon * 2.0 // t=0 and CCW, or t=1 and CW
					} else {
						dirb -= Epsilon * 2.0 // t=0 and CW, or t=1 and CCW
					}
				} else if angleEqual(dira, dirb) || angleEqual(dira, dirb+math.Pi) {
					// directions are parallel but the paths do cross (inflection point)
					// TODO: test better
					deriv2 := cubicBezierDeriv2(p0, p1, p2, p3, root)
					if Equal(deriv2.X, 0.0) && Equal(deriv2.Y, 0.0) {
						deriv3 := cubicBezierDeriv3(p0, p1, p2, p3, root)
						if 0.0 < deriv.PerpDot(deriv3) {
							dirb += Epsilon * 2.0
						} else {
							dirb -= Epsilon * 2.0
						}
						dirb = angleNorm(dirb)
						tangent = false
					}
				}
				zs = zs.add(pos, s, root, dira, dirb, endpoint || tangent, false)
			}
		}
	}
	return zs
}

// handle line-arc intersections and their peculiarities regarding angles
func addLineArcIntersection(zs Intersections, pos Point, dira, dirb, t, t0, t1, angle, theta0, theta1 float64, tangent bool) Intersections {
	if theta0 <= theta1 {
		angle = theta0 - Epsilon + angleNorm(angle-theta0+Epsilon)
	} else {
		angle = theta1 - Epsilon + angleNorm(angle-theta1+Epsilon)
	}
	endpoint := Equal(t, t0) || Equal(t, t1) || Equal(angle, theta0) || Equal(angle, theta1)
	if endpoint {
		// deviate angle slightly at endpoint when aligned to properly set Into
		if (theta0 <= theta1) == (Equal(angle, theta0) || !Equal(angle, theta1) && Equal(t, t0)) {
			dirb += Epsilon * 2.0 // t=0 and CCW, or t=1 and CW
		} else {
			dirb -= Epsilon * 2.0 // t=0 and CW, or t=1 and CCW
		}
		dirb = angleNorm(dirb)
	}

	// snap segment parameters to 0.0 and 1.0 to avoid numerical issues
	var s float64
	if Equal(t, t0) {
		t = 0.0
	} else if Equal(t, t1) {
		t = 1.0
	} else {
		t = (t - t0) / (t1 - t0)
	}
	if Equal(angle, theta0) {
		s = 0.0
	} else if Equal(angle, theta1) {
		s = 1.0
	} else {
		s = (angle - theta0) / (theta1 - theta0)
	}
	return zs.add(pos, t, s, dira, dirb, endpoint || tangent, false)
}

// https://www.geometrictools.com/GTE/Mathematics/IntrLine2Circle2.h
func intersectionLineCircle(zs Intersections, l0, l1, center Point, radius, theta0, theta1 float64) Intersections {
	if l0.Equals(l1) {
		return zs // zero-length Close
	}

	// solve l0 + t*(l1-l0) = P + t*D = X  (line equation)
	// and |X - center| = |X - C| = R = radius  (circle equation)
	// by substitution and squaring: |P + t*D - C|^2 = R^2
	// giving: D^2 t^2 + 2D(P-C) t + (P-C)^2-R^2 = 0
	dir := l1.Sub(l0)
	diff := l0.Sub(center) // P-C
	length := dir.Length()
	D := dir.Div(length)

	// we normalise D to be of length 1, so that the roots are in [0,length]
	a := 1.0
	b := 2.0 * D.Dot(diff)
	c := diff.Dot(diff) - radius*radius

	// find solutions for t ∈ [0,1], the parameter along the line's path
	roots := []float64{}
	r0, r1 := solveQuadraticFormula(a, b, c)
	if !math.IsNaN(r0) {
		roots = append(roots, r0)
		if !math.IsNaN(r1) && !Equal(r0, r1) {
			roots = append(roots, r1)
		}
	}

	// handle common cases with endpoints to avoid numerical issues
	// snap closest root to path's start or end
	if 0 < len(roots) {
		if pos := l0.Sub(center); Equal(pos.Length(), radius) {
			if len(roots) == 1 || math.Abs(roots[0]) < math.Abs(roots[1]) {
				roots[0] = 0.0
			} else {
				roots[1] = 0.0
			}
		}
		if pos := l1.Sub(center); Equal(pos.Length(), radius) {
			if len(roots) == 1 || math.Abs(roots[0]-length) < math.Abs(roots[1]-length) {
				roots[0] = length
			} else {
				roots[1] = length
			}
		}
	}

	// add intersections
	dira := dir.Angle()
	tangent := len(roots) == 1
	for _, root := range roots {
		pos := diff.Add(dir.Mul(root / length))
		angle := math.Atan2(pos.Y*radius, pos.X*radius)
		if Interval(root, 0.0, length) && angleBetween(angle, theta0, theta1) {
			pos = center.Add(pos)
			dirb := ellipseDeriv(radius, radius, 0.0, theta0 <= theta1, angle).Angle()
			zs = addLineArcIntersection(zs, pos, dira, dirb, root, 0.0, length, angle, theta0, theta1, tangent)
		}
	}
	return zs
}

func intersectionLineEllipse(zs Intersections, l0, l1, center, radius Point, phi, theta0, theta1 float64) Intersections {
	if Equal(radius.X, radius.Y) {
		return intersectionLineCircle(zs, l0, l1, center, radius.X, theta0, theta1)
	} else if l0.Equals(l1) {
		return zs // zero-length Close
	}

	// TODO: needs more testing
	// TODO: intersection inconsistency due to numerical stability in finding tangent collisions for subsequent paht segments (line -> ellipse), or due to the endpoint of a line not touching with another arc, but the subsequent segment does touch with its starting point
	dira := l1.Sub(l0).Angle()

	// we take the ellipse center as the origin and counter-rotate by phi
	l0 = l0.Sub(center).Rot(-phi, Origin)
	l1 = l1.Sub(center).Rot(-phi, Origin)

	// line: cx + dy + e = 0
	c := l0.Y - l1.Y
	d := l1.X - l0.X
	e := l0.PerpDot(l1)

	// follow different code paths when line is mostly horizontal or vertical
	horizontal := math.Abs(c) <= math.Abs(d)

	// ellipse: x^2/a + y^2/b = 1
	a := radius.X * radius.X
	b := radius.Y * radius.Y

	// rewrite as a polynomial by substituting x or y to obtain:
	// At^2 + Bt + C = 0, with t either x (horizontal) or y (!horizontal)
	var A, B, C float64
	A = a*c*c + b*d*d
	if horizontal {
		B = 2.0 * a * c * e
		C = a*e*e - a*b*d*d
	} else {
		B = 2.0 * b * d * e
		C = b*e*e - a*b*c*c
	}

	// find solutions
	roots := []float64{}
	r0, r1 := solveQuadraticFormula(A, B, C)
	if !math.IsNaN(r0) {
		roots = append(roots, r0)
		if !math.IsNaN(r1) && !Equal(r0, r1) {
			roots = append(roots, r1)
		}
	}

	for _, root := range roots {
		// get intersection position with center as origin
		var x, y, t0, t1 float64
		if horizontal {
			x = root
			y = -e/d - c*root/d
			t0 = l0.X
			t1 = l1.X
		} else {
			x = -e/c - d*root/c
			y = root
			t0 = l0.Y
			t1 = l1.Y
		}

		tangent := Equal(root, 0.0)
		angle := math.Atan2(y*radius.X, x*radius.Y)
		if Interval(root, t0, t1) && angleBetween(angle, theta0, theta1) {
			pos := Point{x, y}.Rot(phi, Origin).Add(center)
			dirb := ellipseDeriv(radius.X, radius.Y, phi, theta0 <= theta1, angle).Angle()
			zs = addLineArcIntersection(zs, pos, dira, dirb, root, t0, t1, angle, theta0, theta1, tangent)
		}
	}
	return zs
}

func intersectionEllipseEllipse(zs Intersections, c0, r0 Point, phi0, thetaStart0, thetaEnd0 float64, c1, r1 Point, phi1, thetaStart1, thetaEnd1 float64) Intersections {
	// TODO: needs more testing
	if !Equal(r0.X, r0.Y) || !Equal(r1.X, r1.Y) {
		panic("not handled") // ellipses
	}

	arcAngle := func(theta float64, sweep bool) float64 {
		theta += math.Pi / 2.0
		if !sweep {
			theta -= math.Pi
		}
		return angleNorm(theta)
	}

	dtheta0 := thetaEnd0 - thetaStart0
	thetaStart0 = angleNorm(thetaStart0 + phi0)
	thetaEnd0 = thetaStart0 + dtheta0

	dtheta1 := thetaEnd1 - thetaStart1
	thetaStart1 = angleNorm(thetaStart1 + phi1)
	thetaEnd1 = thetaStart1 + dtheta1

	if c0.Equals(c1) && r0.Equals(r1) {
		// parallel
		tOffset1 := 0.0
		dirOffset1 := 0.0
		if (0.0 <= dtheta0) != (0.0 <= dtheta1) {
			thetaStart1, thetaEnd1 = thetaEnd1, thetaStart1 // keep order on first arc
			dirOffset1 = math.Pi
			tOffset1 = 1.0
		}

		// will add either 1 (when touching) or 2 (when overlapping) intersections
		if t := angleTime(thetaStart0, thetaStart1, thetaEnd1); Interval(t, 0.0, 1.0) {
			// ellipse0 starts within/on border of ellipse1
			dir := arcAngle(thetaStart0, 0.0 <= dtheta0)
			pos := EllipsePos(r0.X, r0.Y, 0.0, c0.X, c0.Y, thetaStart0)
			zs = zs.add(pos, 0.0, math.Abs(t-tOffset1), dir, angleNorm(dir+dirOffset1), true, true)
		}
		if t := angleTime(thetaStart1, thetaStart0, thetaEnd0); IntervalExclusive(t, 0.0, 1.0) {
			// ellipse1 starts within ellipse0
			dir := arcAngle(thetaStart1, 0.0 <= dtheta0)
			pos := EllipsePos(r0.X, r0.Y, 0.0, c0.X, c0.Y, thetaStart1)
			zs = zs.add(pos, t, tOffset1, dir, angleNorm(dir+dirOffset1), true, true)
		}
		if t := angleTime(thetaEnd1, thetaStart0, thetaEnd0); IntervalExclusive(t, 0.0, 1.0) {
			// ellipse1 ends within ellipse0
			dir := arcAngle(thetaEnd1, 0.0 <= dtheta0)
			pos := EllipsePos(r0.X, r0.Y, 0.0, c0.X, c0.Y, thetaEnd1)
			zs = zs.add(pos, t, 1.0-tOffset1, dir, angleNorm(dir+dirOffset1), true, true)
		}
		if t := angleTime(thetaEnd0, thetaStart1, thetaEnd1); Interval(t, 0.0, 1.0) {
			// ellipse0 ends within/on border of ellipse1
			dir := arcAngle(thetaEnd0, 0.0 <= dtheta0)
			pos := EllipsePos(r0.X, r0.Y, 0.0, c0.X, c0.Y, thetaEnd0)
			zs = zs.add(pos, 1.0, math.Abs(t-tOffset1), dir, angleNorm(dir+dirOffset1), true, true)
		}
		return zs
	}

	// https://math.stackexchange.com/questions/256100/how-can-i-find-the-points-at-which-two-circles-intersect
	// https://gist.github.com/jupdike/bfe5eb23d1c395d8a0a1a4ddd94882ac
	R := c0.Sub(c1).Length()
	if R < math.Abs(r0.X-r1.X) || r0.X+r1.X < R {
		return zs
	}
	R2 := R * R

	k := r0.X*r0.X - r1.X*r1.X
	a := 0.5
	b := 0.5 * k / R2
	c := 0.5 * math.Sqrt(2.0*(r0.X*r0.X+r1.X*r1.X)/R2-k*k/(R2*R2)-1.0)

	mid := c1.Sub(c0).Mul(a + b)
	dev := Point{c1.Y - c0.Y, c0.X - c1.X}.Mul(c)

	tangent := dev.Equals(Point{})
	anglea0 := mid.Add(dev).Angle()
	anglea1 := c0.Sub(c1).Add(mid).Add(dev).Angle()
	ta0 := angleTime(anglea0, thetaStart0, thetaEnd0)
	ta1 := angleTime(anglea1, thetaStart1, thetaEnd1)
	if Interval(ta0, 0.0, 1.0) && Interval(ta1, 0.0, 1.0) {
		dir0 := arcAngle(anglea0, 0.0 <= dtheta0)
		dir1 := arcAngle(anglea1, 0.0 <= dtheta1)
		endpoint := Equal(ta0, 0.0) || Equal(ta0, 1.0) || Equal(ta1, 0.0) || Equal(ta1, 1.0)
		zs = zs.add(c0.Add(mid).Add(dev), ta0, ta1, dir0, dir1, tangent || endpoint, false)
	}

	if !tangent {
		angleb0 := mid.Sub(dev).Angle()
		angleb1 := c0.Sub(c1).Add(mid).Sub(dev).Angle()
		tb0 := angleTime(angleb0, thetaStart0, thetaEnd0)
		tb1 := angleTime(angleb1, thetaStart1, thetaEnd1)
		if Interval(tb0, 0.0, 1.0) && Interval(tb1, 0.0, 1.0) {
			dir0 := arcAngle(angleb0, 0.0 <= dtheta0)
			dir1 := arcAngle(angleb1, 0.0 <= dtheta1)
			endpoint := Equal(tb0, 0.0) || Equal(tb0, 1.0) || Equal(tb1, 0.0) || Equal(tb1, 1.0)
			zs = zs.add(c0.Add(mid).Sub(dev), tb0, tb1, dir0, dir1, endpoint, false)
		}
	}
	return zs
}

// TODO: bezier-bezier intersection
// TODO: bezier-ellipse intersection

// For Bézier-Bézier intersections:
// see T.W. Sederberg, "Computer Aided Geometric Design", 2012
// see T.W. Sederberg and T. Nishita, "Curve intersection using Bézier clipping", 1990
// see T.W. Sederberg and S.R. Parry, "Comparison of three curve intersection algorithms", 1986

func intersectionRayLine(a0, a1, b0, b1 Point) (Point, bool) {
	da := a1.Sub(a0)
	db := b1.Sub(b0)
	div := da.PerpDot(db)
	if Equal(div, 0.0) {
		// parallel
		return Point{}, false
	}

	tb := da.PerpDot(a0.Sub(b0)) / div
	if Interval(tb, 0.0, 1.0) {
		return b0.Interpolate(b1, tb), true
	}
	return Point{}, false
}

// https://mathworld.wolfram.com/Circle-LineIntersection.html
func intersectionRayCircle(l0, l1, c Point, r float64) (Point, Point, bool) {
	d := l1.Sub(l0).Norm(1.0) // along line direction, anchored in l0, its length is 1
	D := l0.Sub(c).PerpDot(d)
	discriminant := r*r - D*D
	if discriminant < 0 {
		return Point{}, Point{}, false
	}
	discriminant = math.Sqrt(discriminant)

	ax := D * d.Y
	bx := d.X * discriminant
	if d.Y < 0.0 {
		bx = -bx
	}
	ay := -D * d.X
	by := math.Abs(d.Y) * discriminant
	return c.Add(Point{ax + bx, ay + by}), c.Add(Point{ax - bx, ay - by}), true
}

// https://math.stackexchange.com/questions/256100/how-can-i-find-the-points-at-which-two-circles-intersect
// https://gist.github.com/jupdike/bfe5eb23d1c395d8a0a1a4ddd94882ac
func intersectionCircleCircle(c0 Point, r0 float64, c1 Point, r1 float64) (Point, Point, bool) {
	R := c0.Sub(c1).Length()
	if R < math.Abs(r0-r1) || r0+r1 < R || c0.Equals(c1) {
		return Point{}, Point{}, false
	}
	R2 := R * R

	k := r0*r0 - r1*r1
	a := 0.5
	b := 0.5 * k / R2
	c := 0.5 * math.Sqrt(2.0*(r0*r0+r1*r1)/R2-k*k/(R2*R2)-1.0)

	i0 := c0.Add(c1).Mul(a)
	i1 := c1.Sub(c0).Mul(b)
	i2 := Point{c1.Y - c0.Y, c0.X - c1.X}.Mul(c)
	return i0.Add(i1).Add(i2), i0.Add(i1).Sub(i2), true
}
