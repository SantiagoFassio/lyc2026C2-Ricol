# Nested loops with conditions: counts the primes below N by trial division
def is_prime(n):
    if n < 2:
        return False
    d = 2
    while d * d <= n:
        if n % d == 0:
            return False
        d = d + 1
    return True


count = 0
n = 0
while n < 200000:
    if is_prime(n):
        count = count + 1
    n = n + 1
print(count)
