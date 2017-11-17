import json, random, subprocess

processes = [
	"bot.exe",
	"bot.exe",
	# ".\\otherbots\\v32\\mybot.exe",
	# ".\\otherbots\\v32\\mybot.exe",
	".\\offbots\\brine\\brine.exe",
	".\\offbots\\brine\\brine.exe",
]

scores = [0,0,0,0]

positions = [0,1,2,3]

print("{} --- {} --- {} --- {}".format(processes[0], processes[1], processes[2], processes[3]))

while 1:

	random.shuffle(positions)

	output = subprocess.check_output(
		"halite.exe --no-compression -q \"{}\" \"{}\" \"{}\" \"{}\"".format(
			processes[positions[0]], processes[positions[1]], processes[positions[2]], processes[positions[3]]
			)).decode("ascii")

	result = json.loads(output)

	for key in result["stats"]:
		rank = result["stats"][key]["rank"]
		i = positions[int(key)]

		if rank == 1:
			scores[i] += 1

	print(scores)

