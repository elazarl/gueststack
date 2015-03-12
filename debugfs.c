#define pr_fmt(fmt) KBUILD_MODNAME ": " fmt
#include <linux/kernel.h>
#include <linux/slab.h>
#include <linux/debugfs.h>
#include <linux/module.h>
#include <linux/fs.h>
#include <linux/mutex.h>
#include <linux/uaccess.h>
#include "debugfs.h"

struct dentry *root = NULL;

u64 text_begin[MAX_ELF] = {0};
u64 text_end[MAX_ELF] = {0};
DEFINE_PER_CPU(struct buf, stack_buf);
DEFINE_PER_CPU(struct gueststack_stats, gueststack_stats);

static int read_buffer_open(struct inode *inode, struct file *file)
{
	file->private_data = inode->i_private;
	return 0;
}

static ssize_t read_buffer(struct file *file, char __user *buf, size_t len,
			   loff_t *offset)
{
	struct buf *b = file->private_data;
	return simple_read_from_buffer(buf, len, offset, b->data, b->len);
}

ssize_t reset_buffer(struct file *file, const char __user *buf, size_t len,
		     loff_t *ppos)
{
	struct buf *b = file->private_data;
	b->len = 0;
	return len;
}

static const struct file_operations buffer_file_ops = {
    .open = read_buffer_open,
    .read = read_buffer,
    .write = reset_buffer,
    .llseek = default_llseek,
};

DEFINE_MUTEX(relevant_addr_mutex);

ssize_t write_relevant_addr_ops(struct file *file, const char __user *ubuf,
				size_t len, loff_t *ppos)
{
	char *p;
	char *buf;
	if (len >= PAGE_SIZE)
		return -EINVAL;
	buf = (char *)__get_free_page(GFP_TEMPORARY);
	if (!buf)
		return -ENOMEM;

	if (copy_from_user(buf, ubuf, len)) {
		free_page((unsigned long)buf);
		return -EFAULT;
	}
	buf[len] = '\0';
	mutex_lock(&relevant_addr_mutex);
	p = buf;
	for (; *p != '\0' && *ppos < MAX_ELF - 1; (*ppos)++) {
		int i = *ppos;
		char *eol = strchr(p, '\n');
		char *dash;
		if (eol == NULL)
			goto error;
		*eol = '\0';
		dash = strchr(p, '-');
		if (dash == NULL)
			goto error;
		*dash = '\0';
		dash++;
		if (kstrtou64(p, 16, text_begin + i) != 0) {
			pr_warn("illegal number to addr_elevant %s\n", p);
			goto error;
		}
		if (kstrtou64(dash, 16, text_end + i) != 0) {
			pr_warn("illegal number to addr_elevant %s\n", dash);
			goto error;
		}
		p = eol + 1;
	}
	text_begin[*ppos] = 0;
	text_end[*ppos] = 0;
	pr_warn("reset addr_relevant %d: %llu-%llu\n", (int)*ppos,
		text_begin[*ppos], text_end[*ppos]);
	mutex_unlock(&relevant_addr_mutex);
	return len;
error:
	mutex_unlock(&relevant_addr_mutex);
	return -EINVAL;
}

static ssize_t read_relevant_addr_ops(struct file *filp, char __user *ubuf,
				      size_t cnt, loff_t *ppos)
{
	char buf[MAX_ELF * (64 / 8 * 2 + 1) + 1];
	int i;
	int pos = 0;
	mutex_lock(&relevant_addr_mutex);
	for (i = 0; text_begin[i] != 0 && i < MAX_ELF; i++) {
		pos += snprintf(buf + pos, sizeof(buf) - pos, "%llx-%llx\n",
				text_begin[i], text_end[i]);
	}
	mutex_unlock(&relevant_addr_mutex);
	return simple_read_from_buffer(ubuf, cnt, ppos, buf, pos);
}

static const struct file_operations relevant_addr_ops = {
    //.open = open_relevant_addr_ops,
    .read = read_relevant_addr_ops,
    .write = write_relevant_addr_ops,
    .llseek = default_llseek,
};

int init_debugfs(void)
{
	int cpu;
	root = debugfs_create_dir("gueststack", NULL);
	if (root == NULL)
		return 0;
	for_each_possible_cpu(cpu)
	{
		char name[50];
		struct buf *b = &per_cpu(stack_buf, cpu);
		struct gueststack_stats *stats =
		    &per_cpu(gueststack_stats, cpu);
		b->cap = STACK_BUF_SIZE;
		b->data = kmalloc(b->cap, GFP_KERNEL);
		b->len = 0;
		snprintf(name, sizeof(name), "stack%02d", cpu);
		if (debugfs_create_file(name, 0666, root, b,
					&buffer_file_ops) == NULL)
			goto cleanup;
		snprintf(name, sizeof(name), "total_events%02d", cpu);
		if (debugfs_create_u64(name, 0666, root, &stats->total) == NULL)
			goto cleanup;
		snprintf(name, sizeof(name), "relevant_events%02d", cpu);
		if (debugfs_create_u64(name, 0666, root, &stats->relevant) ==
		    NULL)
			goto cleanup;
	}
	if (debugfs_create_file("relevant_addr", 0666, root, NULL,
				&relevant_addr_ops) == NULL)
		goto cleanup;
	return 1;
cleanup:
	debugfs_remove_recursive(root);
	return 0;
}

void remove_debugfs(void)
{
	int cpu;
	if (root)
		debugfs_remove_recursive(root);
	for_each_possible_cpu(cpu)
	{
		struct buf *b = &per_cpu(stack_buf, cpu);
		kfree(b->data);
	}
}
MODULE_LICENSE("GPL");
