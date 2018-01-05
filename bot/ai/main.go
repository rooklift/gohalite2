package ai

import (
	"math/rand"
	"sort"

	gen "../genetic"
	hal "../core"
	pil "../pilot"
)

const (
	NOT_RUSHING = -1
	UNDECIDED = 0
	RUSHING = 1
)

// --------------------------------------------

type Config struct {
	Conservative			bool
	ForceRush				bool
	NoMsg					bool
	Imperfect				bool
	Profile					bool
	Timeseed				bool
}

type Overmind struct {
	Config					*Config
	Pilots					[]*pil.Pilot		// Stored in no particular order, sort at will
	Game					*hal.Game
	CowardFlag				bool
	RushChoice				int					// Affects ChooseTargets(), ResetPilots() and OptimisePilots()
	RushEnemyID				int
	NeverGA					bool
	FirstDockingTurn		int					// The turn we first thought about docking. -1 means never.
	FirstLaunchTurn			int					// The turn we first had a chance to undock. -1 means never.
}

func NewOvermind(game *hal.Game, config *Config) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.Config = config

	if game.InitialPlayers() == 2 {
		game.SetThreatRange(20)					// This value seems to be surprisingly fine-tuned
	} else {
		game.SetThreatRange(10)
	}

	ret.FindRushEnemy()

	if config.Conservative {
		ret.NeverGA = true
		ret.RushChoice = NOT_RUSHING
	} else if config.ForceRush {
		ret.RushChoice = RUSHING
	}

	ret.FirstDockingTurn = -1
	ret.FirstLaunchTurn = -1

	return ret
}

// --------------------------------------------

func (self *Overmind) Step() {

	if self.RushChoice == UNDECIDED {
		self.DecideRush()
	} else if self.RushChoice == RUSHING {
		self.MaybeEndRush()
	}

	if self.FirstLaunchTurn == -1 {
		for _, pilot := range self.Pilots {
			if pilot.DockedStatus == hal.DOCKED {
				self.FirstLaunchTurn = self.Game.Turn()
			}
		}
	}

	self.ResetPilots()

	if self.FirstLaunchTurn == self.Game.Turn() {			// We have a docked ship for the first time. Emergency undock?
		if self.LateRushDetector() {
			self.RushChoice = RUSHING
		}
	}

	self.SetCowardFlag()

	if self.Game.Turn() == 0 {
		if self.RushChoice == RUSHING {
			self.TurnZeroCluster()
			return
		} else {
			self.ChooseThreeDocks()
		}
	}

	if self.CowardFlag {
		self.CowardStep()
		return
	}

	if self.NeverGA == false && self.DetectRushFight() {

		if self.CanAvoidBad2v1() {
			self.AvoidBad2v1()
		} else {
			gen.FightRush(self.Game, self.RushEnemyID, self.Config.Imperfect)
			self.RushChoice = RUSHING
			return
		}
	}

	self.NormalStep()

	// Maybe just cancel our whole order...

	if self.FirstDockingTurn == self.Game.Turn() {
		self.MaybeDefend_4p_Rush()
	}

	self.DebugNavStack()
	self.DebugInhibition()
	self.DebugOrders()
}

func (self *Overmind) NormalStep() {

	self.ChooseTargets()
	self.OptimisePilots()
	self.SetInhibition()							// We might use target info for this in future, so put it here.
	self.ExecuteMoves()

	if self.RushChoice == RUSHING {
		self.UndockAll()
	}
}

// --------------------------------------------

func (self *Overmind) ResetPilots() {

	game := self.Game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		pilot := pil.NewPilot(sid, game)
		self.Pilots = append(self.Pilots, pilot)
	}

	// Clear various variables, including target in most cases. Also delete pilot if the ship is dead.

	for i := 0; i < len(self.Pilots); i++ {
		pilot := self.Pilots[i]
		alive := pilot.ResetAndUpdate(self.RushChoice != RUSHING)			// Reset if not rushing
		if alive == false {
			self.Pilots = append(self.Pilots[:i], self.Pilots[i+1:]...)
			i--
		}
	}
}

// --------------------------------------------

