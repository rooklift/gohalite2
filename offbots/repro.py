pid = int(input())

for n in range(2):
	input()

print("test")

turn = -1

while 1:
	input()
	turn += 1

	if pid == 0:
		print()
		continue

	if turn == 0:
		print("t 4 1 271")
	elif turn == 1:
		print("t 4 6 271")
	elif turn >= 2 and turn <= 4:
		print("t 4 7 271")
	elif turn == 5:
		print("d 4 0")
	else:
		print()

