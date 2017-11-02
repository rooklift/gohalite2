package ai

import (
	"math"
)

func min(a, b float64) float64 {
	if a < b { return a }
	return b
}

func max(a, b float64) float64 {
	if a > b { return a }
	return b
}

func collision_time(r float64, e1 * SimEntity, e2 * SimEntity) (bool, float64) {

	// https://github.com/HaliteChallenge/Halite-II/blob/master/environment/core/SimulationEvent.cpp#L100
	//
	// With credit to Ben Spector
	// Simplified derivation:
	// 1. Set up the distance between the two entities in terms of time,
	//    the difference between their velocities and the difference between
	//    their positions
	// 2. Equate the distance equal to the event radius (max possible distance
	//    they could be)
	// 3. Solve the resulting quadratic

	dx := e1.x - e2.x
	dy := e1.y - e2.y
	dvx := e1.vel_x - e2.vel_x
	dvy := e1.vel_y - e2.vel_y

	// Quadratic formula
	a := dvx * dvx + dvy * dvy				// const auto a = std::pow(dvx, 2) + std::pow(dvy, 2);
	b := 2 * (dx * dvx + dy * dvy)			// const auto b = 2 * (dx * dvx + dy * dvy);
	c := dx * dx + dy * dy - r * r			// const auto c = std::pow(dx, 2) + std::pow(dy, 2) - std::pow(r, 2);

	disc := b * b - 4 * a * c				// disc := std::pow(b, 2) - 4 * a * c;

	if (a == 0.0) {
		if (b == 0.0) {
			if (c <= 0.0) {
				// Implies r^2 >= dx^2 + dy^2 and the two are already colliding
				return true, 0.0
			}
			return false, 0.0
		}
		t := -c / b
		if (t >= 0.0) {
			return true, t
		}
		return false, 0.0
	} else if (disc == 0.0) {
		// One solution
		t := -b / (2 * a)
		return true, t
	} else if (disc > 0) {
		t1 := -b + math.Sqrt(disc)
		t2 := -b - math.Sqrt(disc)

		if (t1 >= 0.0 && t2 >= 0.0) {
			return true, min(t1, t2) / (2 * a)
		} else {
			return true, max(t1, t2) / (2 * a)
		}
	} else {
		return false, 0.0
	}
}
