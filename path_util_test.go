package canvas

import (
	"fmt"
	"math"
	"testing"

	"github.com/tdewolff/test"
)

func TestEllipse(t *testing.T) {
	test.T(t, EllipsePos(2.0, 1.0, math.Pi/2.0, 1.0, 0.5, 0.0), Point{1.0, 2.5})
	test.T(t, ellipseDeriv(2.0, 1.0, math.Pi/2.0, true, 0.0), Point{-1.0, 0.0})
	test.T(t, ellipseDeriv(2.0, 1.0, math.Pi/2.0, false, 0.0), Point{1.0, 0.0})
	test.T(t, ellipseDeriv2(2.0, 1.0, math.Pi/2.0, 0.0), Point{0.0, -2.0})
	test.T(t, ellipseCurvatureRadius(2.0, 1.0, true, 0.0), 0.5)
	test.T(t, ellipseCurvatureRadius(2.0, 1.0, false, 0.0), -0.5)
	test.T(t, ellipseCurvatureRadius(2.0, 1.0, true, math.Pi/2.0), 4.0)
	if !math.IsNaN(ellipseCurvatureRadius(2.0, 0.0, true, 0.0)) {
		test.Fail(t)
	}
	test.T(t, ellipseNormal(2.0, 1.0, math.Pi/2.0, true, 0.0, 1.0), Point{0.0, 1.0})
	test.T(t, ellipseNormal(2.0, 1.0, math.Pi/2.0, false, 0.0, 1.0), Point{0.0, -1.0})

	// https://www.wolframalpha.com/input/?i=arclength+x%28t%29%3D2*cos+t%2C+y%28t%29%3Dsin+t+for+t%3D0+to+0.5pi
	test.Float(t, ellipseLength(2.0, 1.0, 0.0, math.Pi/2.0), 2.4221102220)

	test.Float(t, ellipseRadiiCorrection(Point{0.0, 0.0}, 0.1, 0.1, 0.0, Point{1.0, 0.0}), 5.0)
}

