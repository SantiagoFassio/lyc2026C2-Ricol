// Nested loops with conditions: counts the primes below N by trial division
#include <stdbool.h>
#include <stdio.h>

bool is_prime(long long n) {
    if (n < 2) {
        return false;
    }
    long long d = 2;
    while (d * d <= n) {
        if (n % d == 0) {
            return false;
        }
        d = d + 1;
    }
    return true;
}

int main(void) {
    long long count = 0, n = 0;
    while (n < 200000) {
        if (is_prime(n)) {
            count = count + 1;
        }
        n = n + 1;
    }
    printf("%lld\n", count);
    return 0;
}