func (self *Overmind) ExecuteMoves() {

	raw_avoid_list := self.Game.AllImmobile()
	var avoid_list []hal.Entity

	for _, entity := range raw_avoid_list {
		switch entity.Type() {
		case hal.SHIP:
			if entity.(*hal.Ship).Doomed == false {
				avoid_list = append(avoid_list, entity)
			}
		default:
			avoid_list = append(avoid_list, entity)
		}
	}

	// Setup data structures...

	var mobile_pilots []*pil.Pilot
	var frozen_pilots []*pil.Pilot				// Note that this doesn't include docked / docking / undocking ships.

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.UNDOCKED && pilot.Doomed == false {
			mobile_pilots = append(mobile_pilots, pilot)
		}
	}

	// As a special case (relevant for 1v1 rushes) sort 3 ships by distance to centre...

	if len(mobile_pilots) <= 3 {

		centre_of_gravity := self.Game.AllShipsCentreOfGravity()

		sort.Slice(mobile_pilots, func(a, b int) bool {
			return mobile_pilots[a].Dist(centre_of_gravity) < mobile_pilots[b].Dist(centre_of_gravity)
		})
	}

	// Plan moves, add non-moving ships to the avoid list, then scrap other moves and plan them again...

	for _, pilot := range mobile_pilots {
		pilot.PlanChase(avoid_list)
	}

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		if pilot.HasStationaryPlan() {
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			avoid_list = append(avoid_list, pilot.Ship)
			i--
		}
	}

	for _, pilot := range mobile_pilots {
		pilot.PlanChase(avoid_list)
	}

	// Since our plans are based on the avoid_list, the only danger is 2 "mobile" ships colliding.
	// Note that it's possible that one of the colliding ships will not actually be moving.

	pil.ExecuteSafely(mobile_pilots)

	// Randomly give up for half the ships that still aren't moving, and
	// retry the pathfinding with the other half.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		if pilot.HasExecuted == false && rand.Intn(2) == 0 {
			pilot.PlanThrust(0, 0)
			pilot.Message = pil.MSG_ATC_DEACTIVATED
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			avoid_list = append(avoid_list, pilot.Ship)
			i--
		}
	}

	// Remake plans for our non-moving ships that we didn't freeze...

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			pilot.PlanChase(avoid_list)
		}
	}

	// And execute. Note that pilots that have already executed won't be affected...

	pil.ExecuteSafely(mobile_pilots)

	// Null thrust every "mobile" ship that didn't move. This causes target info to be put into
	// the replay via the Angle Message system.

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			if pilot.Plan != "" {
				pilot.PlanThrust(0, 0)
				pilot.Message = pil.MSG_ATC_RESTRICT
				pilot.ExecutePlan()
			}
		}
	}

	// Don't forget our non-mobile ships!

	for _, pilot := range frozen_pilots {

		pilot.ExecutePlan()

		if self.FirstDockingTurn == -1 {
			if hal.GetOrderType(pilot.Plan) == "d" {
				self.FirstDockingTurn = self.Game.Turn()
			}
		}
	}
}

// --------------------------------------------

func (self *Overmind) SetInhibition() {

	all_ships := self.Game.AllShips()

	for _, pilot := range self.Pilots {
		pilot.SetInhibition(all_ships)
	}
}

// --------------------------------------------

func (self *Overmind) DebugNavStack() {

	const (
		DEBUG_TURN = -1
		DEBUG_SHIP_ID = -1
	)

	if self.Game.Turn() == DEBUG_TURN {
		for _, pilot := range self.Pilots {
			if pilot.Id == DEBUG_SHIP_ID {
				pilot.LogNavStack()
				break
			}
		}
	}
}

func (self *Overmind) DebugInhibition() {

	const (
		DEBUG_TURN = -1
		DEBUG_SHIP_ID = -1
	)

	if self.Game.Turn() == DEBUG_TURN {
		for _, pilot := range self.Pilots {
			if pilot.Id == DEBUG_SHIP_ID {
				pilot.Log("Inhibition: %f; DangerShips: %d", pilot.Inhibition, pilot.DangerShips)
				break
			}
		}
	}
}

func (self *Overmind) DebugOrders() {

	const (
		DEBUG_TURN = 8
		DEBUG_SHIP_ID = 5
	)

	if self.Game.Turn() == DEBUG_TURN {
		for _, pilot := range self.Pilots {
			if pilot.Id == DEBUG_SHIP_ID {
				pilot.Log("Docked: %v. Order: %v", pilot.DockedStatus, self.Game.CurrentOrder(pilot.Ship))
			}
		}
	}
}

// --------------------------------------------

func (self *Overmind) WeAreBeing_4p_Rushed() bool {

	if self.Game.InitialPlayers() <= 2 {
		return false
	}

	relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)

	if len(relevant_enemies) == 0 {
		return false
	}

	// return true
	return false
}

func (self *Overmind) MaybeDefend_4p_Rush() {

	// Called at the time of our first dock.

	if self.WeAreBeing_4p_Rushed() {

		for _, pilot := range self.Pilots {
			pilot.ResetAndUpdate(true)
			pilot.Target = hal.Nothing			// Needed because ResetAndUpdate() won't clear PORT targets.
		}

		self.RushChoice = RUSHING

		self.NormalStep()
	}
}

func (self *Overmind) LateRushDetector() bool {

	// Called on the first turn when we can undock.

	relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)

	docked := 0

	for _, enemy := range relevant_enemies {
		if enemy.DockedStatus != hal.UNDOCKED {
			docked++
		}
	}

	if len(relevant_enemies) - docked > 1 {
		return true
	}

	return false
}

func (self *Overmind) UndockAll() {
	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.DOCKED {
			pilot.PlanUndock()
			pilot.ExecutePlan()
		}
	}
}
