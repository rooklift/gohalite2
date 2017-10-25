package gohalite2

func (self *Game) AngleCollisionID(startx, starty, distance float64, degrees int) int {

	// Returns ID of the first planet we would collide with, or -1 if no hits

	var collision_planets []Planet
	all_planets := self.AllPlanets()

	endx, endy := projection(startx, starty, distance, degrees)

	for _, planet := range all_planets {

		collides := intersect_segment_circle(startx, starty, endx, endy, planet.X, planet.Y, planet.Radius)

		if collides {
			collision_planets = append(collision_planets, planet)
		}
	}

	if len(collision_planets) == 0 {
		return -1
	}

	closest_distance := 999999.9
	closest_planet := Planet{}

	for _, c := range collision_planets {
		if dist(startx, starty, c.X, c.Y) < closest_distance {
			closest_planet = c
		}
	}

	return closest_planet.Id
}
