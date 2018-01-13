import json, requests, sys, time


LIMIT = 250
MY_ID = 1278

games = []
known_game_ids = set()


def get_games():

	global games
	global known_game_ids

	offset = len(games)

	while 1:

		print("Getting from offset {}... ".format(offset), end="")
		sys.stdout.flush()

		some = requests.get("http://api.halite.io/v1/api/user/{}/match?&limit={}&offset={}".format(MY_ID, LIMIT, offset)).json()

		if len(some) == 0:
			print("empty")
			return

		for game in some:
			if game["game_id"] not in known_game_ids:
				games.append(game)
				known_game_ids.add(game["game_id"])
			else:
				return

		print("done")
		time.sleep(5)
		offset += LIMIT


def main():

	global games
	global known_game_ids

	print("Loading games.json")
	try:
		with open("games.json") as data:
			games = json.load(data)
		print("OK")
	except:
		games = []
		print("Failed")


	for game in games:
		known_game_ids.add(game["game_id"])


	get_games()

	print("Overwriting games.json")

	with open("games.json", "w") as outfile:
		outfile.write(json.dumps(games, indent=2))

	print("Done")


main()
