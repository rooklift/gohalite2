package genetic

import (
	"math/rand"
	"sort"
	"time"

	hal "../core"
	pil "../pilot"			// Just for message constants
)

const CHAINS = 10
var thresholds = [CHAINS]float64{1.0, 0.999, 0.995, 0.99, 0.98, 0.96, 0.93, 0.9, 0.8, 0.7}

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

func (self *Genome) Init(size int) {
	self.genes = nil
	for i := 0; i < size; i++ {
		self.genes = append(self.genes, &Gene{
			speed: rand.Intn(8),
			angle: rand.Intn(360),
		})
	}
	self.score = -2147483647
}

func (self *Genome) Mutate() {
	i := rand.Intn(len(self.genes))
	switch rand.Intn(3) {
	case 0:
		self.genes[i].speed = rand.Intn(8)
	case 1:
		self.genes[i].angle = rand.Intn(360)
	case 2:
		self.genes[i].speed = rand.Intn(8)
		self.genes[i].angle = rand.Intn(360)
	}
}

// --------------------------------------------------------------------

func EvolveGenome(game *hal.Game, iterations int, play_perfect bool) (*Genome, int) {

	// We need to take a genome's average score against a variety of scenarios, one of which should be no moves from enemy.
	// Perhaps another should be the enemy ships blinking out of existence, so we don't crash into planets.

	width, height := float64(game.Width()), float64(game.Height())

	initial_sim := SetupSim(game)

	sim_without_enemies := initial_sim.Copy()
	for i := 0; i < len(sim_without_enemies.ships); i++ {
		if sim_without_enemies.ships[i].owner != game.Pid() {
			sim_without_enemies.ships = append(sim_without_enemies.ships[:i], sim_without_enemies.ships[i+1:]...)
			i--
		}
	}

	centre_of_gravity := game.AllShipsCentreOfGravity()

	var genomes []*Genome

	for n := 0; n < CHAINS; n++ {
		g := new(Genome)
		g.Init(len(game.MyShips()))
		genomes = append(genomes, g)
	}

	best_score := -2147483647			// Solely used for
	iterations_required := 0			// reporting info.

	for n := 0; n < iterations; n++ {

		// We run various chains of evolution with different "heats" (i.e. how willing we are to accept bad mutations)
		// in the "metropolis coupling" fashion.

		for c := 0; c < CHAINS; c++ {

			genome := genomes[c].Copy()
			genome.Mutate()

			genome.score = 0

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

				for i := 0; i < len(my_sim_ship_ptrs); i++ {
					speed := genome.genes[i].speed
					angle := genome.genes[i].angle
					vel_x, vel_y := hal.Projection(0, 0, float64(speed), angle)
					my_sim_ship_ptrs[i].vel_x = vel_x
					my_sim_ship_ptrs[i].vel_y = vel_y
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
							last_move := game.LastTurnMoveById(enemy_sim_ship_ptrs[i].id)
							enemy_sim_ship_ptrs[i].vel_x = last_move.Dx
							enemy_sim_ship_ptrs[i].vel_y = last_move.Dy

						case 2:
							enemy_sim_ship_ptrs[i].vel_x = 0
							enemy_sim_ship_ptrs[i].vel_y = 0

						}
					}
				}

				sim.Step()

				var good_thirteens = make(map[int]int)

				for _, ship := range my_sim_ship_ptrs {

					if ship.hp > 0 {
						genome.score += ship.hp * 100
					}

					genome.score -= int(ship.Dist(centre_of_gravity) * 2)

					if ship.x <= 0 || ship.x >= width || ship.y <= 0 || ship.y >= height {
						genome.score -= 200000
					}

					// -------------------------------

					// In "perfect" mode we give huge bonuses to moves that can only ever be hit by 1 enemy;
					// which means being < 13 away from the *starting* location of 1 enemy.

					if play_perfect && scenario == 0 {

						var thirteens []int									// IDs of ships that might be able to hit us.

						for _, enemy_ship := range game.EnemyShips() {		// i.e. using their actual game position without simulation.
							if ship.Dist(enemy_ship) < 13 {
								thirteens = append(thirteens, enemy_ship.Id)
							}
						}

						if len(thirteens) == 1 && ship.hp > 0 {
							genome.score += 100000
							enemy_ship_id := thirteens[0]
							good_thirteens[enemy_ship_id] += 1
						}

						if len(thirteens) > 1 {								// Don't check for ship being alive, so that we don't kill self to avoid this.
							genome.score -= 100000
						}
					}
				}

				// Modest bonus for coordinated thirteens (should be enough)

				for _, hits := range good_thirteens {
					genome.score += (hits - 1) * 5000
				}

				for _, ship := range enemy_sim_ship_ptrs {
					if ship.hp > 0 {
						genome.score -= ship.hp * 100
					}
				}
			}

			if float64(genome.score) > float64(genomes[c].score) * thresholds[c] {
				genomes[c] = genome
			}
		}

		sort.Slice(genomes, func(a, b int) bool {
			return genomes[a].score > genomes[b].score		// Note the reversed sort, high scores come first.
		})

		if genomes[0].score > best_score {
			best_score = genomes[0].score					// This is for
			iterations_required = n							// info only.
		}

		if time.Now().Sub(game.ParseTime()) > 1500 * time.Millisecond {
			game.Log("Emergency timeout in EvolveGenome() after %d iterations.", n)
			return genomes[0], iterations_required
		}
	}

	return genomes[0], iterations_required
}

func FightRush(game *hal.Game) {

	game.LogOnce("Entering dangerous rush situation!")

	play_perfect := true

	if len(game.MyShips()) == 1 || len(game.MyShips()) < len(game.EnemyShips()) {
		play_perfect = false
	}
	for _, ship := range game.EnemyShips() {
		if ship.DockedStatus != hal.UNDOCKED {
			play_perfect = false
		}
	}

	genome, iterations_required := EvolveGenome(game, 15000, play_perfect)

	var order_elements []int

	for i, ship := range game.MyShips() {									// Guaranteed sorted by ID
		game.Thrust(ship, genome.genes[i].speed, genome.genes[i].angle)
		game.SetMessage(ship, pil.MSG_SECRET_SAUCE)
		order_elements = append(order_elements, genome.genes[i].speed, genome.genes[i].angle)
	}

	game.Log("Rush Evo! Score: %v (iter %5v). Orders: %v", genome.score, iterations_required, order_elements)
}
