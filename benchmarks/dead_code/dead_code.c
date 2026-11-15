// Loop with branches that become dead after folding their literal conditions
#include <stdbool.h>
#include <stdio.h>

int main(void) {
    long long i = 0, acc = 0;
    while (i < 2000000) {
        if (1 > 2) {
            acc = acc * 2;
        }
        if (2 * 3 == 6) {
            acc = (acc + i) % 1000003;
        } else {
            acc = 0;
        }
        while (false) {
            acc = acc + 1;
        }
        i = i + 1;
    }
    printf("%lld\n", acc);
    return 0;
}
