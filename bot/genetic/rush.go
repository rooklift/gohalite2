package genetic

import (
	"sort"
	"time"

	hal "../core"
	pil "../pilot"
)

func (self *Evolver) RunRushFight(iterations int, play_perfect bool) {

	const (
		PANIC_RANGE = 30		// How far the enemy can get before we worry
	)

	width, height := float64(self.game.Width()), float64(self.game.Height())
	pid := self.game.Pid()

	var real_enemy_ships []*hal.Ship
	for i := self.first_enemy_index; i < len(self.sim.ships); i++ {
		real_enemy_ship, _ := self.game.GetShip(self.sim.ships[i].id)
		real_enemy_ships = append(real_enemy_ships, real_enemy_ship)
	}

	self.iterations_required = 0
	best_score := -2147483647

	// We used to make a copy of each genome every time we modified it; but now we save and rollback,
	// which is faster. Here's the storage space to do that with:

	genome_backup := new(Genome)
	genome_backup.Init(self.genome_length, false)

	for n := 0; n < iterations; n++ {

		// We run various chains of evolution with different "heats" (i.e. how willing we are to accept bad mutations)
		// in the "metropolis coupling" fashion.

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

				if scenario == 0 {

					// A good scenario to run every other check in.
					// Note that enemy_sim_ship_ptrs is empty here, so use real ships...

					// EDGES OF SPACE / PLANET AVOIDANCE -----------------------------------------------------------------------------

					for _, ship := range my_mutable_simships {

						// Modest penalty for getting near edge of space...

						horiz_clearance := hal.MinFloat(ship.x, width - ship.x)
						vert_clearance := hal.MinFloat(ship.y, height - ship.y)

						if horiz_clearance < 12.5 {
							genome.score -= int(1000.0 - horiz_clearance * 20)		// Needs to be able to override get-close-to-ship reward.
						}
						if vert_clearance < 12.5 {
							genome.score -= int(1000.0 - vert_clearance * 20)
						}

						// Getting really near planets is like death...

						for _, planet := range sim.planets {

							clearance := hal.Dist(ship.x, ship.y, planet.x, planet.y) - (planet.radius + 0.5)

							if clearance < 0.5 {
								genome.score -= (500000 - int(clearance * 20))		// Amusing subtraction but should be effective.
							}
						}
					}

					// DISTANCE ------------------------------------------------------------------------------------------------------

					// Keep close to enemy. Deal with split enemies. The important cases are 3v3, 3v2, and 2v2.
					// I tried writing general stuff but it was simpler just to handle the individual cases.

					if len(genome.genes) == 3 && len(real_enemy_ships) >= 2 {

						// Handles both 3v2 and 3v3.

						dist0 := 999999.9
						dist1 := 999999.9
						dist2 := 999999.9

						permutations := chase_permutations_3v2;
						if len(real_enemy_ships) > 2 {
							permutations = chase_permutations_3v3;
						}

						for _, perm := range permutations {

							this_dist0 := sim.ships[0].Dist(real_enemy_ships[perm[0]])
							this_dist1 := sim.ships[1].Dist(real_enemy_ships[perm[1]])
							this_dist2 := sim.ships[2].Dist(real_enemy_ships[perm[2]])

							if  (this_dist0 + this_dist1 + this_dist2)   <   (dist0 + dist1 + dist2)  {

								dist0, dist1, dist2 = this_dist0, this_dist1, this_dist2

							}
						}

						if dist0 < PANIC_RANGE {
							genome.score -= int(dist0 * 9)
						} else {
							genome.score -= int(dist0 * 9000)
						}

						if dist1 < PANIC_RANGE {
							genome.score -= int(dist1 * 9)
						} else {
							genome.score -= int(dist1 * 9000)
						}

						if dist2 < PANIC_RANGE {
							genome.score -= int(dist2 * 9)
						} else {
							genome.score -= int(dist2 * 9000)
						}

					} else if len(genome.genes) == 2 && len(real_enemy_ships) == 2 {

						dist0 := 999999.9
						dist1 := 999999.9

						for _, perm := range chase_permutations_2v2 {

							this_dist0 := sim.ships[0].Dist(real_enemy_ships[perm[0]])
							this_dist1 := sim.ships[1].Dist(real_enemy_ships[perm[1]])

							if  (this_dist0 + this_dist1)   <   (dist0 + dist1)  {

								dist0, dist1 = this_dist0, this_dist1

							}
						}

						if dist0 < PANIC_RANGE {
							genome.score -= int(dist0 * 9)
						} else {
							genome.score -= int(dist0 * 9000)
						}

						if dist1 < PANIC_RANGE {
							genome.score -= int(dist1 * 9)
						} else {
							genome.score -= int(dist1 * 9000)
						}

					} else {

						// Minimise the biggest distances in other cases...

						// Use a small, overridable score, unless the distance is > 40
						// in which case use a massive all-encompassing score.

						highest_enemy_clearance := -1.0

						for _, enemy := range real_enemy_ships {

							closest_range := 999999.9

							for _, ship := range my_mutable_simships {
								d := ship.Dist(enemy)
								if d < closest_range {
									closest_range = d
								}
							}

							if closest_range > highest_enemy_clearance {
								highest_enemy_clearance = closest_range
							}
						}

						highest_friendly_clearance := -1.0

						for _, ship := range my_mutable_simships {

							closest_range := 999999.9

							for _, enemy := range real_enemy_ships {
								d := ship.Dist(enemy)
								if d < closest_range {
									closest_range = d
								}
							}

							if closest_range > highest_friendly_clearance {
								highest_friendly_clearance = closest_range
							}

							// While we're at it, make sure the ship wants to move nearer to some enemy.
							// Otherwise, it might stand still if it's not affecting the clearances.

							genome.score -= int(closest_range * 2)
						}

						if highest_enemy_clearance < 40 {
							genome.score -= int(highest_enemy_clearance * 9)		// Use different numbers such that this can override...
						} else {
							genome.score -= int(highest_enemy_clearance * 9000)
						}

						if highest_friendly_clearance < 40 {
							genome.score -= int(highest_friendly_clearance * 6)		// ...the desire to approach the nearest enemy if need be.
						} else {
							genome.score -= int(highest_friendly_clearance * 6000)
						}
					}

					// PERFECT THIRTEEN RANGE TRICK ----------------------------------------------------------------------------------

					var good_thirteens = make(map[int]int)

					for _, ship := range my_mutable_simships {

						if play_perfect {

							// In "perfect" mode we give huge bonuses to moves that can only ever be hit by 1 enemy;
							// which means being < 13 away from the *starting* location of 1 enemy.

							var thirteens	[]int										// IDs of ships that might be able to hit us.
							var twelves		[]int										// As above, but with some tolerance.
							var eights		[]int										// IDs of ships that might be able to ram us.

							for _, enemy_ship := range real_enemy_ships {				// Must use real_enemy_ships, since sim enemies aren't present.

								if enemy_ship.Doomed {
									continue				// No need to worry about getting to the right distance away from doomed ships.
								}

								if ship.Dist(enemy_ship) < 13 {
									thirteens = append(thirteens, enemy_ship.Id)
								}
								if ship.Dist(enemy_ship) < 12 {
									twelves = append(twelves, enemy_ship.Id)
								}
								if ship.Dist(enemy_ship) < 8 {
									eights = append(eights, enemy_ship.Id)
								}
							}

							if len(thirteens) == 1 && ship.fires_at_time_0 == false {
								genome.score += 100000
								enemy_ship_id := thirteens[0]
								good_thirteens[enemy_ship_id] += 1
							}

							ideal_thirteens := 1
							if ship.fires_at_time_0 {		// If we're already committed to shooting (because a target's in range
								ideal_thirteens = 0			// already) then we should just back away from everything if we can.
							}

							if len(thirteens) > ideal_thirteens {
								genome.score -= 100000 * (len(thirteens) - ideal_thirteens)
							}

							if len(twelves) > 1 {			// We have this in case we just can't find a way to avoid > 2 thirteens.
								genome.score -= 200000		// In which case we need to punish it more if it goes even worse.
							}

							if len(eights) > 0 {			// Note > 0. This stops us getting accidentally rammed when enemy ship is solo.
								genome.score -= 300000
							}
						}
					}

					// Modest bonus for coordinated thirteens (should be enough)

					for _, hits := range good_thirteens {
						genome.score += (hits - 1) * 15000
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
			self.game.Log("Emergency timeout in RunRushFight() after %d iterations.", n)
			return
		}
	}
}

func FightRush2(game *hal.Game, enemy_pid int, play_perfect bool) {

	game.LogOnce("Entering FightRush2() genetic algorithm!")

	var my_mutable_ships []*hal.Ship
	var my_immutable_ships []*hal.Ship
	var enemy_ships []*hal.Ship

	for _, ship := range game.AllShips() {
		if ship.Owner == game.Pid() {
			if ship.DockedStatus == hal.UNDOCKED {
				my_mutable_ships = append(my_mutable_ships, ship)
			} else {
				my_immutable_ships = append(my_immutable_ships, ship)
			}
		} else if ship.Owner == enemy_pid {
			enemy_ships = append(enemy_ships, ship)
		}
	}

	start_time := time.Now()

	evolver := NewEvolver(game, my_mutable_ships, my_immutable_ships, enemy_ships, 10)
	evolver.RunRushFight(15000, play_perfect)

	msg := pil.MSG_SECRET_SAUCE; if play_perfect { msg = pil.MSG_PERFECT_SAUCE }
	evolver.ExecuteGenome(msg)

	game.Log("Score: %v (i: %v, dvn: %v, cs: %v, t: %v)",
		evolver.genomes[0].score,
		evolver.iterations_required,
		evolver.genomes[0].score - evolver.null_score,
		evolver.cold_swaps,
		time.Now().Sub(start_time).Truncate(1 * time.Millisecond),
	)

	for _, ship := range game.MyShips() {
		if ship.DockedStatus != hal.UNDOCKED {
			game.Undock(ship)
		}
	}
}
