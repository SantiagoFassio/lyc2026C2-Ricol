// Loop whose arithmetic mixes variables with subexpressions made only of literals
#include <stdio.h>

int main(void) {
    long long i = 0, acc = 0;
    while (i < 2000 * 1000) {
        acc = (acc + i * (60 * 60 * 24) + (1000 - 7 * 3) - -5) % 1000003;
        i = i + 1;
    }
    printf("%lld\n", acc);
    return 0;
}
