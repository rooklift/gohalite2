package genetic

import (
	"sort"

	hal "../core"
)

func EvolveGlobal(game *hal.Game) {

	my_mutable_ship_map := make(map[int]*hal.Ship)
	my_immutable_ship_map := make(map[int]*hal.Ship)
	relevant_enemy_map := make(map[int]*hal.Ship)

	my_ships := game.MyShips()
	enemy_ships := game.EnemyShips()

	for _, ship := range my_ships {
		if ship.DockedStatus == hal.UNDOCKED {
			for _, enemy := range enemy_ships {
				if ship.Dist(enemy) < 20 {
					my_mutable_ship_map[ship.Id] = ship
					relevant_enemy_map[enemy.Id] = enemy
				}
			}
		}
	}

	if len(my_mutable_ship_map) == 0 {
		return
	}

	for _, ship := range my_ships {
		if my_mutable_ship_map[ship.Id] == nil {
			for _, other := range my_mutable_ship_map {
				if ship.Dist(other) < 8 {
					my_immutable_ship_map[ship.Id] = ship
				}
			}
		}
	}

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

	// We keep our own mutable ships sorted by ID, I forget if this is really needed.
	// May help with determinism.

	sort.Slice(my_mutable_ships, func(a, b int) bool {
		return my_mutable_ships[a].Id < my_mutable_ships[b].Id
	})

	/* evolver := */ NewEvolver(game, my_mutable_ships, my_immutable_ships, relevant_enemy_ships, 1)
}
