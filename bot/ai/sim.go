package ai

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	hal "../gohalite2"
)

type Sim struct {					// Using pointers, unlike in most of the code
	planets			[]*SimPlanet
	ships			[]*SimShip
}

func (self *Sim) String() string {
	var s []string
	for _, ship := range self.ships {
		s = append(s, fmt.Sprintf("%d", ship.hp))
	}
	return fmt.Sprintf("len %d... ", len(self.ships)) + strings.Join(s, "/")
}

func (self *Sim) Copy() *Sim {
	ret := new(Sim)
	for _, planet := range self.planets {
		new_planet := new(SimPlanet)
		*new_planet = *planet
		ret.planets = append(ret.planets, new_planet)
	}
	for _, ship := range self.ships {
		new_ship := new(SimShip)
		*new_ship = *ship
		new_ship.actual_targets = nil
		ret.ships = append(ret.ships, new_ship)
	}
	return ret
}

type SimEntity struct {
	x				float64
	y				float64
	radius			float64
	vel_x			float64
	vel_y			float64
}

type SimPlanet struct {
	SimEntity
}

type SimShip struct {
	SimEntity
	ship_state		ShipState
	weapon_state	WeaponState
	actual_targets	[]*SimShip			// Who we actually, really, definitely shoot at.
	dockedstatus	hal.DockedStatus
	owner			int
	hp				int
	id				int
}

func (self *SimShip) Dist(e hal.Entity) float64 {
	dx := self.x - e.GetX()
	dy := self.y - e.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

type PossibleEvent struct {
	ship_a			*SimShip
	ship_b			*SimShip
	planet			*SimPlanet
	t				float64
	what			EventType
}

type ShipState int; const (
	ALIVE ShipState = iota
	DEAD
)

type WeaponState int; const (
	READY WeaponState = iota
	FIRING
	SPENT
)

type EventType int; const (
	ATTACK EventType = iota
	SHIP_COLLISION
	PLANET_COLLISION
)

func (self *Sim) Step() {

	var possible_events []*PossibleEvent

	for _, ship := range self.ships {
		ship.actual_targets = nil
	}

	// Possible attacks...

	for i, ship_a := range self.ships {
		if ship_a.hp > 0 {
			for _, ship_b := range self.ships[i+1:] {
				if ship_b.hp > 0 {
					if ship_a.owner != ship_b.owner {
						t, ok := CollisionTime(6.0, &ship_a.SimEntity, &ship_b.SimEntity)		// Should be 6.0 since bugfixes around 6 Nov 2017.
						if ok && t >= 0 && t <= 1 {
							possible_events = append(possible_events, &PossibleEvent{ship_a, ship_b, nil, t, ATTACK})
						}
					}
				}
			}
		}
	}

	// Possible ship collisions...

	for i, ship_a := range self.ships {
		if ship_a.hp > 0 {
			for _, ship_b := range self.ships[i+1:] {
				if ship_b.hp > 0 {
					t, ok := CollisionTime(1.0, &ship_a.SimEntity, &ship_b.SimEntity)
					if ok && t >= 0 && t <= 1 {
						possible_events = append(possible_events, &PossibleEvent{ship_a, ship_b, nil, t, SHIP_COLLISION})
					}
				}
			}
		}
	}

	// Possible ship-planet collisions...

	for _, planet := range self.planets {
		for _, ship := range self.ships {
			t, ok := CollisionTime(planet.radius + 0.5, &ship.SimEntity, &planet.SimEntity)
			if ok && t >= 0 && t <= 1 {
				possible_events = append(possible_events, &PossibleEvent{ship, nil, planet, t, PLANET_COLLISION})
			}
		}
	}

	// We want possible events sorted into groups of simultaneous events.

	sort.Slice(possible_events, func(a, b int) bool {
		return possible_events[a].t < possible_events[b].t
	})

	var grouped_events [][]*PossibleEvent

	current_t := -1.0

	for _, event := range possible_events {

		t := event.t
		if t > current_t {
			current_t = t
			grouped_events = append(grouped_events, nil)
		}

		grouped_events[len(grouped_events) - 1] = append(grouped_events[len(grouped_events) - 1], event)
	}

	// Now update ships...

	for _, grouping := range grouped_events {

		// Go through each event and update stuff if the conditions are right...
		// Note that having 0 HP is not sufficient to be irrelevant, because we may
		// have died this very moment. Instead, check ship_state == DEAD.

		for _, event := range grouping {

			if event.what == SHIP_COLLISION {

				ship_a, ship_b := event.ship_a, event.ship_b

				if ship_a.ship_state == DEAD || ship_b.ship_state == DEAD {
					continue
				}

				ship_a.hp = 0
				ship_b.hp = 0

			} else if event.what == ATTACK {

				ship_a, ship_b := event.ship_a, event.ship_b

				if ship_a.ship_state == DEAD || ship_b.ship_state == DEAD {
					continue
				}

				if ship_a.weapon_state != SPENT && ship_a.dockedstatus == hal.UNDOCKED {
					ship_a.weapon_state = FIRING
					ship_a.actual_targets = append(ship_a.actual_targets, ship_b)
				}

				if ship_b.weapon_state != SPENT && ship_b.dockedstatus == hal.UNDOCKED {
					ship_b.weapon_state = FIRING
					ship_b.actual_targets = append(ship_b.actual_targets, ship_a)
				}

			} else if event.what == PLANET_COLLISION {				// FIXME: if we use this for real sims, we need to do planet damage.

				event.ship_a.hp = 0

			}
		}

		// Apply weapon damage...

		for _, ship := range self.ships {
			if len(ship.actual_targets) > 0 {
				damage := 64 / len(ship.actual_targets)				// Right? A straight up integer truncation?
				for _, target := range ship.actual_targets {
					target.hp -= damage
				}
			}
		}

		// Update transitional states to final state...

		for _, ship := range self.ships {

			if ship.weapon_state == FIRING {
				ship.weapon_state = SPENT
			}

			if ship.hp <= 0 {
				ship.ship_state = DEAD
				ship.hp = 0
			}

			ship.actual_targets = nil
		}
	}

	for _, ship := range self.ships {
		ship.x += ship.vel_x
		ship.y += ship.vel_y
	}
}

func SetupSim(game *hal.Game) *Sim {

	sim := new(Sim)

	for _, planet := range game.AllPlanets() {

		for _, ship := range game.AllShips() {

			if planet.Dist(ship) < planet.Radius + 8 {		// Only include relevant planets

				sim.planets = append(sim.planets, &SimPlanet{
					SimEntity{
						x: planet.X,
						y: planet.Y,
						radius: planet.Radius,
					},
				})

				break
			}
		}
	}

	for _, ship := range game.AllShips() {					// Guaranteed sorted by ship ID
		sim.ships = append(sim.ships, &SimShip{
			SimEntity: SimEntity{
				x: ship.X,
				y: ship.Y,
				radius: hal.SHIP_RADIUS,
			},
			ship_state: ALIVE,
			weapon_state: READY,
			dockedstatus: ship.DockedStatus,
			hp: ship.HP,
			owner: ship.Owner,
			id: ship.Id,
		})
	}

	return sim
}

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

	game.Log("Rush Evo! Scores: %v", all_scores)
	game.Log("           Steps: %v", all_steps)

	for i, ship := range game.MyShips() {									// Guaranteed sorted by ID
		game.ThrustWithMessage(ship, genome.genes[i].speed, genome.genes[i].angle, int(MSG_SECRET_SAUCE))
	}
}