func TestEllipseToCenter(t *testing.T) {
	var tests = []struct {
		x1, y1       float64
		rx, ry, phi  float64
		large, sweep bool
		x2, y2       float64

		cx, cy, theta0, theta1 float64
	}{
		{0.0, 0.0, 2.0, 2.0, 0.0, false, false, 2.0, 2.0, 2.0, 0.0, math.Pi, math.Pi / 2.0},
		{0.0, 0.0, 2.0, 2.0, 0.0, true, false, 2.0, 2.0, 0.0, 2.0, math.Pi * 3.0 / 2.0, 0.0},
		{0.0, 0.0, 2.0, 2.0, 0.0, true, true, 2.0, 2.0, 2.0, 0.0, math.Pi, math.Pi * 5.0 / 2.0},
		{0.0, 0.0, 2.0, 1.0, math.Pi / 2.0, false, false, 1.0, 2.0, 1.0, 0.0, math.Pi / 2.0, 0.0},

		// radius correction
		{0.0, 0.0, 0.1, 0.1, 0.0, false, false, 1.0, 0.0, 0.5, 0.0, math.Pi, 0.0},

		// start == end
		{0.0, 0.0, 1.0, 1.0, 0.0, false, false, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},

		// precision issues
		{8.2, 18.0, 0.2, 0.2, 0.0, false, true, 7.8, 18.0, 8.0, 18.0, 0.0, math.Pi},
		{7.8, 18.0, 0.2, 0.2, 0.0, false, true, 8.2, 18.0, 8.0, 18.0, math.Pi, 2.0 * math.Pi},

		// bugs
		{-1.0 / math.Sqrt(2), 0.0, 1.0, 1.0, 0.0, false, false, 1.0 / math.Sqrt(2.0), 0.0, 0.0, -1.0 / math.Sqrt(2.0), 3.0 / 4.0 * math.Pi, 1.0 / 4.0 * math.Pi},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("(%g,%g) %g %g %g %v %v (%g,%g)", tt.x1, tt.y1, tt.rx, tt.ry, tt.phi, tt.large, tt.sweep, tt.x2, tt.y2), func(t *testing.T) {
			cx, cy, theta0, theta1 := ellipseToCenter(tt.x1, tt.y1, tt.rx, tt.ry, tt.phi, tt.large, tt.sweep, tt.x2, tt.y2)
			test.Floats(t, []float64{cx, cy, theta0, theta1}, []float64{tt.cx, tt.cy, tt.theta0, tt.theta1})
		})
	}

	//cx, cy, theta0, theta1 := ellipseToCenter(0.0, 0.0, 2.0, 2.0, 0.0, false, false, 2.0, 2.0)
	//test.Float(t, cx, 2.0)
	//test.Float(t, cy, 0.0)
	//test.Float(t, theta0, math.Pi)
	//test.Float(t, theta1, math.Pi/2.0)

	//cx, cy, theta0, theta1 = ellipseToCenter(0.0, 0.0, 2.0, 2.0, 0.0, true, false, 2.0, 2.0)
	//test.Float(t, cx, 0.0)
	//test.Float(t, cy, 2.0)
	//test.Float(t, theta0, math.Pi*3.0/2.0)
	//test.Float(t, theta1, 0.0)

	//cx, cy, theta0, theta1 = ellipseToCenter(0.0, 0.0, 2.0, 2.0, 0.0, true, true, 2.0, 2.0)
	//test.Float(t, cx, 2.0)
	//test.Float(t, cy, 0.0)
	//test.Float(t, theta0, math.Pi)
	//test.Float(t, theta1, math.Pi*5.0/2.0)

	//cx, cy, theta0, theta1 = ellipseToCenter(0.0, 0.0, 2.0, 1.0, math.Pi/2.0, false, false, 1.0, 2.0)
	//test.Float(t, cx, 1.0)
	//test.Float(t, cy, 0.0)
	//test.Float(t, theta0, math.Pi/2.0)
	//test.Float(t, theta1, 0.0)

	//cx, cy, theta0, theta1 = ellipseToCenter(0.0, 0.0, 0.1, 0.1, 0.0, false, false, 1.0, 0.0)
	//test.Float(t, cx, 0.5)
	//test.Float(t, cy, 0.0)
	//test.Float(t, theta0, math.Pi)
	//test.Float(t, theta1, 0.0)

	//cx, cy, theta0, theta1 = ellipseToCenter(0.0, 0.0, 1.0, 1.0, 0.0, false, false, 0.0, 0.0)
	//test.Float(t, cx, 0.0)
	//test.Float(t, cy, 0.0)
	//test.Float(t, theta0, 0.0)
	//test.Float(t, theta1, 0.0)
}

func TestEllipseSplit(t *testing.T) {
	mid, large0, large1, ok := ellipseSplit(2.0, 1.0, 0.0, 0.0, 0.0, math.Pi, 0.0, math.Pi/2.0)
	test.That(t, ok)
	test.T(t, mid, Point{0.0, 1.0})
	test.That(t, !large0)
	test.That(t, !large1)

	_, _, _, ok = ellipseSplit(2.0, 1.0, 0.0, 0.0, 0.0, math.Pi, 0.0, -math.Pi/2.0)
	test.That(t, !ok)

	mid, large0, large1, ok = ellipseSplit(2.0, 1.0, 0.0, 0.0, 0.0, 0.0, math.Pi*7.0/4.0, math.Pi/2.0)
	test.That(t, ok)
	test.T(t, mid, Point{0.0, 1.0})
	test.That(t, !large0)
	test.That(t, large1)

	mid, large0, large1, ok = ellipseSplit(2.0, 1.0, 0.0, 0.0, 0.0, 0.0, math.Pi*7.0/4.0, math.Pi*3.0/2.0)
	test.That(t, ok)
	test.T(t, mid, Point{0.0, -1.0})
	test.That(t, large0)
	test.That(t, !large1)
}

func TestArcToQuad(t *testing.T) {
	test.T(t, arcToQuad(Point{0.0, 0.0}, 100.0, 100.0, 0.0, false, false, Point{200.0, 0.0}), MustParseSVGPath("Q0 100 100 100Q200 100 200 0"))
}

