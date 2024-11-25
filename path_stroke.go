package canvas

import (
	"math"
)

// NOTE: implementation inspired from github.com/golang/freetype/raster/stroke.go

// Capper implements Cap, with rhs the path to append to, halfWidth the half width of the stroke, pivot the pivot point around which to construct a cap, and n0 the normal at the start of the path. The length of n0 is equal to the halfWidth.
type Capper interface {
	Cap(*Path, float64, Point, Point)
}

// RoundCap caps the start or end of a path by a round cap.
var RoundCap Capper = RoundCapper{}

// RoundCapper is a round capper.
type RoundCapper struct{}

// Cap adds a cap to path p of width 2*halfWidth, at a pivot point and initial normal direction of n0.
func (RoundCapper) Cap(p *Path, halfWidth float64, pivot, n0 Point) {
	end := pivot.Sub(n0)
	p.ArcTo(halfWidth, halfWidth, 0, false, true, end.X, end.Y)
}

func (RoundCapper) String() string {
	return "Round"
}

// ButtCap caps the start or end of a path by a butt cap.
var ButtCap Capper = ButtCapper{}

// ButtCapper is a butt capper.
type ButtCapper struct{}

// Cap adds a cap to path p of width 2*halfWidth, at a pivot point and initial normal direction of n0.
func (ButtCapper) Cap(p *Path, halfWidth float64, pivot, n0 Point) {
	end := pivot.Sub(n0)
	p.LineTo(end.X, end.Y)
}

func (ButtCapper) String() string {
	return "Butt"
}

// SquareCap caps the start or end of a path by a square cap.
var SquareCap Capper = SquareCapper{}

// SquareCapper is a square capper.
type SquareCapper struct{}

// Cap adds a cap to path p of width 2*halfWidth, at a pivot point and initial normal direction of n0.
func (SquareCapper) Cap(p *Path, halfWidth float64, pivot, n0 Point) {
	e := n0.Rot90CCW()
	corner1 := pivot.Add(e).Add(n0)
	corner2 := pivot.Add(e).Sub(n0)
	end := pivot.Sub(n0)
	p.LineTo(corner1.X, corner1.Y)
	p.LineTo(corner2.X, corner2.Y)
	p.LineTo(end.X, end.Y)
}

func (SquareCapper) String() string {
	return "Square"
}

////////////////

// Joiner implements Join, with rhs the right path and lhs the left path to append to, pivot the intersection of both path elements, n0 and n1 the normals at the start and end of the path respectively. The length of n0 and n1 are equal to the halfWidth.
type Joiner interface {
	Join(*Path, *Path, float64, Point, Point, Point, float64, float64)
}

// BevelJoin connects two path elements by a linear join.
var BevelJoin Joiner = BevelJoiner{}

// BevelJoiner is a bevel joiner.
type BevelJoiner struct{}

// Join adds a join to a right-hand-side and left-hand-side path, of width 2*halfWidth, around a pivot point with starting and ending normals of n0 and n1, and radius of curvatures of the previous and next segments.
func (BevelJoiner) Join(rhs, lhs *Path, halfWidth float64, pivot, n0, n1 Point, r0, r1 float64) {
	rEnd := pivot.Add(n1)
	lEnd := pivot.Sub(n1)
	rhs.LineTo(rEnd.X, rEnd.Y)
	lhs.LineTo(lEnd.X, lEnd.Y)
}

func (BevelJoiner) String() string {
	return "Bevel"
}

// RoundJoin connects two path elements by a round join.
var RoundJoin Joiner = RoundJoiner{}

// RoundJoiner is a round joiner.
type RoundJoiner struct{}

