My bot in Go for the [Halite 2](https://halite.io/) (2017) programming contest.

# Initial Stateful Algorithm (before v45)

The algorithm I used until v45 was fairly simple in principle...

* Send new ships to the nearest target, which is either an enemy ship, or a planet that *needs help*, defined as:
  - Not fully docked by us, or:
  - Under threat of attack.
* When at a planet:
  - Attack the nearest ship that needs attacking, if there are any.
  - Otherwise, dock if possible.
  - Otherwise, choose a new destination as above.

# Stateless Algorithm (after v45)

I scrapped the above algorithm in version 45, becoming stateless. Now, each turn the bot generates a list of "problems" (e.g. planets to be attacked) and assigns nearby ships to those problems, with no state saved between turns.

* Generate all problems; each problem has a number of ships required (its "need").
* Iterate through the ships; go to the nearest problem that still needs help; reduce that need by 1.
* Make some tactical choices; e.g. if the problem is a planet, we may actually target an enemy ship.

# Collision Avoidance

Collision avoidance is fairly straightforward. Each ship starts off with its actual move set to null (stationary) but chooses a preferred move (e.g. thrust 7, 180) that it wants to make, if it can.

* Set each ship's *actual* move to null.
* Create a list of all entities that can't move (planets, docked ships).
* Avoiding those stationary objects, choose each ship's *preferred* move.
* Iterate through the ships:
  - Pretend that the ship will get its preferred move.
  - Check for collisions against other ships' *actual* moves.
  - If no collisions, set this ship's actual move to be its preferred move.
  - Repeat this whole loop several times.
* After a number of loops, if a ship still isn't moving, try reducing its speed.

# 1v1 Genetic Algorithm + Metropolis Coupling

In 2-player games it's sometimes sensible to rush the enemy, ignoring planets and going straight at 'em. In this case, when the ships are close to the enemy, the bot uses a genetic algorithm to find which moves are best, i.e. we generate a random "genome" (list of moves) and then do the following:

* Mutate the genome randomly.
* Simulate the results.
* If the new genome is good, keep it, otherwise discard.
* Repeat.

The problem is, the simulation needs to know what the opponent will do. I currently do very crude guessing. A more advanced technique would be to evolve the *opponent's* moves as above, and only then evolve our own moves to counteract them.

Another problem is local optima. To avoid these, I run multiple chains of evolution at once, with different "heats". Hot chains are allowed to accept bad mutations (the hotter the chain, the looser its standards are). Between iterations, the chains are sorted so that the colder chains have the better genomes. In this way, the cold chains can be pulled out of local optima. I believe this whole process is called "Metropolis Coupling".

Honestly this whole thing is over-engineered and under-performing.

# Global Strategy - Conceptual Breakthroughs

Some key conceptual breakthroughs that seemed to improve the bot were:

* Don't crash our ships into each other. This is important.

* Never dock when there are enemy ships around. This is obvious in hindsight. Docking is basically suicide. Instead, fight whatever enemy ships are in the area.

* It's OK to send ships to target interplanetary enemy ships. Before v23, I only sent ships to planets, then dealt with whatever enemies I found there. But directly targetting interplanetary ships increased mu by about 2 or 3. Such ships are up to no good and need destroying. More than that, you can get some nice clustering behaviour, with your ships coming together as they kill 1 enemy; then they will help each other as they go on to the next target.

* However... chasing ships is quite exploitable. An enemy can [send a single ship](https://halite.io/play/?game_id=2424227&replay_class=1&replay_name=replay-20171108-160208%2B0000--3470758710-312-208-1510156921) to distract many of your ships. Therefore, I started only allowing 1 ship to chase any particular interstellar enemy.

* When your ship is in range of an enemy at the start of a turn, it's guaranteed to attack that target (plus anything else in range at turn start). Instead of sitting there motionless, it's good to back away, making it harder for incoming enemies to attack you.

* In 4 player games, when you are seriously weak, it's best to flee and try and survive to take advantage of the (dubious) tiebreaker rules.

* Sometimes swapping 2 ships' targets reduces the overall distance they have to travel. So do this.

* Attacks right at the start of the turn are very predictable (only unexpected docking commands can mess this up). One can thus determine which ships will "certainly" die, and pretend they're not there. Using this information wisely is the hard part. At the very least, one can use it for navigation; i.e. skipping unneeded collision avoidance. One might also use it for strategic decisions, but this is harder.

* One should avoid unwise fights. Starting at v62 (but with a big number fix at v64), I use sum-of-distances-squared to decide whether each ship is "inhibited" or not; i.e. whether it has more enemies than friends nearby. If so, it flees.
