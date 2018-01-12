package genetic

import (
	"math/rand"

	hal "../core"
)

var thresholds = [10]float64{1.0, 0.999, 0.995, 0.99, 0.98, 0.96, 0.93, 0.9, 0.8, 0.7}

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

// --------------------------------------------------------------------

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

func (self *Genome) Mutate() {

	if len(self.genes) == 0 {
		return
	}

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

type Evolver struct {

	// Note that we keep our sim's ships in order: mutable friendly, immutable friendly, enemy.
	// The sim itself doesn't know or care, but we do.

	game					*hal.Game
	genomes					[]*Genome
	genome_length			int
	sim						*Sim
	sim_without_enemies		*Sim
	first_enemy_index		int			// Doesn't mean we have enemies. Equal to number of friendlies (mutable or not) in the sim.

	iterations_required		int
	null_score				int
	cold_swaps				int

}

func NewEvolver(game *hal.Game, my_mutable_ships, my_immutable_ships, enemy_ships []*hal.Ship, mc_chains int) *Evolver {

	ret := new(Evolver)

	ret.game = game

	for n := 0; n < mc_chains; n++ {
		ret.genomes = append(ret.genomes, new(Genome))
		if n == 0 {
			ret.genomes[n].Init(len(my_mutable_ships), false)
		} else {
			ret.genomes[n].Init(len(my_mutable_ships), true)
		}
	}

	ret.genome_length = len(my_mutable_ships)

	// We ensure our mutable ships are at the start of the baseSim's ships slice...

	var relevant_ships []*hal.Ship
	relevant_ships = append(relevant_ships, my_mutable_ships...)
	relevant_ships = append(relevant_ships, my_immutable_ships...)
	ret.first_enemy_index = len(relevant_ships)
	relevant_ships = append(relevant_ships, enemy_ships...)

	ret.sim = SetupSim(game, relevant_ships)

	ret.sim_without_enemies = ret.sim.Copy()
	for i := 0; i < len(ret.sim_without_enemies.ships); i++ {
		if ret.sim_without_enemies.ships[i].owner != game.Pid() {
			ret.sim_without_enemies.ships = append(ret.sim_without_enemies.ships[:i], ret.sim_without_enemies.ships[i+1:]...)
			i--
		}
	}

	return ret
}

func (self *Evolver) ExecuteGenome(msg int) {

	for i, gene := range self.genomes[0].genes {

		real_ship := self.sim.ships[i].real_ship			// Relying on our mutable ships being stored first.

		if real_ship.DockedStatus == hal.UNDOCKED {
			self.game.Thrust(real_ship, gene.speed, gene.angle)
			self.game.SetMessage(real_ship, msg)
		}
	}
}