// Join adds a join to a right-hand-side and left-hand-side path, of width 2*halfWidth, around a pivot point with starting and ending normals of n0 and n1, and radius of curvatures of the previous and next segments.
func (RoundJoiner) Join(rhs, lhs *Path, halfWidth float64, pivot, n0, n1 Point, r0, r1 float64) {
	rEnd := pivot.Add(n1)
	lEnd := pivot.Sub(n1)
	cw := 0.0 <= n0.Rot90CW().Dot(n1)
	if cw { // bend to the right, ie. CW (or 180 degree turn)
		rhs.LineTo(rEnd.X, rEnd.Y)
		lhs.ArcTo(halfWidth, halfWidth, 0.0, false, false, lEnd.X, lEnd.Y)
	} else { // bend to the left, ie. CCW
		rhs.ArcTo(halfWidth, halfWidth, 0.0, false, true, rEnd.X, rEnd.Y)
		lhs.LineTo(lEnd.X, lEnd.Y)
	}
}

func (RoundJoiner) String() string {
	return "Round"
}

// MiterJoin connects two path elements by extending the ends of the paths as lines until they meet. If this point is further than the limit, this will result in a bevel join (MiterJoin) or they will meet at the limit (MiterClipJoin).
var MiterJoin Joiner = MiterJoiner{BevelJoin, 4.0}
var MiterClipJoin Joiner = MiterJoiner{nil, 4.0} // TODO: should extend limit*halfwidth before bevel

// MiterJoiner is a miter joiner.
type MiterJoiner struct {
	GapJoiner Joiner
	Limit     float64
}

// Join adds a join to a right-hand-side and left-hand-side path, of width 2*halfWidth, around a pivot point with starting and ending normals of n0 and n1, and radius of curvatures of the previous and next segments.
func (j MiterJoiner) Join(rhs, lhs *Path, halfWidth float64, pivot, n0, n1 Point, r0, r1 float64) {
	if n0.Equals(n1.Neg()) {
		BevelJoin.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
		return
	}

	cw := 0.0 <= n0.Rot90CW().Dot(n1)
	hw := halfWidth
	if cw {
		hw = -hw // used to calculate |R|, when running CW then n0 and n1 point the other way, so the sign of r0 and r1 is negated
	}

	// note that cos(theta) below refers to sin(theta/2) in the documentation of stroke-miterlimit
	// in https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/stroke-miterlimit
	theta := n0.AngleBetween(n1) / 2.0 // half the angle between normals
	d := hw / math.Cos(theta)          // half the miter length
	limit := math.Max(j.Limit, 1.001)  // otherwise nearly linear joins will also get clipped
	clip := !math.IsNaN(limit) && limit*halfWidth < math.Abs(d)
	if clip && j.GapJoiner != nil {
		j.GapJoiner.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
		return
	}

	rEnd := pivot.Add(n1)
	lEnd := pivot.Sub(n1)
	mid := pivot.Add(n0.Add(n1).Norm(d))
	if clip {
		// miter-clip
		t := math.Abs(limit * halfWidth / d)
		if cw { // bend to the right, ie. CW
			mid0 := lhs.Pos().Interpolate(mid, t)
			mid1 := lEnd.Interpolate(mid, t)
			lhs.LineTo(mid0.X, mid0.Y)
			lhs.LineTo(mid1.X, mid1.Y)
		} else {
			mid0 := rhs.Pos().Interpolate(mid, t)
			mid1 := rEnd.Interpolate(mid, t)
			rhs.LineTo(mid0.X, mid0.Y)
			rhs.LineTo(mid1.X, mid1.Y)
		}
	} else {
		if cw { // bend to the right, ie. CW
			lhs.LineTo(mid.X, mid.Y)
		} else {
			rhs.LineTo(mid.X, mid.Y)
		}
	}
	rhs.LineTo(rEnd.X, rEnd.Y)
	lhs.LineTo(lEnd.X, lEnd.Y)
}

func (j MiterJoiner) String() string {
	if j.GapJoiner == nil {
		return "MiterClip"
	}
	return "Miter"
}

