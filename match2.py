import json, random, subprocess

processes = [
	"bot.exe --conservative",
	".\\otherbots\\v63\\mybot.exe --conservative",
]

WIDTH = 360
HEIGHT = 240

scores = [0,0]

positions = [0,1]

print("{} --- {}".format(processes[0], processes[1]))

while 1:

	random.shuffle(positions)

	output = subprocess.check_output(
		"halite.exe -d \"{} {}\" --no-compression -q \"{}\" \"{}\"".format(WIDTH, HEIGHT, processes[positions[0]], processes[positions[1]])
		).decode("ascii")

	result = json.loads(output)

	for key in result["stats"]:
		rank = result["stats"][key]["rank"]
		i = positions[int(key)]

		if rank == 1:
			scores[i] += 1

	print(scores)

