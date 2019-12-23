#include "cryptonight/hash-ops.h"

void cryptonightbbcslow(char *input, int size, char *output, int variant, int prehashed)
{
	cn_slow_hash(input, size, output, variant, prehashed, 0);
}
