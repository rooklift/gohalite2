package ai

import (
	// "math/rand"
	// "sort"

	// gen "../genetic"
	hal "../core"
	pil "../pilot"
)

const (
	NOT_RUSHING = -1
	UNDECIDED = 0
	RUSHING = 1
)

const (
	DEBUG_TURN = -1
	DEBUG_SHIP_ID = -1
)

// --------------------------------------------

type Config struct {
	Centre					bool
	Conservative			bool
	ForceRush				bool
	NoMsg					bool
	Imperfect				bool
	Profile					bool
	Split					bool
	Timeseed				bool

	TestGA					int
}

type Overmind struct {
	Config					*Config
	Pilots					[]*pil.Pilot		// Stored in no particular order, sort at will
	Game					*hal.Game
	CowardFlag				bool
	RushChoice				int					// Affects ChooseTargets(), ResetPilots() and OptimisePilots()
	RushEnemyID				int
	MyRushSide				hal.Edge			// Which side I am on when facing a rush (e.g. I might be LEFT side)
	NeverGA					bool
	FirstLaunchTurn			int					// The turn we first had a chance to undock. -1 means never.
	AvoidingBad2v1			bool				// AvoidBad2v1() has been called.

	RushEnemiesTouched		map[int]bool		// For deciding whether we can enter GA.
	EverDocked				bool				// Also allows us to enter the GA.
}

func NewOvermind(game *hal.Game, config *Config) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.Config = config

	if game.InitialPlayers() == 2 {
		game.SetThreatRange(20)					// This value seems to be surprisingly fine-tuned
	} else {
		//game.SetThreatRange(10)
		game.SetThreatRange(20)					// EXPERIMENT
	}

	ret.FindRushEnemy()

	if config.Conservative {
		ret.NeverGA = true
		ret.RushChoice = NOT_RUSHING
	} else if config.ForceRush {
		ret.RushChoice = RUSHING
	}

	ret.FirstLaunchTurn = -1
	ret.RushEnemiesTouched = make(map[int]bool)

	return ret
}

// --------------------------------------------

func (self *Overmind) Step() {

	if self.EverDocked == false {
		for _, ship := range self.Game.MyShips() {
			if ship.DockedStatus != hal.UNDOCKED {
				self.EverDocked = true
				break
			}
		}
	}

	if self.RushChoice == UNDECIDED {
		self.DecideRush()
		if self.RushChoice == RUSHING {
			self.ClearAllTargets()
		}
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

	if self.FirstLaunchTurn == self.Game.Turn() && self.AvoidingBad2v1 == false {	// We have a docked ship for the first time. Emergency undock?
		if self.Config.Conservative == false {
			self.Game.Log("Running late rush detector...")
			if self.LateRushDetector() {
				self.RushChoice = RUSHING
				self.ClearAllTargets()
			}
		}
	}

	self.SetCowardFlag()

	if self.Game.Turn() == 0 {
		if self.RushChoice != RUSHING {
			self.ChooseThreeDocks()
		} else {
			self.TurnZeroCluster()		// For tactical reasons - helps destroy single enemy ship sent to centre before it collides with us.
			return
		}
	}

	if self.CowardFlag {
		self.CowardStep()
		return
	}

	if self.NeverGA == false && self.RushChoice == RUSHING && self.DetectRushFight() {

		if self.CanAvoidBad2v1() {
			self.AvoidBad2v1()
		} else {
			self.EnterGeneticAlgorithm()
			return
		}
	}

	self.NormalStep()

	// gen.EvolveGlobal(self.Game)					// Will take our moves and improve them. Maybe.
	// FIXME: we will have assumed our out-of-range ships weren't going to move. Check for collisions
	// due to this false assumption...

	self.DebugNavStack()
	self.DebugInhibition()
	self.DebugOrders()
	self.DebugTargets()
}

func (self *Overmind) NormalStep() {

	self.ChooseTargets()
	self.OptimisePilots()
	self.SetInhibition()							// We might use target info for this in future, so put it here.
	// self.ExecuteMoves()
	self.ExecuteMoves2()		// EXPERIMENT / FIXME

	if self.RushChoice == RUSHING && self.AvoidingBad2v1 == false {
		self.UndockAll()
	}
}

// --------------------------------------------

func (self *Overmind) ResetPilots() {

	// Add new AIs for new ships...

	my_new_ships := self.Game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		pilot := pil.NewPilot(sid, self.Game)
		self.Pilots = append(self.Pilots, pilot)
	}

	// Clear various variables, including target in most cases. Also delete pilot if the ship is dead.

	for i := 0; i < len(self.Pilots); i++ {
		pilot := self.Pilots[i]
		alive := pilot.ResetAndUpdate()
		if alive == false {
			self.Pilots = append(self.Pilots[:i], self.Pilots[i+1:]...)
			i--
		}
	}
}

func (self *Overmind) ClearAllTargets() {
	for _, pilot := range self.Pilots {
		pilot.Locked = false
		pilot.ResetAndUpdate()
		pilot.Target = hal.Nothing				// Clears PORT targets
	}
}

// --------------------------------------------

/*

func (self *Overmind) ExecuteMoves() {

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Dist(self.Pilots[a].Target) < self.Pilots[b].Dist(self.Pilots[b].Target)
	})

	raw_avoid_list := self.Game.AllImmobile()
	var avoid_list []hal.Entity

	ignore_inhibition := (self.RushChoice == RUSHING)

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
	var frozen_pilots []*pil.Pilot				// Note that this doesn't include (already) docking / docked / undocking ships.

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
		pilot.PlanChase(avoid_list, ignore_inhibition)
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
		pilot.PlanChase(avoid_list, ignore_inhibition)
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
			pilot.PlanChase(avoid_list, ignore_inhibition)
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
				// pilot.Message = pil.MSG_ATC_RESTRICT
				pilot.ExecutePlan()
			}
		}
	}

	// Don't forget our non-mobile ships!

	for _, pilot := range frozen_pilots {
		pilot.ExecutePlan()
	}
}

*/

// --------------------------------------------

func (self *Overmind) SetInhibition() {

	all_ships := self.Game.AllShips()

	for _, pilot := range self.Pilots {
		pilot.SetInhibition(all_ships)
	}
}

// --------------------------------------------

func (self *Overmind) DebugNavStack() {
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
	if self.Game.Turn() == DEBUG_TURN {
		for _, pilot := range self.Pilots {
			if pilot.Id == DEBUG_SHIP_ID {
				pilot.Log("Docked: %v. Order: %v", pilot.DockedStatus, self.Game.CurrentOrder(pilot.Ship))
			}
		}
	}
}

func (self *Overmind) DebugTargets() {
	if self.Game.Turn() == DEBUG_TURN {
		for _, pilot := range self.Pilots {
			if pilot.Id == DEBUG_SHIP_ID {
				pilot.Log("Target: %v", pilot.Target)
			}
		}
	}
}

// --------------------------------------------

func (self *Overmind) LateRushDetector() bool {

	// Called on the first turn when we can undock.

	relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)
	my_centre_of_gravity := self.Game.MyShipsCentreOfGravity()

	dangerous := 0

	for _, enemy := range relevant_enemies {
		if enemy.DockedStatus == hal.UNDOCKED {
			if enemy.VagueDirection() == self.MyRushSide {
				if enemy.Dist(my_centre_of_gravity) < 90 {
					dangerous++
				}
			}
		}
	}

	if dangerous > 1 {
		self.Game.Log("LateRushDetector(): dangerous == %v", dangerous)
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
