package genetic

import (
	"sort"

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
						my_mutable_ship_map[ship.Id] = ship
						relevant_enemy_map[enemy.Id] = enemy
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

	evolver.RunGlobalFight()
	evolver.ExecuteGenome(pil.MSG_GLOBAL_SAUCE)

	game.Log("EvolveGlobal() lens: %v, %v, %v", len(my_mutable_ships), len(my_immutable_ships), len(relevant_enemy_ships))
}

func (self *Evolver) RunGlobalFight() {
	return
}
