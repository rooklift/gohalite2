package genetic

import (
	"fmt"
	"math"
	"sort"
	"strings"

	hal "../core"
)

type Sim struct {
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

			} else if event.what == PLANET_COLLISION {						// FIXME: if we use this for real sims, we need to do planet damage.

				event.ship_a.hp = 0

			}
		}

		// Apply weapon damage...

		for _, ship := range self.ships {
			if len(ship.actual_targets) > 0 {
				damage := hal.WEAPON_DAMAGE / len(ship.actual_targets)		// Right? A straight up integer truncation?
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

