#include "mathkit.h"

int mathkit_add(int a, int b) { return a + b; }

int mathkit_clamp(int v, int lo, int hi) {
    if (v < lo) return lo;
    if (v > hi) return hi;
    return v;
}

unsigned int mathkit_hash(unsigned int seed, unsigned int v) {
    return (seed ^ (v + 0x9e3779b9u)) * 16777619u;
}
