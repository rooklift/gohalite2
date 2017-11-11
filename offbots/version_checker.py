import subprocess

raw = subprocess.check_output("g++ -v", stderr=subprocess.STDOUT, shell=True, timeout=3, universal_newlines=True)
subvers = raw[raw.find(" version ") + 9:].split(" ")[0].split(".")

pid = int(input())

for n in range(2):
	input()

print("version checker")

while 1:
	input()
	print("t {} 0 {} t {} 0 {} t {} 0 {}".format(pid * 3, subvers[0], pid * 3 + 1, subvers[1], pid * 3 + 2, subvers[2]))

