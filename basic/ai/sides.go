package ai

const (
	LEFT Side = iota
	RIGHT
)

type Side int

func (s Side) String() string {
	if s == 0 { return "LEFT" } else if s == 1 { return "RIGHT" } else { return "???" }
}

// Given a ship and some target, and some blocker to navigate around,
// which side should we go?

func DecideSide(ship *Ship, target Entity, blocker Entity) Side {

	to_blocker := ship.Angle(blocker)
	to_target := ship.Angle(target)

	diff := to_blocker - to_target

	var side Side = RIGHT

	if diff >= 0 && diff <= 180 {
		side = LEFT
	} else if diff >= 180 {
		side = RIGHT
	} else if diff <= -180 {
		side = LEFT
	} else if diff >= -180 && diff <= 0 {
		side = RIGHT
	}

	return side
}

func DecideSideFromTarget(ship *Ship, target Entity, game *Game) Side {

	// If the first planet in our path isn't our target, we choose a side to navigate around.
	// Docked ships also count as part of the planet for these purposes.

	side := ArbitrarySide(ship)

	if target.Type() == NOTHING {
		return side
	}

	// By using AllImmobile() as the avoid_list, any collision will be with a planet or docked ship.

	collision_entity, ok := FirstCollision(ship, 1000, ship.Angle(target), game.AllImmobile())

	if ok {

		var blocking_planet *Planet

		if collision_entity.Type() == PLANET {
			blocking_planet = collision_entity.(*Planet)
		} else {
			s := collision_entity.(*Ship)
			blocking_planet, _ = game.GetPlanet(s.DockedPlanet)
		}

		if target.Type() != PLANET || blocking_planet.Id != target.GetId() {
			side = DecideSide(ship, target, blocking_planet)
		}
	}

	return side
}

func ArbitrarySide(ship *Ship) Side {
	if ship.Id % 2 == 0 { return RIGHT }
	return LEFT
}