func TestArcToCube(t *testing.T) {
	defer setEpsilon(1e-3)()
	test.T(t, arcToCube(Point{0.0, 0.0}, 100.0, 100.0, 0.0, false, false, Point{200.0, 0.0}), MustParseSVGPath("C0 54.858 45.142 100 100 100C154.858 100 200 54.858 200 0"))
}

func TestXMonotoneEllipse(t *testing.T) {
	test.T(t, xmonotoneEllipticArc(Point{0.0, 0.0}, 100.0, 50.0, 0.0, false, false, Point{0.0, 100.0}), MustParseSVGPath("M0 0A100 50 0 0 0 -100 50A100 50 0 0 0 0 100"))

	defer setEpsilon(1e-3)()
	test.T(t, xmonotoneEllipticArc(Point{0.0, 0.0}, 50.0, 25.0, math.Pi/4.0, false, false, Point{100.0 / math.Sqrt(2.0), 100.0 / math.Sqrt(2.0)}), MustParseSVGPath("M0 0A50 25 45 0 0 -4.1731 11.6383A50 25 45 0 0 70.71067811865474 70.71067811865474"))
}

func TestFlattenEllipse(t *testing.T) {
	defer setEpsilon(1e-3)()
	tolerance := 1.0

	// circular
	test.T(t, flattenEllipticArc(Point{0.0, 0.0}, 100.0, 100.0, 0.0, false, false, Point{200.0, 0.0}, tolerance), MustParseSVGPath("L3.8513 30.6285L20.8474 62.5902L48.0297 86.4971L81.9001 99.2726L118.0998 99.2726L151.9702 86.4971L179.1525 62.5902L196.1486 30.6285L200 0"))
}

func TestQuadraticBezier(t *testing.T) {
	p1, p2 := quadraticToCubicBezier(Point{0.0, 0.0}, Point{1.5, 0.0}, Point{3.0, 0.0})
	test.T(t, p1, Point{1.0, 0.0})
	test.T(t, p2, Point{2.0, 0.0})

	p1, p2 = quadraticToCubicBezier(Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0})
	test.T(t, p1, Point{2.0 / 3.0, 0.0})
	test.T(t, p2, Point{1.0, 1.0 / 3.0})

	p0, p1, p2, q0, q1, q2 := quadraticBezierSplit(Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 0.5)
	test.T(t, p0, Point{0.0, 0.0})
	test.T(t, p1, Point{0.5, 0.0})
	test.T(t, p2, Point{0.75, 0.25})
	test.T(t, q0, Point{0.75, 0.25})
	test.T(t, q1, Point{1.0, 0.5})
	test.T(t, q2, Point{1.0, 1.0})
}

func TestQuadraticBezierPos(t *testing.T) {
	var tests = []struct {
		p0, p1, p2 Point
		t          float64
		q          Point
	}{
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 0.0, Point{0.0, 0.0}},
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 0.5, Point{0.75, 0.25}},
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 1.0, Point{1.0, 1.0}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.t), func(t *testing.T) {
			q := quadraticBezierPos(tt.p0, tt.p1, tt.p2, tt.t)
			test.T(t, q, tt.q)
		})
	}
}

func TestQuadraticBezierDeriv(t *testing.T) {
	var tests = []struct {
		p0, p1, p2 Point
		t          float64
		q          Point
	}{
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 0.0, Point{2.0, 0.0}},
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 0.5, Point{1.0, 1.0}},
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 1.0, Point{0.0, 2.0}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.t), func(t *testing.T) {
			q := quadraticBezierDeriv(tt.p0, tt.p1, tt.p2, tt.t)
			test.T(t, q, tt.q)
		})
	}
}

func TestQuadraticBezierLength(t *testing.T) {
	var tests = []struct {
		p0, p1, p2 Point
		l          float64
	}{
		{Point{0.0, 0.0}, Point{0.5, 0.0}, Point{2.0, 0.0}, 2.0},
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{2.0, 0.0}, 2.0},

		// https://www.wolframalpha.com/input/?i=length+of+the+curve+%7Bx%3D2*%281-t%29*t*1.00+%2B+t%5E2*1.00%2C+y%3Dt%5E2*1.00%7D+from+0+to+1
		{Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 1.0}, 1.623225},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v", tt.p0, tt.p1, tt.p2), func(t *testing.T) {
			l := quadraticBezierLength(tt.p0, tt.p1, tt.p2)
			test.FloatDiff(t, l, tt.l, 1e-6)
		})
	}
}

