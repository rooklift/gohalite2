package genetic

// The genetic algorithm code is a mess, having gained any number of kludges over the competition.
// Ideally I would rewrite it, but it works very well so I prefer to leave it alone.

import (
	"math/rand"
	"sort"
	"time"

	hal "../core"
	pil "../pilot"			// Just for message constants
)

const (
	CHAINS = 10
	PANIC_RANGE = 30		// How far the enemy can get before we worry
)

var thresholds = [CHAINS]float64{1.0, 0.999, 0.995, 0.99, 0.98, 0.96, 0.93, 0.9, 0.8, 0.7}

var chase_permutations_3v3 = [][]int{
	[]int{0,1,2},
	[]int{0,2,1},
	[]int{1,0,2},
	[]int{1,2,0},
	[]int{2,0,1},
	[]int{2,1,0},
}

var chase_permutations_3v2 = [][]int{
	[]int{0,0,1},
	[]int{0,1,0},
	[]int{1,0,0},
	[]int{1,1,0},
	[]int{1,0,1},
	[]int{0,1,1},
}

var chase_permutations_2v2 = [][]int{
	[]int{0,1},
	[]int{1,0},
}

// --------------------------------------------------------------------

type Gene struct {			// A gene is an instruction to a ship.
	speed		int
	angle		int
}

type Genome struct {
	genes		[]*Gene
	score		int
}

func (self *Genome) Copy() *Genome {
	ret := new(Genome)
	for _, gene := range self.genes {
		new_gene := new(Gene)
		*new_gene = *gene
		ret.genes = append(ret.genes, new_gene)
	}
	ret.score = self.score
	return ret
}

func (self *Genome) Init(size int, randomise bool) {
	self.genes = nil
	for i := 0; i < size; i++ {

		speed, angle := 0, 0;
		if randomise { speed, angle = rand.Intn(8), rand.Intn(360) }

		self.genes = append(self.genes, &Gene{
			speed: speed,
			angle: angle,
		})
	}
	self.score = -2147483647
}

func (self *Genome) Mutate(mutable_ship_ordinals []int) {		// [0,1,2] to be able to mutate every ship...

	i := rand.Intn(len(mutable_ship_ordinals))

	switch rand.Intn(3) {
	case 0:
		self.genes[mutable_ship_ordinals[i]].speed = rand.Intn(8)
	case 1:
		self.genes[mutable_ship_ordinals[i]].angle = rand.Intn(360)
	case 2:
		self.genes[mutable_ship_ordinals[i]].speed = rand.Intn(8)
		self.genes[mutable_ship_ordinals[i]].angle = rand.Intn(360)
	}
}

// --------------------------------------------------------------------

