package ai

import (
	"math"
	"sort"
)

// Simplified game info for 3v3 (6 ships, thus "hex") battle sims

type HexSim struct {					// Using pointers, unlike in most of the code
	planets			[]*SimPlanet
	all_ships		[]*SimShip
}

type SimPlanet struct {
	x				float64
	y				float64
	radius			float64
}

type ShipState int

const (
	ALIVE ShipState = iota
	DYING
	DEAD
)

type WeaponState int

const (
	READY WeaponState = iota
	FIRING
	SPENT
)

type SimShip struct {
	x				float64
	y				float64
	vel_x			float64
	vel_y			float64
	ship_state		ShipState
	weapon_state	WeaponState
	actual_targets	[]*SimShip		// Who we actually, really, definitely shoot at.
	hp				int
	owner			int
}

type EventType int

const (
	ATTACK EventType = iota
	SHIP_COLLISION
)

type PossibleEvent struct {
	ship_a		*SimShip
	ship_b		*SimShip
	t			float64
	what		EventType
}

func min(a, b float64) float64 {
	if a < b { return a }
	return b
}

func max(a, b float64) float64 {
	if a > b { return a }
	return b
}

func collision_time(r float64, ship1 * SimShip, ship2 * SimShip) (bool, float64) {

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

	dx := ship1.x - ship2.x
	dy := ship1.y - ship2.y
	dvx := ship1.vel_x - ship2.vel_x
	dvy := ship1.vel_y - ship2.vel_y

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

func (self *HexSim) Step() {

	var possible_events []*PossibleEvent

	// Attacks...

	for i, ship_a := range self.all_ships {
		for _, ship_b := range self.all_ships[i+1:] {
			if ship_a.owner != ship_b.owner {
				ok, t := collision_time(5.0, ship_a, ship_b)		// 5.0 seems to be right. Uh, but see #191.
				if ok && t >= 0 && t <= 1 {
					possible_events = append(possible_events, &PossibleEvent{ship_a, ship_b, t, ATTACK})
				}
			}
		}
	}

	// Collisions...

	for i, ship_a := range self.all_ships {
		for _, ship_b := range self.all_ships[i+1:] {
			ok, t := collision_time(1.0, ship_a, ship_b)
			if ok {
				possible_events = append(possible_events, &PossibleEvent{ship_a, ship_b, t, ATTACK})
			}
		}
	}

	// We want events sorted into groups of simultaneous events.

	sort.Slice(possible_events, func(a, b int) bool {
		return possible_events[a].t < possible_events[b].t
	})

	var grouped_events [][]*PossibleEvent

	current_t := -1.0

	for _, event := range possible_events {

		t := event.t
		if t > current_t {
			current_t = t
			grouped_events = append(grouped_events, nil)
		}

		grouped_events[len(grouped_events) - 1] = append(grouped_events[len(grouped_events) - 1], event)
	}

	// Now update ships...

	for _, grouping := range grouped_events {

		// Complete state transitions for weapon and alive/dead...

		for _, ship := range self.all_ships {

			if ship.weapon_state == FIRING {
				ship.weapon_state = SPENT
			}

			if ship.hp <= 0 {
				ship.ship_state = DEAD
			}
		}

		// Now go through each event and update stuff if the conditions are right...
		// Note that having 0 HP is not sufficient to be irrelevant, because we may
		// have died this very moment. Instead, check ship_state == DEAD.

		for _, event := range grouping {

			ship_a, ship_b := event.ship_a, event.ship_b

			if ship_a.ship_state == DEAD || ship_b.ship_state == DEAD {
				continue
			}

			if event.what == SHIP_COLLISION {

				ship_a.hp = 0
				ship_a.ship_state = DYING
				ship_b.hp = 0
				ship_b.ship_state = DYING

			} else if event.what == ATTACK {

				if ship_a.weapon_state != SPENT {
					ship_a.weapon_state = FIRING
					ship_a.actual_targets = append(ship_a.actual_targets, ship_b)
				}

				if ship_b.weapon_state != SPENT {
					ship_b.weapon_state = FIRING
					ship_b.actual_targets = append(ship_b.actual_targets, ship_a)
				}
			}
		}

		// Apply weapon damage...

		for _, ship := range self.all_ships {
			if len(ship.actual_targets) > 0 {
				damage := 64 / len(ship.actual_targets)				// Right? A straight up integer truncation?
				for _, target := range ship.actual_targets {
					target.hp -= damage
				}
			}
		}
	}
}