func TestQuadraticBezierDistance(t *testing.T) {
	var tests = []struct {
		p0, p1, p2 Point
		q          Point
		d          float64
	}{
		{Point{0.0, 0.0}, Point{4.0, 6.0}, Point{8.0, 0.0}, Point{9.0, 0.5}, math.Sqrt(1.25)},
		{Point{0.0, 0.0}, Point{1.0, 1.0}, Point{2.0, 0.0}, Point{0.0, 0.0}, 0.0},
		{Point{0.0, 0.0}, Point{1.0, 1.0}, Point{2.0, 0.0}, Point{1.0, 1.0}, 0.5},
		{Point{0.0, 0.0}, Point{1.0, 1.0}, Point{2.0, 0.0}, Point{2.0, 0.0}, 0.0},
		{Point{0.0, 0.0}, Point{1.0, 1.0}, Point{2.0, 0.0}, Point{1.0, 0.0}, 0.5},
		{Point{0.0, 0.0}, Point{1.0, 1.0}, Point{2.0, 0.0}, Point{-1.0, 0.0}, 1.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.q), func(t *testing.T) {
			d := quadraticBezierDistance(tt.p0, tt.p1, tt.p2, tt.q)
			test.Float(t, d, tt.d)
		})
	}
}

func TestXMonotoneQuadraticBezier(t *testing.T) {
	test.T(t, xmonotoneQuadraticBezier(Point{2.0, 0.0}, Point{0.0, 1.0}, Point{2.0, 2.0}), MustParseSVGPath("M2 0Q1 0.5 1 1Q1 1.5 2 2"))
}

func TestQuadraticBezierFlatten(t *testing.T) {
	tolerance := 0.1
	tests := []struct {
		path     string
		expected string
	}{
		{"Q1 0 1 1", "L0.8649110641 0.4L1 1"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			path := MustParseSVGPath(tt.path)
			p0 := Point{path.d[1], path.d[2]}
			p1 := Point{path.d[5], path.d[6]}
			p2 := Point{path.d[7], path.d[8]}

			p := flattenQuadraticBezier(p0, p1, p2, tolerance)
			test.T(t, p, MustParseSVGPath(tt.expected))
		})
	}
}

func TestCubicBezierPos(t *testing.T) {
	p0, p1, p2, p3 := Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}
	var tests = []struct {
		p0, p1, p2, p3 Point
		t              float64
		q              Point
	}{
		{p0, p1, p2, p3, 0.0, Point{0.0, 0.0}},
		{p0, p1, p2, p3, 0.5, Point{0.75, 0.25}},
		{p0, p1, p2, p3, 1.0, Point{1.0, 1.0}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.p3, tt.t), func(t *testing.T) {
			q := cubicBezierPos(tt.p0, tt.p1, tt.p2, tt.p3, tt.t)
			test.T(t, q, tt.q)
		})
	}
}

func TestCubicBezierDeriv(t *testing.T) {
	p0, p1, p2, p3 := Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}
	var tests = []struct {
		p0, p1, p2, p3 Point
		t              float64
		q              Point
	}{
		{p0, p1, p2, p3, 0.0, Point{2.0, 0.0}},
		{p0, p1, p2, p3, 0.5, Point{1.0, 1.0}},
		{p0, p1, p2, p3, 1.0, Point{0.0, 2.0}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.p3, tt.t), func(t *testing.T) {
			q := cubicBezierDeriv(tt.p0, tt.p1, tt.p2, tt.p3, tt.t)
			test.T(t, q, tt.q)
		})
	}
}

