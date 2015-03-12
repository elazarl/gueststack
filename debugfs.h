#pragma once
int init_debugfs(void);
void remove_debugfs(void);

struct buf
{
	int len;
	int cap;
	void *data;
};

#define MAX_ELF 10
extern u64 text_begin[MAX_ELF];
extern u64 text_end[MAX_ELF];

struct gueststack_stats
{
	u64 total;
	u64 relevant;
};

DECLARE_PER_CPU(struct gueststack_stats, gueststack_stats);

#define STACK_BUF_SIZE PAGE_SIZE * 1000
// PAGE_SIZE/sizeof(u64)
DECLARE_PER_CPU(u64[STACK_BUF_SIZE], stack_buf_buffer);
DECLARE_PER_CPU(struct buf, stack_buf);
