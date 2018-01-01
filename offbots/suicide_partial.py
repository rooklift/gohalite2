pid = int(input())

ship_ids = [pid * 3, pid * 3 + 1, pid * 3 + 2]

for n in range(2):
	input()

print("test")

turn = -1

while 1:
	input()
	turn += 1

	if turn == 0:
		print("t {} 7 45 t {} 7 315".format(ship_ids[1], ship_ids[2]))
	else:
		print()