// ArcsJoin connects two path elements by extending the ends of the paths as circle arcs until they meet. If this point is further than the limit, this will result in a bevel join (ArcsJoin) or they will meet at the limit (ArcsClipJoin).
var ArcsJoin Joiner = ArcsJoiner{BevelJoin, 4.0}
var ArcsClipJoin Joiner = ArcsJoiner{nil, 4.0}

// ArcsJoiner is an arcs joiner.
type ArcsJoiner struct {
	GapJoiner Joiner
	Limit     float64
}

func closestArcIntersection(c Point, cw bool, pivot, i0, i1 Point) Point {
	thetaPivot := pivot.Sub(c).Angle()
	dtheta0 := i0.Sub(c).Angle() - thetaPivot
	dtheta1 := i1.Sub(c).Angle() - thetaPivot
	if cw { // arc runs clockwise, so look the other way around
		dtheta0 = -dtheta0
		dtheta1 = -dtheta1
	}
	if angleNorm(dtheta1) < angleNorm(dtheta0) {
		return i1
	}
	return i0
}

// Join adds a join to a right-hand-side and left-hand-side path, of width 2*halfWidth, around a pivot point with starting and ending normals of n0 and n1, and radius of curvatures of the previous and next segments, which are positive for CCW arcs.
func (j ArcsJoiner) Join(rhs, lhs *Path, halfWidth float64, pivot, n0, n1 Point, r0, r1 float64) {
	if n0.Equals(n1.Neg()) {
		BevelJoin.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
		return
	} else if math.IsNaN(r0) && math.IsNaN(r1) {
		MiterJoiner(j).Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
		return
	}
	limit := math.Max(j.Limit, 1.001) // 1.001 so that nearly linear joins will not get clipped

	cw := 0.0 <= n0.Rot90CW().Dot(n1)
	hw := halfWidth
	if cw {
		hw = -hw // used to calculate |R|, when running CW then n0 and n1 point the other way, so the sign of r0 and r1 is negated
	}

	// r is the radius of the original curve, R the radius of the stroke curve, c are the centers of the circles
	c0 := pivot.Add(n0.Norm(-r0))
	c1 := pivot.Add(n1.Norm(-r1))
	R0, R1 := math.Abs(r0+hw), math.Abs(r1+hw)

	// TODO: can simplify if intersection returns angles too?
	var i0, i1 Point
	var ok bool
	if math.IsNaN(r0) {
		line := pivot.Add(n0)
		if cw {
			line = pivot.Sub(n0)
		}
		i0, i1, ok = intersectionRayCircle(line, line.Add(n0.Rot90CCW()), c1, R1)
	} else if math.IsNaN(r1) {
		line := pivot.Add(n1)
		if cw {
			line = pivot.Sub(n1)
		}
		i0, i1, ok = intersectionRayCircle(line, line.Add(n1.Rot90CCW()), c0, R0)
	} else {
		i0, i1, ok = intersectionCircleCircle(c0, R0, c1, R1)
	}
	if !ok {
		// no intersection
		BevelJoin.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
		return
	}

	// find the closest intersection when following the arc (using either arc r0 or r1 with center c0 or c1 respectively)
	var mid Point
	if !math.IsNaN(r0) {
		mid = closestArcIntersection(c0, r0 < 0.0, pivot, i0, i1)
	} else {
		mid = closestArcIntersection(c1, 0.0 <= r1, pivot, i0, i1)
	}

	// check arc limit
	d := mid.Sub(pivot).Length()
	clip := !math.IsNaN(limit) && limit*halfWidth < d
	if clip && j.GapJoiner != nil {
		j.GapJoiner.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
		return
	}

	mid2 := mid
	if clip {
		// arcs-clip
		start, end := pivot.Add(n0), pivot.Add(n1)
		if cw {
			start, end = pivot.Sub(n0), pivot.Sub(n1)
		}

		var clipMid, clipNormal Point
		if !math.IsNaN(r0) && !math.IsNaN(r1) && (0.0 < r0) == (0.0 < r1) {
			// circle have opposite direction/sweep
			// NOTE: this may cause the bevel to be imperfectly oriented
			clipMid = mid.Sub(pivot).Norm(limit * halfWidth)
			clipNormal = clipMid.Rot90CCW()
		} else {
			// circle in between both stroke edges
			rMid := (r0 - r1) / 2.0
			if math.IsNaN(r0) {
				rMid = -(r1 + hw) * 2.0
			} else if math.IsNaN(r1) {
				rMid = (r0 + hw) * 2.0
			}

			sweep := 0.0 < rMid
			RMid := math.Abs(rMid)
			cx, cy, a0, _ := ellipseToCenter(pivot.X, pivot.Y, RMid, RMid, 0.0, false, sweep, mid.X, mid.Y)
			cMid := Point{cx, cy}
			dtheta := limit * halfWidth / rMid

			clipMid = EllipsePos(RMid, RMid, 0.0, cMid.X, cMid.Y, a0+dtheta)
			clipNormal = ellipseNormal(RMid, RMid, 0.0, sweep, a0+dtheta, 1.0)
		}

		if math.IsNaN(r1) {
			i0, ok = intersectionRayLine(clipMid, clipMid.Add(clipNormal), mid, end)
			if !ok {
				// not sure when this occurs
				BevelJoin.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
				return
			}
			mid2 = i0
		} else {
			i0, i1, ok = intersectionRayCircle(clipMid, clipMid.Add(clipNormal), c1, R1)
			if !ok {
				// not sure when this occurs
				BevelJoin.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
				return
			}
			mid2 = closestArcIntersection(c1, 0.0 <= r1, pivot, i0, i1)
		}

		if math.IsNaN(r0) {
			i0, ok = intersectionRayLine(clipMid, clipMid.Add(clipNormal), start, mid)
			if !ok {
				// not sure when this occurs
				BevelJoin.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
				return
			}
			mid = i0
		} else {
			i0, i1, ok = intersectionRayCircle(clipMid, clipMid.Add(clipNormal), c0, R0)
			if !ok {
				// not sure when this occurs
				BevelJoin.Join(rhs, lhs, halfWidth, pivot, n0, n1, r0, r1)
				return
			}
			mid = closestArcIntersection(c0, r0 < 0.0, pivot, i0, i1)
		}
	}

	rEnd := pivot.Add(n1)
	lEnd := pivot.Sub(n1)
	if cw { // bend to the right, ie. CW
		rhs.LineTo(rEnd.X, rEnd.Y)
		if math.IsNaN(r0) {
			lhs.LineTo(mid.X, mid.Y)
		} else {
			lhs.ArcTo(R0, R0, 0.0, false, 0.0 < r0, mid.X, mid.Y)
		}
		if clip {
			lhs.LineTo(mid2.X, mid2.Y)
		}
		if math.IsNaN(r1) {
			lhs.LineTo(lEnd.X, lEnd.Y)
		} else {
			lhs.ArcTo(R1, R1, 0.0, false, 0.0 < r1, lEnd.X, lEnd.Y)
		}
	} else { // bend to the left, ie. CCW
		if math.IsNaN(r0) {
			rhs.LineTo(mid.X, mid.Y)
		} else {
			rhs.ArcTo(R0, R0, 0.0, false, 0.0 < r0, mid.X, mid.Y)
		}
		if clip {
			rhs.LineTo(mid2.X, mid2.Y)
		}
		if math.IsNaN(r1) {
			rhs.LineTo(rEnd.X, rEnd.Y)
		} else {
			rhs.ArcTo(R1, R1, 0.0, false, 0.0 < r1, rEnd.X, rEnd.Y)
		}
		lhs.LineTo(lEnd.X, lEnd.Y)
	}
}

