import json, random, subprocess

processes = [
	"bot.exe --conservative",
	".\\offbots\\brine\\brine.exe --conservative",
]

scores = [0,0]

positions = [0,1]

print("{} --- {}".format(processes[0], processes[1]))

while 1:

	random.shuffle(positions)

	output = subprocess.check_output(
		"halite.exe --no-compression -q \"{}\" \"{}\"".format(processes[positions[0]], processes[positions[1]])
		).decode("ascii")

	result = json.loads(output)

	for key in result["stats"]:
		rank = result["stats"][key]["rank"]
		i = positions[int(key)]

		if rank == 1:
			scores[i] += 1

	print(scores)

