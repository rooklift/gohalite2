package genetic

/*

import (
	"sort"
	"time"

	hal "../core"
	pil "../pilot"
)

func EvolveGlobal(game *hal.Game) {

	my_mutable_ship_map := make(map[int]*hal.Ship)
	my_immutable_ship_map := make(map[int]*hal.Ship)
	relevant_enemy_map := make(map[int]*hal.Ship)

	my_ships := game.MyShips()
	enemy_ships := game.EnemyShips()

	// Make maps of my ships that are near enemies, and vice versa...

	for _, ship := range my_ships {
		if ship.DockedStatus == hal.UNDOCKED {
			if hal.GetOrderType(game.CurrentOrder(ship)) == "t" || hal.GetOrderType(game.CurrentOrder(ship)) == "" {
				for _, enemy := range enemy_ships {
					if ship.Dist(enemy) < 20 {
						relevant_enemy_map[enemy.Id] = enemy
						if ship.Doomed == false {
							my_mutable_ship_map[ship.Id] = ship
						} else {
							my_immutable_ship_map[ship.Id] = ship
						}
					}
				}
			}
		}
	}

	if len(my_mutable_ship_map) == 0 {
		return
	}

	// Make map of my ships that aren't near enemies, but which we could collide into...

	for _, ship := range my_ships {
		if my_mutable_ship_map[ship.Id] == nil {
			for _, other := range my_mutable_ship_map {
				if ship.Dist(other) < 8 {
					my_immutable_ship_map[ship.Id] = ship
				}
			}
		}
	}

	// Convert maps to slices...

	var my_mutable_ships []*hal.Ship
	var my_immutable_ships []*hal.Ship
	var relevant_enemy_ships []*hal.Ship

	for _, ship := range my_mutable_ship_map {
		my_mutable_ships = append(my_mutable_ships, ship)
	}

	for _, ship := range my_immutable_ship_map {
		my_immutable_ships = append(my_immutable_ships, ship)
	}

	for _, enemy := range relevant_enemy_map {
		relevant_enemy_ships = append(relevant_enemy_ships, enemy)
	}

	// Sort everything by ID for determinism purposes. (Since we iterated over a map.)

	sort.Slice(my_mutable_ships, func(a, b int) bool {
		return my_mutable_ships[a].Id < my_mutable_ships[b].Id
	})

	sort.Slice(my_immutable_ships, func(a, b int) bool {
		return my_immutable_ships[a].Id < my_immutable_ships[b].Id
	})

	sort.Slice(relevant_enemy_ships, func(a, b int) bool {
		return relevant_enemy_ships[a].Id < relevant_enemy_ships[b].Id
	})

	// Set up and run evolver...

	evolver := NewEvolver(game, my_mutable_ships, my_immutable_ships, relevant_enemy_ships, 1)

	for i, gene := range evolver.genomes[0].genes {
		ship := my_mutable_ships[i]
		planned_speed, planned_angle := hal.CourseFromString(game.CurrentOrder(ship))
		gene.speed = planned_speed
		gene.angle = planned_angle
	}

	evolver.RunGlobalFight(10000)
	evolver.ExecuteGenome(pil.MSG_GLOBAL_SAUCE)

	game.Log("EvolveGlobal() lens: %v, %v, %v", len(my_mutable_ships), len(my_immutable_ships), len(relevant_enemy_ships))
}

func (self *Evolver) RunGlobalFight(iterations int) {

	width, height := float64(self.game.Width()), float64(self.game.Height())
	pid := self.game.Pid()

	var real_enemy_ships []*hal.Ship
	for i := self.first_enemy_index; i < len(self.sim.ships); i++ {
		real_enemy_ship, _ := self.game.GetShip(self.sim.ships[i].id)
		real_enemy_ships = append(real_enemy_ships, real_enemy_ship)
	}

	self.iterations_required = 0
	best_score := -2147483647

	genome_backup := new(Genome)
	genome_backup.Init(self.genome_length, false)

	for n := 0; n < iterations; n++ {

		for c := 0; c < len(self.genomes); c++ {

			genome := self.genomes[c]

			genome_backup.score = genome.score
			for i := 0; i < len(genome.genes); i++ {
				*genome_backup.genes[i] = *genome.genes[i]
			}

			if n > 0 {						// Don't mutate first iteration, so we get a true score for our initial genome. Makes the results stable.
				genome.Mutate()
			}

			genome.score = 0

			// We run some different scenarios of what the enemy will do.

			for scenario := 0; scenario < 3; scenario++ {

				var sim *Sim

				if scenario == 0 {						// Scenario 0 is the enemy ships not existing at at all (so we don't hit planets, etc)
					sim = self.sim_without_enemies
				} else {
					sim = self.sim
				}

				sim.Reset()								// We used to make a copy of the sim, but that was slower. Now just reset every time.

				my_mutable_simships := sim.ships[0:self.genome_length]

				for i := 0; i < len(genome.genes); i++ {

					if sim.ships[i].dockedstatus == hal.UNDOCKED {		// This really should be true

						speed := genome.genes[i].speed
						angle := genome.genes[i].angle

						vel_x, vel_y := hal.Projection(0, 0, float64(speed), angle)

						sim.ships[i].vel_x = vel_x						// Relying on our mutable
						sim.ships[i].vel_y = vel_y						// ships being stored first.

					} else {
						panic("RunRushFight(): got docked ship where mutable ship should be")
					}
				}

				for i := len(genome.genes); i < len(sim.ships); i++ {

					if sim.ships[i].dockedstatus == hal.UNDOCKED {

						if sim.ships[i].owner != pid {

							switch scenario {

							case 0:
								// Scenario 0 is the enemy ships not existing at at all (so we don't hit planets, etc)
								panic("RunRushFight(): got enemy ship in scenario 0")

							case 1:
								sim.ships[i].vel_x = sim.ships[i].real_ship.Dx
								sim.ships[i].vel_y = sim.ships[i].real_ship.Dy

							case 2:
								// Scenario 2 is the enemy ships making no move

							}
						}
					}
				}

				sim.Step()

				// SCORING -----------------------------------------------------------------------------------------------------------

				// Damage...

				for _, ship := range sim.ships {
					if ship.hp > 0 {
						if ship.owner != pid {
							genome.score -= ship.hp * 100
						} else {
							genome.score += ship.hp * 100
						}
					}
				}

				// Other scores only need to be run in one scenario to work...

				if scenario == 2 {

					// A good scenario to run our stupidity checks in. In particular, enemy docked ships exist
					// in this scenario and our ships are flagged as stupid if they have collided with them.

					for _, ship := range my_mutable_simships {
						if ship.stupid_death || ship.x <= 0 || ship.x >= width || ship.y <= 0 || ship.y >= height {
							genome.score -= 9999999
						}
					}
				}
			}

			if n == 0 && c == 0 {
				self.null_score = genome.score		// Record the score of not moving. Relies on no mutation in n0 and non-randomised c0.
			}

			if float64(genome.score) <= float64(genome_backup.score) * thresholds[c] {

				// Reset the genome to how it was.

				genome.score = genome_backup.score
				for i := 0; i < len(genome.genes); i++ {
					*genome.genes[i] = *genome_backup.genes[i]
				}
			}
		}

		score_0 := self.genomes[0].score

		sort.SliceStable(self.genomes, func(a, b int) bool {
			return self.genomes[a].score > self.genomes[b].score		// Note the reversed sort, high scores come first.
		})

		if self.genomes[0].score > score_0 {
			self.cold_swaps++
		}

		if self.genomes[0].score > best_score {
			self.iterations_required = n								// info only.
			best_score = self.genomes[0].score
		}

		if time.Now().Sub(self.game.ParseTime()) > 1500 * time.Millisecond {
			self.game.Log("Emergency timeout in RunGlobalFight() after %d iterations.", n)
			return
		}
	}

	return
}

*/
