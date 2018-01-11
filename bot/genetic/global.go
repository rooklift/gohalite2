package genetic

import (
	"sort"

	hal "../core"
)

func EvolveGlobal(game *hal.Game) {

	my_mutable_ship_map := make(map[int]*hal.Ship)
	relevant_enemy_map := make(map[int]*hal.Ship)
	my_immutable_ship_map := make(map[int]*hal.Ship)

	my_ships := game.MyShips()
	enemy_ships := game.EnemyShips()

	for _, ship := range my_ships {
		if ship.DockedStatus == hal.UNDOCKED {
			for _, enemy := range enemy_ships {
				if ship.Dist(enemy) < 20 {
					my_mutable_ship_map[ship.Id] = ship
					relevant_enemy_map[enemy.Id] = enemy
				}
			}
		}
	}

	if len(my_mutable_ship_map) == 0 {
		return
	}

	for _, ship := range my_ships {
		if my_mutable_ship_map[ship.Id] == nil {
			for _, other := range my_mutable_ship_map {
				if ship.Dist(other) < 8 {
					my_immutable_ship_map[ship.Id] = ship
				}
			}
		}
	}

	var relevant_ships []*hal.Ship

	for _, ship := range my_mutable_ship_map {
		relevant_ships = append(relevant_ships, ship)
	}

	for _, ship := range my_immutable_ship_map {
		relevant_ships = append(relevant_ships, ship)
	}

	// We keep our own ships sorted by ID, I forget if this is really needed.

	sort.Slice(relevant_ships, func(a, b int) bool {
		return relevant_ships[a].Id < relevant_ships[b].Id
	})

	var mutable_ship_ordinals []int		// Index locations in relevant_ships[].

	for i, ship := range relevant_ships {
		if my_mutable_ship_map[ship.Id] != nil {
			mutable_ship_ordinals = append(mutable_ship_ordinals, i)
		}
	}

	// Enemies can go in whatever order...

	for _, enemy := range relevant_enemy_map {
		relevant_ships = append(relevant_ships, enemy)
	}

	// We construct our genome in this function - if we want to (with a bit of extra work)
	// we can start our genome with some moves produced in the conventional way.

	genome := new(Genome)
	genome.Init(len(my_mutable_ship_map), false)

	GlobalEvolver(game, genome, relevant_ships, mutable_ship_ordinals, 10000)

	// Once that's done, assign moves to the ships, BEARING IN MIND THAT THE GENOME ONLY COVERS RELEVANT SHIPS
	// (unlike in the original GA where it covered all our ships).
}


func GlobalEvolver(game *hal.Game, cold_genome *Genome, relevant_ships []*hal.Ship, mutable_ship_ordinals []int, iterations int) {

	initial_sim := SetupSim(game, -1, relevant_ships)

	sim_without_enemies := initial_sim.Copy()
	for i := 0; i < len(sim_without_enemies.ships); i++ {
		if sim_without_enemies.ships[i].owner != game.Pid() {
			sim_without_enemies.ships = append(sim_without_enemies.ships[:i], sim_without_enemies.ships[i+1:]...)
			i--
		}
	}

	for n := 0; n < iterations; n++ {

		genome := cold_genome.Copy()

		if n > 0 {
			genome.Mutate(mutable_ship_ordinals)
		}

		genome.score = 0

		for scenario := 0; scenario < 3; scenario++ {

			var sim *Sim

			if scenario == 0 {
				sim = sim_without_enemies.Copy()
			} else {
				sim = initial_sim.Copy()
			}

			for g, i := range mutable_ship_ordinals {

				speed := genome.genes[g].speed
				angle := genome.genes[g].angle
				vel_x, vel_y := hal.Projection(0, 0, float64(speed), angle)
				sim.ships[i].vel_x = vel_x
				sim.ships[i].vel_y = vel_y

			}

			for _, ship := range sim.ships {

				if ship.owner != game.Pid() && ship.dockedstatus != hal.UNDOCKED {

					switch scenario {

					case 0:
						// Scenario 0 is the enemy ships not existing at at all (so we don't hit planets, etc)

					case 1:
						real_ship, _ := game.GetShip(ship.id)		// FIXME: could speed-optimise this
						ship.vel_x = real_ship.Dx
						ship.vel_y = real_ship.Dy

					case 2:
						// ship.vel_x = 0											// Already so.
						// ship.vel_y = 0

					}
				}
			}

			sim.Step()

			// Now score it according to damage and stupidity...

		}
	}

	return
}

