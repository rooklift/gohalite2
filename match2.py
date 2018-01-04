import json, random, subprocess

processes = [
	"bot.exe --conservative",
	".\\otherbots\\v67\\mybot.exe --conservative",
]

FORCED_WIDTH = None
FORCED_HEIGHT = None

# ------------------------------------------------------------------------

map_sizes = [80, 80, 88, 88, 96, 96, 96, 104, 104, 104, 104, 112, 112, 112, 120, 120, 128, 128]

scores = [0,0]
positions = [0,1]

print("{} --- {}".format(processes[0], processes[1]))

while 1:

	random.shuffle(positions)

	if (not FORCED_WIDTH) or (not FORCED_HEIGHT):
		base_size = random.choice(map_sizes)
		width = base_size * 3
		height = base_size * 2
	else:
		width = FORCED_WIDTH
		height = FORCED_HEIGHT

	output = subprocess.check_output(
		"halite.exe -d \"{} {}\" --no-compression -q \"{}\" \"{}\"".format(width, height, processes[positions[0]], processes[positions[1]])
		).decode("ascii")

	result = json.loads(output)

	for key in result["stats"]:
		rank = result["stats"][key]["rank"]
		i = positions[int(key)]

		if rank == 1:
			scores[i] += 1

	print(scores)

