package genetic

import (
	"math/rand"
	"time"

	hal "../core"
	pil "../pilot"			// Just for message constants
)

type Gene struct {			// A gene is an instruction to a ship.
	speed		int
	angle		int
}

type Genome struct {
	genes		[]*Gene
}

func (self *Genome) Copy() *Genome {
	ret := new(Genome)
	for _, gene := range self.genes {
		new_gene := new(Gene)
		*new_gene = *gene
		ret.genes = append(ret.genes, new_gene)
	}
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

func EvolveGenome(game *hal.Game, iterations int) (*Genome, int, int) {

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

	best_genome := new(Genome)
	best_genome.Init(len(game.MyShips()))

	best_score := -999999

	steps := 0		// Counted for info - how many successful mutations we make.

	for n := 0; n < iterations; n++ {

		genome := best_genome.Copy()
		genome.Mutate()

		score := 0

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

			for _, ship := range my_sim_ship_ptrs {

				if ship.hp > 0 {
					score += ship.hp * 100
				}

				score -= int(ship.Dist(centre_of_gravity))

				if ship.x <= 0 || ship.x >= width || ship.y <= 0 || ship.y >= height {
					score -= 100000
				}
			}

			for _, ship := range enemy_sim_ship_ptrs {
				if ship.hp > 0 {
					score -= ship.hp * 100
				}
			}
		}

		if score > best_score {
			steps++
			best_score = score
			best_genome = genome.Copy()
		}
	}

	return best_genome, best_score, steps
}

func FightRush(game *hal.Game) {

	game.LogOnce("Entering dangerous rush situation!")

	var genome *Genome
	var best_score int

	var all_scores []int
	var all_steps []int			// Count of mutations used to produce each genome (for info only)

	for n := 0; n < 20; n++ {

		new_genome, score, steps := EvolveGenome(game, 2500)

		if score > best_score || genome == nil {
			genome = new_genome
			best_score = score
		}

		all_scores = append(all_scores, score)
		all_steps = append(all_steps, steps)

		// Hopefully the following is adequate to prevent timeouts...

		if time.Now().Sub(game.ParseTime()) > 1500 * time.Millisecond {
			game.Log("Emergency timeout in FightRush() after %d genomes.", n)
			break
		}
	}

	var order_elements []int

	for i, ship := range game.MyShips() {									// Guaranteed sorted by ID
		game.Thrust(ship, genome.genes[i].speed, genome.genes[i].angle)
		game.SetMessage(ship, pil.MSG_SECRET_SAUCE)
		order_elements = append(order_elements, genome.genes[i].speed, genome.genes[i].angle)
	}

	game.Log("Rush Evo! Scores: %v", all_scores)
	game.Log("           Steps: %v", all_steps)
	game.Log("          Orders: %v", order_elements)
}