func TestCubicBezierDeriv2(t *testing.T) {
	p0, p1, p2, p3 := Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}
	var tests = []struct {
		p0, p1, p2, p3 Point
		t              float64
		q              Point
	}{
		{p0, p1, p2, p3, 0.0, Point{-2.0, 2.0}},
		{p0, p1, p2, p3, 0.5, Point{-2.0, 2.0}},
		{p0, p1, p2, p3, 1.0, Point{-2.0, 2.0}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.p3, tt.t), func(t *testing.T) {
			q := cubicBezierDeriv2(tt.p0, tt.p1, tt.p2, tt.p3, tt.t)
			test.T(t, q, tt.q)
		})
	}
}

func TestCubicBezierCurvatureRadius(t *testing.T) {
	p0, p1, p2, p3 := Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}
	var tests = []struct {
		p0, p1, p2, p3 Point
		t              float64
		r              float64
	}{
		{p0, p1, p2, p3, 0.0, 2.0},
		{p0, p1, p2, p3, 0.5, 1.0 / math.Sqrt(2)},
		{p0, p1, p2, p3, 1.0, 2.0},
		{p0, Point{1.0, 0.0}, Point{2.0, 0.0}, Point{3.0, 0.0}, 0.0, math.NaN()},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.p3, tt.t), func(t *testing.T) {
			r := cubicBezierCurvatureRadius(tt.p0, tt.p1, tt.p2, tt.p3, tt.t)
			test.Float(t, r, tt.r)
		})
	}
}

func TestCubicBezierNormal(t *testing.T) {
	p0, p1, p2, p3 := Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}
	var tests = []struct {
		p0, p1, p2, p3 Point
		t              float64
		q              Point
	}{
		{p0, p1, p2, p3, 0.0, Point{0.0, -1.0}},
		{p0, p0, p1, p3, 0.0, Point{0.0, -1.0}},
		{p0, p0, p0, p1, 0.0, Point{0.0, -1.0}},
		{p0, p0, p0, p0, 0.0, Point{0.0, 0.0}},
		{p0, p1, p2, p3, 1.0, Point{1.0, 0.0}},
		{p0, p2, p3, p3, 1.0, Point{1.0, 0.0}},
		{p2, p3, p3, p3, 1.0, Point{1.0, 0.0}},
		{p3, p3, p3, p3, 1.0, Point{0.0, 0.0}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v%v--%v", tt.p0, tt.p1, tt.p2, tt.p3, tt.t), func(t *testing.T) {
			q := cubicBezierNormal(tt.p0, tt.p1, tt.p2, tt.p3, tt.t, 1.0)
			test.T(t, q, tt.q)
		})
	}
}

func TestCubicBezierLength(t *testing.T) {
	p0, p1, p2, p3 := Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}
	var tests = []struct {
		p0, p1, p2, p3 Point
		l              float64
	}{
		// https://www.wolframalpha.com/input/?i=length+of+the+curve+%7Bx%3D3*%281-t%29%5E2*t*0.666667+%2B+3*%281-t%29*t%5E2*1.00+%2B+t%5E3*1.00%2C+y%3D3*%281-t%29*t%5E2*0.333333+%2B+t%5E3*1.00%7D+from+0+to+1
		{p0, p1, p2, p3, 1.623225},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v%v%v%v", tt.p0, tt.p1, tt.p2, tt.p3), func(t *testing.T) {
			l := cubicBezierLength(tt.p0, tt.p1, tt.p2, tt.p3)
			test.FloatDiff(t, l, tt.l, 1e-6)
		})
	}
}

func TestCubicBezierSplit(t *testing.T) {
	p0, p1, p2, p3, q0, q1, q2, q3 := cubicBezierSplit(Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}, 0.5)
	test.T(t, p0, Point{0.0, 0.0})
	test.T(t, p1, Point{1.0 / 3.0, 0.0})
	test.T(t, p2, Point{7.0 / 12.0, 1.0 / 12.0})
	test.T(t, p3, Point{0.75, 0.25})
	test.T(t, q0, Point{0.75, 0.25})
	test.T(t, q1, Point{11.0 / 12.0, 5.0 / 12.0})
	test.T(t, q2, Point{1.0, 2.0 / 3.0})
	test.T(t, q3, Point{1.0, 1.0})
}

