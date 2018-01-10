package genetic

import (
	"sort"

	hal "../core"
)

func EvolveGlobal(game *hal.Game) {

	my_mutable_ship_map := make(map[int]*hal.Ship)
	relevant_enemy_map := make(map[int]*hal.Ship)
	my_immutable_ship_map := make(map[int]*hal.Ship)

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

	for _, ship := range my_ships {
		if my_mutable_ship_map[ship.Id] == nil {
			for _, other := range my_mutable_ship_map {
				if ship.Dist(other) < 8 {
					my_immutable_ship_map[ship.Id] = ship
				}
			}
		}
	}

	var relevant_ships []*hal.Ship

	for _, ship := range my_mutable_ship_map {
		relevant_ships = append(relevant_ships, ship)
	}

	for _, ship := range my_immutable_ship_map {
		relevant_ships = append(relevant_ships, ship)
	}

	// We keep our own ships sorted by ID so that we can cleanly refer to them by ordinal position...

	sort.Slice(relevant_ships, func(a, b int) bool {
		return relevant_ships[a].Id < relevant_ships[b].Id
	})

	var mutable_ship_ordinals []int

	for i, ship := range relevant_ships {
		if my_mutable_ship_map[ship.Id] != nil {
			mutable_ship_ordinals = append(mutable_ship_ordinals, i)
		}
	}

	// Enemies can go in whatever order...

	for _, e := range relevant_enemy_map {
		relevant_ships = append(relevant_ships, e)
	}
}
