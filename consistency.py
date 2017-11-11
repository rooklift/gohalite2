import subprocess

processes = [
	"bot.exe --conservative",
	".\\otherbots\\v27\\mybot.exe --conservative",
]

n = 0

while 1:
	subprocess.check_output("halite.exe -s {} --no-compression -q \"{}\" \"{}\"".format(n, processes[0], processes[1]))
	subprocess.check_output("halite.exe -s {} --no-compression -q \"{}\" \"{}\"".format(n, processes[1], processes[0]))
	n += 1
	print(n)
