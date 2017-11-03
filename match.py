import json, subprocess

scores = [0,0,0,0]

while 1:

	output = subprocess.check_output(
		"halite.exe -r -q -i \"replays\" \"bot.exe\" \".\\otherbots\\v18\\mybot.exe\" \".\\otherbots\\v18\\mybot.exe\" \"bot.exe\"").decode("ascii")

	result = json.loads(output)

	for key in result["stats"]:
		rank = result["stats"][key]["rank"]
		pid = int(key)

		if rank == 1:
			scores[pid] += 1

	print(scores)