func TestCubicBezierStrokeHelpers(t *testing.T) {
	p0, p1, p2, p3 := Point{0.0, 0.0}, Point{2.0 / 3.0, 0.0}, Point{1.0, 1.0 / 3.0}, Point{1.0, 1.0}

	p := &Path{}
	addCubicBezierLine(p, p0, p1, p0, p0, 0.0, 0.5)
	test.That(t, p.Empty())

	p = &Path{}
	addCubicBezierLine(p, p0, p1, p2, p3, 0.0, 0.5)
	test.T(t, p, MustParseSVGPath("L0 -0.5"))

	p = &Path{}
	addCubicBezierLine(p, p0, p1, p2, p3, 1.0, 0.5)
	test.T(t, p, MustParseSVGPath("L1.5 1"))
}

func TestXMonotoneCubicBezier(t *testing.T) {
	test.T(t, xmonotoneCubicBezier(Point{1.0, 0.0}, Point{0.0, 0.0}, Point{0.0, 1.0}, Point{1.0, 1.0}), MustParseSVGPath("M1 0C0.5 0 0.25 0.25 0.25 0.5C0.25 0.75 0.5 1 1 1"))
	test.T(t, xmonotoneCubicBezier(Point{0.0, 0.0}, Point{3.0, 0.0}, Point{-2.0, 1.0}, Point{1.0, 1.0}), MustParseSVGPath("M0 0C0.75 0 1 0.0625 1 0.15625C1 0.34375 0.0 0.65625 0.0 0.84375C0.0 0.9375 0.25 1 1 1"))
}

func TestCubicBezierStrokeFlatten(t *testing.T) {
	tests := []struct {
		path      string
		d         float64
		tolerance float64
		expected  string
	}{
		{"C0.666667 0 1 0.333333 1 1", 0.5, 0.5, "L1.5 1"},
		{"C0.666667 0 1 0.333333 1 1", 0.5, 0.125, "L1.376154 0.308659L1.5 1"},
		{"C1 0 2 1 3 2", 0.0, 0.1, "L1.095445 0.351314L2.579154 1.581915L3 2"},
		{"C0 0 1 0 2 2", 0.0, 0.1, "L1.22865 0.8L2 2"},       // p0 == p1
		{"C1 1 2 2 3 5", 0.0, 0.1, "L2.481111 3.612482L3 5"}, // s2 == 0
	}
	origEpsilon := Epsilon
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			Epsilon = origEpsilon
			path := MustParseSVGPath(tt.path)
			p0 := Point{path.d[1], path.d[2]}
			p1 := Point{path.d[5], path.d[6]}
			p2 := Point{path.d[7], path.d[8]}
			p3 := Point{path.d[9], path.d[10]}

			p := &Path{}
			flattenSmoothCubicBezier(p, p0, p1, p2, p3, tt.d, tt.tolerance)
			Epsilon = 1e-6
			test.T(t, p, MustParseSVGPath(tt.expected))
		})
	}
	Epsilon = origEpsilon
}

func TestCubicBezierInflectionPoints(t *testing.T) {
	tests := []struct {
		p0, p1, p2, p3 Point
		x1, x2         float64
	}{
		{Point{0.0, 0.0}, Point{0.0, 1.0}, Point{1.0, 1.0}, Point{1.0, 0.0}, math.NaN(), math.NaN()},
		{Point{0.0, 0.0}, Point{1.0, 1.0}, Point{0.0, 1.0}, Point{1.0, 0.0}, 0.5, math.NaN()},

		// see "Analysis of Inflection Points for Planar Cubic Bezier Curve" by Z.Zhang et al. from 2009
		// https://cie.nwsuaf.edu.cn/docs/20170614173651207557.pdf
		{Point{16, 467}, Point{185, 95}, Point{673, 545}, Point{810, 17}, 0.4565900353, math.NaN()},
		{Point{859, 676}, Point{13, 422}, Point{781, 12}, Point{266, 425}, 0.6810755245, 0.7052992723},
		{Point{872, 686}, Point{11, 423}, Point{779, 13}, Point{220, 376}, 0.5880709424, 0.8868629954},
		{Point{819, 566}, Point{43, 18}, Point{826, 18}, Point{25, 533}, 0.4761686269, 0.5392953369},
		{Point{884, 574}, Point{135, 14}, Point{678, 14}, Point{14, 566}, 0.3208363269, 0.6822908688},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v %v %v %v", tt.p0, tt.p1, tt.p2, tt.p3), func(t *testing.T) {
			x1, x2 := findInflectionPointsCubicBezier(tt.p0, tt.p1, tt.p2, tt.p3)
			test.Floats(t, []float64{x1, x2}, []float64{tt.x1, tt.x2})
		})
	}
}

