My bot in Go for the [Halite 2](https://halite.io/) (2017) programming contest.

At a strategic level there's not much to the bot. We have a few principles:

* Don't crash our ships into each other.
* Send new ships to the nearest planet that *needs help* (either not fully owned by us, or under threat).
* When at a planet, attack the nearest ship that needs attacking, if there are any...
* ...otherwise dock if possible.
* ...otherwise find the nearest planet that *needs help* and go there.
* Don't dock while there are enemy ships around.
* In 1v1, we sometimes just rush the enemy.

That's about it - everything else is implementation details.
