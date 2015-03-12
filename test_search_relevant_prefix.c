
#include <stdio.h>
#include <stdint.h>
#include <string.h>

typedef int64_t s64;
typedef uint32_t u32;
typedef uint8_t u8;

#define IS_RX_W(x) (((x) | 0x7) == 0x4f)
#define MOD(x) ((x) >> 6)
#define RM(x) ((x)&0x7)
#define REG(x) (((x) >> 3) & 0x7)
#define BASE(x) ((x)&0x7)
#define INDEX(x) (((x) >> 3) & 0x7)

/* Search for instruction that moves data form GS to a register */
/*             GS RX MOV MOD.RM SIB DISPLACEMENT     */
/* For example 65 48 8b  3c     25  80 41 01 00      */
/*             mov    r15,QWORD PTR gs:0x14180       */
s64 search_relevant_prefix(void *c, int size)
{
	/* sought instruction is:
	 * - GS segment override prefix
	 * - RX.W prefix to use 64 bit
	 * - MOV opcode
	 * - MOD/RM + SIB
	 * - 4 bytes of displacement address
	 * at least nine bytes.
	 */
	const int inst_len = 9;
	const int GS_SEG_OVERRIDE = 0x65;
	const int MOV_M_TO_R_OPCODE = 0x8B;
	void *end = c + size - inst_len + 1;
	for (;;) {
		u8 *p;
		c = memchr(c, GS_SEG_OVERRIDE, end - c);
		if (c == NULL)
			return -1;
		c++;
		p = c;
		if (!IS_RX_W(*p))
			continue;
		p++;
		if (*p != MOV_M_TO_R_OPCODE)
			continue;
		p++;
		if (MOD(*p) != 0 || RM(*p) != 0b100)
			continue;
		p++;
		if (BASE(*p) != 0b101 || INDEX(*p) != 0b100)
			continue;
		p++;
		/* grab displacement32 value */
		return *(u32 *)p;
	}
}

int main(int argc, char **argv)
{
	unsigned char b[] = {0,    0,    0x65, 0x48, 0x8b, 0x3c,
			     0x25, 0x80, 0x41, 0x01, 0x00};
	unsigned char nothing[] = {0,    0,    0x65, 0x48, 0x8b, 0x3c,
				   0x15, 0x80, 0x41, 0x01, 0x00};
	printf("%lx\n", (int64_t)search_relevant_prefix(b, sizeof(b)));
	printf("%lx\n", (int64_t)search_relevant_prefix(b + 2, sizeof(b) - 2));
	printf("%lx\n",
	       (int64_t)search_relevant_prefix(nothing, sizeof(nothing)));
	return 0;
}
