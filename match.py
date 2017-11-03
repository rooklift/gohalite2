import json, subprocess

process1 = "bot.exe"
process2 = "bot.exe"
process3 = ".\\otherbots\\v19\\mybot.exe"
process4 = ".\\otherbots\\v19\\mybot.exe"

scores = [0,0,0,0]

print("{} --- {} --- {} --- {}".format(process1, process2, process3, process4))

while 1:

	output = subprocess.check_output(
		"halite.exe -r -q -i \"replays\" \"{}\" \"{}\" \"{}\" \"{}\"".format(
			process1, process2, process3, process4
			)).decode("ascii")

	result = json.loads(output)

	for key in result["stats"]:
		rank = result["stats"][key]["rank"]
		pid = int(key)

		if rank == 1:
			scores[pid] += 1

	print(scores)

