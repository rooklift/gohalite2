package opening

/*

import (
	hal "../core"
)

// Here we simulate what happens when we send our initial ships some way on turn zero.
// There's no reason why we can't do a full resolution sim that's perfect, but I can't
// be bothered. This is a crude approximation.

type OS_Planet struct {
	x				float64
	y				float64
	id				int
	dockingspots	int
	production		int
	ships			[]int			// List of ETA on incoming ships
}

func (self *OS_Planet) Dist(other *OS_Planet) float64 {
	return hal.Dist(self.x, self.y, other.x, other.y)
}

func RunOpeningSim(game *hal.Game, initial []int) int {			// Arguments are planet IDs to start with.

	var planets []*OS_Planet

	for _, planet := range game.AllPlanets() {

		planets = append(planets, &OS_Planet{
			x: planet.X,
			y: planet.Y,
			id: planet.Id,
			dockingspots: planet.DockingSpots,
			production: 0,
			ships: nil,
		})
	}

	my_cog := game.MyShipsCentreOfGravity()

	for _, plid := range initial {
		for _, planet := range planets {
			if planet.id == plid {
				d := hal.Dist(my_cog.X, my_cog.Y, planet.x, planet.y)
				eta := d / 7 + 5
				planet.ships = append(planet.ships, int(eta))
				break
			}
		}
	}

	// --------------------------------------------

	timer := 0

	for {

		timer++
		total_docked := 0

		for _, planet := range planets {

			for i, eta := range planet.ships {
				if eta > 0 {
					planet.ships[i] -= 1
				}
				if eta == 0 {
					planet.production += 6
					total_docked++
				}
			}

			if planet.production >= 72 {

				planet.production -= 72

				var best_target *OS_Planet
				var best_dist float64

				for _, other := range planets {		// Other might actually be self

					if other.dockingspots - len(other.ships) > 0 {

						d := planet.Dist(other)
						if best_target == nil || d < best_dist {

							best_target = other
							best_dist = d
						}
					}
				}

				if best_target == nil {
					return timer
				} else {
					eta := int(best_dist / 7 + 5)
					best_target.ships = append(best_target.ships, eta)
				}
			}
		}

		if total_docked > 6 {
			return timer
		}
	}
}

*/
