// Deeply nested linear recursion, close to the Ricol limit of 10000 nested calls
#include <stdio.h>

long long sum(long long n) {
    if (n == 0) {
        return 0;
    }
    return n + sum(n - 1);
}

int main(void) {
    long long total = 0, k = 0;
    while (k < 300) {
        total = total + sum(9000 + k % 900);
        k = k + 1;
    }
    printf("%lld\n", total);
    return 0;
}
