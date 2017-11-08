My bot in Go for the [Halite 2](https://halite.io/) (2017) programming contest.

At a strategic level there's not much to the bot. In most games, the bot acts on these principles:

* Don't crash our ships into each other.
* Send new ships to the nearest target, which is either an enemy ship, or a planet that *needs help*, defined as:
  - Not fully docked by us, or:
  - Under threat of attack.
* When at a planet:
  - Attack the nearest ship that needs attacking, if there are any.
  - Otherwise, dock if possible.
  - Otherwise, choose a new destination as above.

That's about it - everything else is implementation details.

# 1v1 Genetic Algorithm

In 2-player games it's sometimes sensible to rush the enemy, ignoring planets and going straight at 'em. In this case, when the ships are close to the enemy, the bot uses a genetic algorithm to find which moves are best, i.e. we generate a random "genome" (list of moves) and then do the following:

* Mutate the genome randomly.
* Simulate the results.
* If the new genome is better, keep it, otherwise discard.
* Repeat.

The problem is, the simulation needs to know what the opponent will do. I currently do very crude guessing. A more advanced technique would be to evolve the *opponent's* moves as above, and only then evolve our own moves to counteract them.

# Global Strategy - Conceptual Breakthroughs

I suspect implementation details are less important than overall strategy. I only made a couple of conceptual breakthroughs:

* Firstly, never dock when there are enemy ships around. This is obvious in hindsight. Docking is basically suicide. Instead, fight whatever enemy ships are in the area.

* Secondly, it's OK to send ships to target interplanetary enemy ships. Before v23, I only sent ships to planets, then dealt with whatever enemies I found there. But directly targetting interplanetary ships increased mu by about 2 or 3. Such ships are up to no good and need destroying. More than that, you can get some nice clustering behaviour, with your ships coming together as they kill 1 enemy; then they will help each other as they go on to the next target.