func TestCubicBezierInflectionPointRange(t *testing.T) {
	tests := []struct {
		p0, p1, p2, p3 Point
		t, tolerance   float64
		x1, x2         float64
	}{
		{Point{0.0, 0.0}, Point{1.0, 1.0}, Point{0.0, 1.0}, Point{1.0, 0.0}, math.NaN(), 0.25, math.Inf(1.0), math.Inf(1.0)},

		// p0==p1==p2
		{Point{0.0, 0.0}, Point{0.0, 0.0}, Point{0.0, 0.0}, Point{1.0, 0.0}, 0.0, 0.25, 0.0, 1.0},

		// p0==p1, s3==0
		{Point{0.0, 0.0}, Point{0.0, 0.0}, Point{1.0, 0.0}, Point{1.0, 0.0}, 0.0, 0.25, 0.0, 1.0},

		// all within tolerance
		{Point{0.0, 0.0}, Point{0.0, 1.0}, Point{1.0, 1.0}, Point{1.0, 0.0}, 0.5, 1.0, -0.0503212081, 1.0503212081},
		{Point{0.0, 0.0}, Point{0.0, 1.0}, Point{1.0, 1.0}, Point{1.0, 0.0}, 0.5, 1e-9, 0.4994496788, 0.5005503212},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v %v %v %v", tt.p0, tt.p1, tt.p2, tt.p3), func(t *testing.T) {
			x1, x2 := findInflectionPointRangeCubicBezier(tt.p0, tt.p1, tt.p2, tt.p3, tt.t, tt.tolerance)
			test.Floats(t, []float64{x1, x2}, []float64{tt.x1, tt.x2})
		})
	}
}

func TestCubicBezierStroke(t *testing.T) {
	tests := []struct {
		p []Point
	}{
		// see "Analysis of Inflection Points for Planar Cubic Bezier Curve" by Z.Zhang et al. from 2009
		// https://cie.nwsuaf.edu.cn/docs/20170614173651207557.pdf
		{[]Point{{16, 467}, {185, 95}, {673, 545}, {810, 17}}},
		{[]Point{{859, 676}, {13, 422}, {781, 12}, {266, 425}}},
		{[]Point{{872, 686}, {11, 423}, {779, 13}, {220, 376}}},
		{[]Point{{819, 566}, {43, 18}, {826, 18}, {25, 533}}},
		{[]Point{{884, 574}, {135, 14}, {678, 14}, {14, 566}}},

		// be aware that we offset the bezier by 0.1
		// single inflection point, ranges outside t=[0,1]
		{[]Point{{0, 0}, {1, 1}, {0, 1}, {1, 0}}},

		// two inflection points, ranges outside t=[0,1]
		{[]Point{{0, 0}, {0.9, 1}, {0.1, 1}, {1, 0}}},

		// one inflection point, max range outside t=[0,1]
		{[]Point{{0, 0}, {80, 100}, {80, -100}, {100, 0}}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v %v %v %v", tt.p[0], tt.p[1], tt.p[2], tt.p[3]), func(t *testing.T) {
			length := cubicBezierLength(tt.p[0], tt.p[1], tt.p[2], tt.p[3])
			flatLength := strokeCubicBezier(tt.p[0], tt.p[1], tt.p[2], tt.p[3], 0.0, 0.001).Length()
			test.FloatDiff(t, flatLength, length, 0.25)
		})
	}

	test.T(t, strokeCubicBezier(Point{0, 0}, Point{30, 0}, Point{30, 10}, Point{25, 10}, 5.0, 0.01).Bounds(), Rect{0.0, -5.0, 32.4787516156, 20.0})
}
