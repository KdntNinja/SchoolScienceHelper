from random import randint as rint

def dice() -> int:
	num1: int = rint(0,6)
	num2: int = rint(0,6)
	total: int = num1 + num2
	return total

if __name__ == "__main__":
	number: int = dice()
	print(number)
