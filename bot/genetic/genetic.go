package genetic

import (
	"math/rand"
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
