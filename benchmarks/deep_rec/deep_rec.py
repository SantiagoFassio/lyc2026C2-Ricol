# Deeply nested linear recursion, close to the Ricol limit of 10000 nested calls
import sys

sys.setrecursionlimit(20000)


def sum_to(n):
    if n == 0:
        return 0
    return n + sum_to(n - 1)


total = 0
k = 0
while k < 300:
    total = total + sum_to(9000 + k % 900)
    k = k + 1
print(total)