func (j ArcsJoiner) String() string {
	if j.GapJoiner == nil {
		return "ArcsClip"
	}
	return "Arcs"
}

func (p *Path) optimizeInnerBend(i int) {
	// i is the index of the line segment in the inner bend connecting both edges
	ai := i - cmdLen(p.d[i-1])
	bi := i + cmdLen(p.d[i])
	if ai == 0 {
		return
	}

	a0 := Point{p.d[ai-3], p.d[ai-2]}
	b0 := Point{p.d[bi-3], p.d[bi-2]}
	if bi == len(p.d) {
		// inner bend is at the path's start
		bi = 4
	}

	// TODO: implement other segment combinations
	zs_ := [2]Intersection{}
	zs := zs_[:]
	if (p.d[ai] == LineToCmd || p.d[ai] == CloseCmd) && (p.d[bi] == LineToCmd || p.d[bi] == CloseCmd) {
		zs = intersectionSegment(zs[:0], a0, p.d[ai:ai+4], b0, p.d[bi:bi+4])
		// TODO: check conditions for pathological cases
		if len(zs) == 1 && zs[0].T[0] != 0.0 && zs[0].T[0] != 1.0 && zs[0].T[1] != 0.0 && zs[0].T[1] != 1.0 {
			p.d[ai+1] = zs[0].X
			p.d[ai+2] = zs[0].Y
			if bi == 4 {
				// inner bend is at the path's start
				p.d = p.d[:i]
				p.d[1] = zs[0].X
				p.d[2] = zs[0].Y
			} else {
				p.d = append(p.d[:i], p.d[bi:]...)
			}
		}
	}
}