func EvolveGenome(game *hal.Game, iterations int, play_perfect bool, enemy_pid int) (*Genome, int, int) {

	// We need to take a genome's average score against a variety of scenarios, one of which should be no moves from enemy.
	// Perhaps another should be the enemy ships blinking out of existence, so we don't crash into planets.

	width, height := float64(game.Width()), float64(game.Height())

	initial_sim := SetupSim(game, enemy_pid)

	sim_without_enemies := initial_sim.Copy()
	for i := 0; i < len(sim_without_enemies.ships); i++ {
		if sim_without_enemies.ships[i].owner != game.Pid() {
			sim_without_enemies.ships = append(sim_without_enemies.ships[:i], sim_without_enemies.ships[i+1:]...)
			i--
		}
	}

	var genomes []*Genome

	for n := 0; n < CHAINS; n++ {
		g := new(Genome)
		if n == 0 {
			g.Init(len(game.MyShips()), false)		// Don't randomise the coldest genome; start with thrust 0 angle 0.
		} else {
			g.Init(len(game.MyShips()), true)
		}
		genomes = append(genomes, g)
	}

	var mutable_ship_ordinals []int
	for i, ship := range game.MyShips() {
		if ship.Owner == game.Pid() && ship.DockedStatus == hal.UNDOCKED {
			mutable_ship_ordinals = append(mutable_ship_ordinals, i)
		}
	}

	if len(mutable_ship_ordinals) == 0 {
		return genomes[0], -1, -1
	}

	best_score := -2147483647			// Solely used for
	iterations_required := 0			// reporting info.
	var null_score int

	for n := 0; n < iterations; n++ {

		// We run various chains of evolution with different "heats" (i.e. how willing we are to accept bad mutations)
		// in the "metropolis coupling" fashion.

		for c := 0; c < CHAINS; c++ {

			genome := genomes[c].Copy()

			if n > 0 {						// Don't mutate first iteration, so we get a true score for our initial genome. Makes the results stable.
				genome.Mutate(mutable_ship_ordinals)
			}

			genome.score = 0

			// We run some different scenarios of what the enemy will do.

			for scenario := 0; scenario < 3; scenario++ {

				var sim *Sim

				if scenario == 0 {						// Scenario 0 is the enemy ships not existing at at all (so we don't hit planets, etc)
					sim = sim_without_enemies.Copy()
				} else {
					sim = initial_sim.Copy()
				}

				var my_sim_ship_ptrs []*SimShip
				var enemy_sim_ship_ptrs []*SimShip

				for _, ship := range sim.ships {
					if ship.owner == game.Pid() {
						my_sim_ship_ptrs = append(my_sim_ship_ptrs, ship)
					} else {
						enemy_sim_ship_ptrs = append(enemy_sim_ship_ptrs, ship)
					}
				}

				previous_sid := -1

				for i := 0; i < len(my_sim_ship_ptrs); i++ {

					if my_sim_ship_ptrs[i].dockedstatus == hal.UNDOCKED {

						speed := genome.genes[i].speed
						angle := genome.genes[i].angle
						vel_x, vel_y := hal.Projection(0, 0, float64(speed), angle)
						my_sim_ship_ptrs[i].vel_x = vel_x
						my_sim_ship_ptrs[i].vel_y = vel_y

					} else {

						my_sim_ship_ptrs[i].vel_x = 0
						my_sim_ship_ptrs[i].vel_y = 0
					}

					if my_sim_ship_ptrs[i].id <= previous_sid {
						panic("EvolveGenome(): my_sim_ship_ptrs not in order")
					}
					previous_sid = my_sim_ship_ptrs[i].id
				}

				for i := 0; i < len(enemy_sim_ship_ptrs); i++ {

					if enemy_sim_ship_ptrs[i].dockedstatus != hal.UNDOCKED {

						enemy_sim_ship_ptrs[i].vel_x = 0
						enemy_sim_ship_ptrs[i].vel_y = 0

					} else {

						switch scenario {

						case 0:
							// Scenario 0 is the enemy ships not existing at at all (so we don't hit planets, etc)

						case 1:
							real_ship, _ := game.GetShip(enemy_sim_ship_ptrs[i].id)
							enemy_sim_ship_ptrs[i].vel_x = real_ship.Dx
							enemy_sim_ship_ptrs[i].vel_y = real_ship.Dy

						case 2:
							enemy_sim_ship_ptrs[i].vel_x = 0
							enemy_sim_ship_ptrs[i].vel_y = 0

						}
					}
				}

				sim.Step()

				// SCORING -----------------------------------------------------------------------------------------------------------

				// Damage...

				for _, ship := range enemy_sim_ship_ptrs {
					if ship.hp > 0 {
						genome.score -= ship.hp * 100
					}
				}

				for _, ship := range my_sim_ship_ptrs {
					if ship.hp > 0 {
						genome.score += ship.hp * 100
					}
				}

				// Other scores only need to be run in one scenario to work...

				if scenario == 2 {

					// A good scenario to run our stupidity checks in. In particular, enemy docked ships exist
					// in this scenario and our ships are flagged as stupid if they have collided with them.

					for _, ship := range my_sim_ship_ptrs {
						if ship.stupid_death || ship.x <= 0 || ship.x >= width || ship.y <= 0 || ship.y >= height {
							genome.score -= 9999999
						}
					}
				}

				if scenario == 0 {

					// A good scenario to run every other check in.
					// Note that enemy_sim_ship_ptrs is empty here, so use real ships...

					real_enemy_ships := game.ShipsOwnedBy(enemy_pid)

					// EDGES OF SPACE / PLANET AVOIDANCE -----------------------------------------------------------------------------

					for _, ship := range my_sim_ship_ptrs {

						// Modest penalty for getting near edge of space...

						horiz_clearance := hal.MinFloat(ship.x, width - ship.x)
						vert_clearance := hal.MinFloat(ship.y, height - ship.y)

						if horiz_clearance < 20 {
							genome.score -= int(1000.0 - horiz_clearance)
						}
						if vert_clearance < 20 {
							genome.score -= int(1000.0 - vert_clearance)
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

					if len(my_sim_ship_ptrs) == 3 && len(real_enemy_ships) >= 2 {

						// Handles both 3v2 and 3v3.

						dist0 := 999999.9
						dist1 := 999999.9
						dist2 := 999999.9

						permutations := chase_permutations_3v2;
						if len(real_enemy_ships) > 2 {
							permutations = chase_permutations_3v3;
						}

						for _, perm := range permutations {

							this_dist0 := my_sim_ship_ptrs[0].Dist(real_enemy_ships[perm[0]])
							this_dist1 := my_sim_ship_ptrs[1].Dist(real_enemy_ships[perm[1]])
							this_dist2 := my_sim_ship_ptrs[2].Dist(real_enemy_ships[perm[2]])

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

					} else if len(my_sim_ship_ptrs) == 2 && len(real_enemy_ships) == 2 {

						dist0 := 999999.9
						dist1 := 999999.9

						for _, perm := range chase_permutations_2v2 {

							this_dist0 := my_sim_ship_ptrs[0].Dist(real_enemy_ships[perm[0]])
							this_dist1 := my_sim_ship_ptrs[1].Dist(real_enemy_ships[perm[1]])

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

							for _, ship := range my_sim_ship_ptrs {
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

						for _, ship := range my_sim_ship_ptrs {

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

					for _, ship := range my_sim_ship_ptrs {

						if play_perfect {

							// In "perfect" mode we give huge bonuses to moves that can only ever be hit by 1 enemy;
							// which means being < 13 away from the *starting* location of 1 enemy.

							var thirteens	[]int										// IDs of ships that might be able to hit us.
							var twelves		[]int										// As above, but with some tolerance.
							var eights		[]int										// IDs of ships that might be able to ram us.

							for _, enemy_ship := range real_enemy_ships {				// Must use real_enemy_ships, since enemy_sim_ship_ptrs is []

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
				null_score = genome.score		// Report the score of not moving. Relies on no mutation in n0 and non-randomised c0.
			}

			if float64(genome.score) > float64(genomes[c].score) * thresholds[c] {
				genomes[c] = genome
			}
		}

		sort.SliceStable(genomes, func(a, b int) bool {
			return genomes[a].score > genomes[b].score		// Note the reversed sort, high scores come first.
		})

		if genomes[0].score > best_score {
			best_score = genomes[0].score					// This is for
			iterations_required = n							// info only.
		}

		if time.Now().Sub(game.ParseTime()) > 1500 * time.Millisecond {
			game.Log("Emergency timeout in EvolveGenome() after %d iterations.", n)
			return genomes[0], iterations_required, null_score
		}
	}

	return genomes[0], iterations_required, null_score
}

func FightRush(game *hal.Game, enemy_pid int, play_perfect bool) {

	game.LogOnce("Entering genetic algorithm!")

	genome, iterations_required, null_score := EvolveGenome(game, 15000, play_perfect, enemy_pid)

	var order_elements []int

	msg := pil.MSG_SECRET_SAUCE; if play_perfect { msg = pil.MSG_PERFECT_SAUCE }

	for i, ship := range game.MyShips() {									// Guaranteed sorted by ID
		if ship.DockedStatus == hal.UNDOCKED {
			game.Thrust(ship, genome.genes[i].speed, genome.genes[i].angle)
			game.SetMessage(ship, msg)
			order_elements = append(order_elements, genome.genes[i].speed, genome.genes[i].angle)
		} else {
			game.Undock(ship)
			order_elements = append(order_elements, -1, -1)
		}
	}

	game.Log("Score: %v (iter %v, dvn: %v). Orders: %v", genome.score, iterations_required, genome.score - null_score, order_elements)
	// dvn is difference versus null
}
