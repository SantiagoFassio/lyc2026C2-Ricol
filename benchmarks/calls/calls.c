// Many calls to a short function: measures the cost of each call
#include <stdio.h>

__attribute__((noinline)) long long step(long long acc, long long i) {
    return (acc + i * 7) % 1000003;
}

int main(void) {
    long long acc = 0, i = 0;
    while (i < 1000000) {
        acc = step(acc, i);
        i = i + 1;
    }
    printf("%lld\n", acc);
    return 0;
}