type pathStrokeState struct {
	cmd    float64
	p0, p1 Point   // position of start and end
	n0, n1 Point   // normal of start and end (points right when walking the path)
	r0, r1 float64 // radius of start and end

	cp1, cp2                    Point   // Béziers
	rx, ry, rot, theta0, theta1 float64 // arcs
	large, sweep                bool    // arcs
}

// offset returns the rhs and lhs paths from offsetting a path (must not have subpaths). It closes rhs and lhs when p is closed as well.
func (p *Path) offset(halfWidth float64, cr Capper, jr Joiner, strokeOpen bool, tolerance float64) (*Path, *Path) {
	// only non-empty paths are evaluated
	closed := false
	states := []pathStrokeState{}
	var start, end Point
	for i := 0; i < len(p.d); {
		cmd := p.d[i]
		switch cmd {
		case MoveToCmd:
			end = Point{p.d[i+1], p.d[i+2]}
		case LineToCmd:
			end = Point{p.d[i+1], p.d[i+2]}
			n := end.Sub(start).Rot90CW().Norm(halfWidth)
			states = append(states, pathStrokeState{
				cmd: LineToCmd,
				p0:  start,
				p1:  end,
				n0:  n,
				n1:  n,
				r0:  math.NaN(),
				r1:  math.NaN(),
			})
		case QuadToCmd, CubeToCmd:
			var cp1, cp2 Point
			if cmd == QuadToCmd {
				cp := Point{p.d[i+1], p.d[i+2]}
				end = Point{p.d[i+3], p.d[i+4]}
				cp1, cp2 = quadraticToCubicBezier(start, cp, end)
			} else {
				cp1 = Point{p.d[i+1], p.d[i+2]}
				cp2 = Point{p.d[i+3], p.d[i+4]}
				end = Point{p.d[i+5], p.d[i+6]}
			}
			n0 := cubicBezierNormal(start, cp1, cp2, end, 0.0, halfWidth)
			n1 := cubicBezierNormal(start, cp1, cp2, end, 1.0, halfWidth)
			r0 := cubicBezierCurvatureRadius(start, cp1, cp2, end, 0.0)
			r1 := cubicBezierCurvatureRadius(start, cp1, cp2, end, 1.0)
			states = append(states, pathStrokeState{
				cmd: CubeToCmd,
				p0:  start,
				p1:  end,
				n0:  n0,
				n1:  n1,
				r0:  r0,
				r1:  r1,
				cp1: cp1,
				cp2: cp2,
			})
		case ArcToCmd:
			rx, ry, phi := p.d[i+1], p.d[i+2], p.d[i+3]
			large, sweep := toArcFlags(p.d[i+4])
			end = Point{p.d[i+5], p.d[i+6]}
			_, _, theta0, theta1 := ellipseToCenter(start.X, start.Y, rx, ry, phi, large, sweep, end.X, end.Y)
			n0 := ellipseNormal(rx, ry, phi, sweep, theta0, halfWidth)
			n1 := ellipseNormal(rx, ry, phi, sweep, theta1, halfWidth)
			r0 := ellipseCurvatureRadius(rx, ry, sweep, theta0)
			r1 := ellipseCurvatureRadius(rx, ry, sweep, theta1)
			states = append(states, pathStrokeState{
				cmd:    ArcToCmd,
				p0:     start,
				p1:     end,
				n0:     n0,
				n1:     n1,
				r0:     r0,
				r1:     r1,
				rx:     rx,
				ry:     ry,
				rot:    phi * 180.0 / math.Pi,
				theta0: theta0,
				theta1: theta1,
				large:  large,
				sweep:  sweep,
			})
		case CloseCmd:
			end = Point{p.d[i+1], p.d[i+2]}
			if !Equal(start.X, end.X) || !Equal(start.Y, end.Y) {
				n := end.Sub(start).Rot90CW().Norm(halfWidth)
				states = append(states, pathStrokeState{
					cmd: LineToCmd,
					p0:  start,
					p1:  end,
					n0:  n,
					n1:  n,
					r0:  math.NaN(),
					r1:  math.NaN(),
				})
			}
			closed = true
		}
		start = end
		i += cmdLen(cmd)
	}

	rhs, lhs := &Path{}, &Path{}
	rStart := states[0].p0.Add(states[0].n0)
	lStart := states[0].p0.Sub(states[0].n0)
	rhs.MoveTo(rStart.X, rStart.Y)
	lhs.MoveTo(lStart.X, lStart.Y)
	rhsJoinIndex, lhsJoinIndex := -1, -1
	for i, cur := range states {
		switch cur.cmd {
		case LineToCmd:
			rEnd := cur.p1.Add(cur.n1)
			lEnd := cur.p1.Sub(cur.n1)
			rhs.LineTo(rEnd.X, rEnd.Y)
			lhs.LineTo(lEnd.X, lEnd.Y)
		case CubeToCmd:
			rhs = rhs.Join(strokeCubicBezier(cur.p0, cur.cp1, cur.cp2, cur.p1, halfWidth, tolerance))
			lhs = lhs.Join(strokeCubicBezier(cur.p0, cur.cp1, cur.cp2, cur.p1, -halfWidth, tolerance))
		case ArcToCmd:
			rStart := cur.p0.Add(cur.n0)
			lStart := cur.p0.Sub(cur.n0)
			rEnd := cur.p1.Add(cur.n1)
			lEnd := cur.p1.Sub(cur.n1)
			dr := halfWidth
			if !cur.sweep { // bend to the right, ie. CW
				dr = -dr
			}

			rLambda := ellipseRadiiCorrection(rStart, cur.rx+dr, cur.ry+dr, cur.rot*math.Pi/180.0, rEnd)
			lLambda := ellipseRadiiCorrection(lStart, cur.rx-dr, cur.ry-dr, cur.rot*math.Pi/180.0, lEnd)
			if rLambda <= 1.0 && lLambda <= 1.0 {
				rLambda, lLambda = 1.0, 1.0
			}
			rhs.ArcTo(rLambda*(cur.rx+dr), rLambda*(cur.ry+dr), cur.rot, cur.large, cur.sweep, rEnd.X, rEnd.Y)
			lhs.ArcTo(lLambda*(cur.rx-dr), lLambda*(cur.ry-dr), cur.rot, cur.large, cur.sweep, lEnd.X, lEnd.Y)
		}

		// optimize inner bend
		if 0 < i {
			prev := states[i-1]
			cw := 0.0 <= prev.n1.Rot90CW().Dot(cur.n0)
			if cw && rhsJoinIndex != -1 {
				rhs.optimizeInnerBend(rhsJoinIndex)
			} else if !cw && lhsJoinIndex != -1 {
				lhs.optimizeInnerBend(lhsJoinIndex)
			}
		}
		rhsJoinIndex = -1
		lhsJoinIndex = -1

		// join the cur and next path segments
		if i+1 < len(states) || closed {
			next := states[0]
			if i+1 < len(states) {
				next = states[i+1]
			}
			if !cur.n1.Equals(next.n0) {
				rhsJoinIndex = len(rhs.d)
				lhsJoinIndex = len(lhs.d)
				jr.Join(rhs, lhs, halfWidth, cur.p1, cur.n1, next.n0, cur.r1, next.r0)
			}
		}
	}

	if closed {
		rhs.Close()
		lhs.Close()

		// optimize inner bend
		if 1 < len(states) {
			cw := 0.0 <= states[len(states)-1].n1.Rot90CW().Dot(states[0].n0)
			if cw && rhsJoinIndex != -1 {
				rhs.optimizeInnerBend(rhsJoinIndex)
			} else if !cw && lhsJoinIndex != -1 {
				lhs.optimizeInnerBend(lhsJoinIndex)
			}
		}

		rhs.optimizeClose()
		lhs.optimizeClose()
	} else if strokeOpen {
		lhs = lhs.Reverse()
		cr.Cap(rhs, halfWidth, states[len(states)-1].p1, states[len(states)-1].n1)
		rhs = rhs.Join(lhs)
		cr.Cap(rhs, halfWidth, states[0].p0, states[0].n0.Neg())
		lhs = nil

		rhs.Close()
		rhs.optimizeClose()
	}
	return rhs, lhs
}

