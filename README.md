My bot in Go for the [Halite 2](https://halite.io/) (2017) programming contest.

At a strategic level there's not much to the bot. We have a few principles:

* Don't crash our ships into each other.
* Send new ships to the nearest target, which is either a planet that *needs help* (either not fully owned by us, or under threat) or an enemy ship.
* When at a planet, attack the nearest ship that needs attacking, if there are any...
* ...otherwise dock if possible.
* ...otherwise find the nearest target (as above) and go there.
* The above rules imply: we don't dock while there are enemy ships around (which is suicidal).
* In 1v1, we sometimes just rush the enemy.

That's about it - everything else is implementation details.

# 1v1 Genetic Algorithm

In 2-player rush scenarios, where both players send their ships to attack at the start, the bot uses a genetic algorithm to find which moves are best, i.e. we generate a random "genome" (list of moves) and then do the following:

* Mutate the genome randomly.
* Simulate the results.
* If the new genome is better, keep it, otherwise discard.
* Repeat.

The problem is, the simulation needs to know what the opponent will do. I currently do very crude guessing. A more advanced technique would be to *evolve* the opponent's moves as above, and then evolve our own moves against that.

# Global strategy - Conceptual breakthroughs

I think it's fair to say that minor fiddling is less important than overall strategy. I only made a couple of conceptual breakthroughs:

Firstly, never dock when there are enemy ships around. This is obvious in hindsight. Docking is basically suicide. Instead, fight whatever enemy ships are in the area.

Secondly, it's OK to send ships to target interplanetary enemy ships. (Before v23, I only sent ships to planets, then dealt with whatever enemies I found there. But actually targetting interplanetary ships increased mu by about 3 or 4. Such ships are up to no good and need destroying. More than that, you can get some nice clustering behaviour, with your ships merging together as they kill 1 enemy, and then they will help each other as they go on to the next target.