// Offset offsets the path to expand by w and returns a new path. If w is negative it will contract. For open paths, a positive w will offset the path to the right-hand side. The tolerance is the maximum deviation from the actual offset when flattening Béziers and optimizing the path. Subpaths may not (self-)intersect, use Settle to remove (self-)intersections.
func (p *Path) Offset(w float64, fillRule FillRule, tolerance float64) *Path {
	if Equal(w, 0.0) {
		return p
	}

	positive := 0.0 < w
	w = math.Abs(w)

	q := &Path{}
	filling := p.Filling(fillRule)
	for i, pi := range p.Split() {
		r := &Path{}
		ccw := pi.CCW()
		closed := pi.Closed()
		rhs, lhs := pi.offset(w, ButtCap, RoundJoin, false, tolerance)
		if !closed || (ccw != filling[i]) != positive {
			r = rhs
		} else {
			r = lhs
		}

		if closed {
			r = r.Settle(Positive)
			if !filling[i] {
				r = r.Reverse()
			}
		}
		q = q.Append(r)
	}
	return q
}

// Stroke converts a path into a stroke of width w and returns a new path. It uses cr to cap the start and end of the path, and jr to join all path elements. If the path closes itself, it will use a join between the start and end instead of capping them. The tolerance is the maximum deviation from the original path when flattening Béziers and optimizing the stroke.
func (p *Path) Stroke(w float64, cr Capper, jr Joiner, tolerance float64) *Path {
	if cr == nil {
		cr = ButtCap
	}
	if jr == nil {
		jr = MiterJoin
	}
	q := &Path{}
	halfWidth := w / 2.0
	for _, pi := range p.Split() {
		rhs, lhs := pi.offset(halfWidth, cr, jr, true, tolerance)
		if lhs != nil { // closed path
			// inner path should go opposite direction to cancel the outer path
			q = q.Append(rhs.Settle(Positive))
			q = q.Append(lhs.Settle(Positive).Reverse())
		} else {
			q = q.Append(rhs.Settle(Positive))
		}
	}
	return q
}
