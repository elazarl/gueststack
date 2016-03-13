#define pr_fmt(fmt) KBUILD_MODNAME ": " fmt
#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/nmi.h>
#include <linux/kvm_host.h>
#include <linux/percpu-defs.h>
#include <linux/perf_event.h>
#include "debugfs.h"

#define PFERR_PRESENT_MASK (1U << 0)
#define PFERR_WRITE_MASK (1U << 1)
#define PFERR_USER_MASK (1U << 2)
#define PFERR_RSVD_MASK (1U << 3)
#define PFERR_FETCH_MASK (1U << 4)

static inline unsigned long kvm_register_read(struct kvm_vcpu *vcpu,
					      enum kvm_reg reg)
{
	if (!test_bit(reg, (unsigned long *)&vcpu->arch.regs_avail))
		kvm_x86_ops->cache_reg(vcpu, reg);

	return vcpu->arch.regs[reg];
}

gpa_t kvm_mmu_gva_to_gpa_read(struct kvm_vcpu *vcpu, gva_t gva,
			      struct x86_exception *exception)
{
	u32 access = (kvm_x86_ops->get_cpl(vcpu) == 3) ? PFERR_USER_MASK : 0;
	return vcpu->arch.walk_mmu->gva_to_gpa(vcpu, gva, access, exception);
}

static inline int buf_printf(struct buf *b, const char *fmt, ...)
{
	va_list vl;
	int len;
	va_start(vl, fmt);
	len = vsnprintf(b->data + b->len, b->cap - b->len, fmt, vl);
	if (b->len + len > b->cap) {
		va_end(vl);
		return 0;
	}
	b->len += len;
	va_end(vl);
	return 1;
}

#define BUF_PRINTF(...)                                                        \
	{                                                                      \
		if (!buf_printf(__VA_ARGS__))                                  \
			return 1;                                              \
	}

DEFINE_PER_CPU(u64[PAGE_SIZE / sizeof(u64)], frames_buf);

int addr_relevant(u64 addr)
{
	int i = 0;
	for (; i < MAX_ELF && text_begin[i] != 0; i++) {
		if (addr < text_end[i] && addr > text_begin[i])
			return 1;
	}
	return 0;
}

static u64 vcpu_offset;

struct kvm_vcpu *__get_current_vcpu(void)
{
	return *this_cpu_ptr((struct kvm_vcpu **)vcpu_offset);
}

static int perf_event_nmi_handler(unsigned int cmd, struct pt_regs *regs)
{
	struct gueststack_stats *stats = this_cpu_ptr(&gueststack_stats);
	struct kvm_vcpu *vcpu = __get_current_vcpu();
	stats->total++;
	if (vcpu) {
		u64 *frames = this_cpu_ptr(frames_buf);
		int i, page_remainder;
		struct x86_exception exception = {};
		struct buf *b = this_cpu_ptr(&stack_buf);
		gva_t rsp = kvm_register_read(vcpu, VCPU_REGS_RSP);
		gpa_t phys_rsp = kvm_mmu_gva_to_gpa_read(vcpu, rsp, &exception);

		stats->relevant++;

		page_remainder = (phys_rsp | (PAGE_SIZE - 1)) - phys_rsp;
		if ((i = kvm_read_guest(vcpu->kvm, phys_rsp, frames,
					page_remainder)) < 0) {
			pr_warn("can't read kernel page. ERROR %d\n", i);
			return 1;
		}
		BUF_PRINTF(b, "CPU:%d RIP: %llx\n", vcpu->vcpu_id,
			   kvm_register_read(vcpu, VCPU_REGS_RIP));
		for (i = 0; i < page_remainder / sizeof(u64); i++) {
			if (addr_relevant(frames[i])) {
				BUF_PRINTF(b, "%llx\n", frames[i]);
			}
		}
	}

	return 1;
}

#define IS_RX_W(x) (((x) | 0x7) == 0x4f)
#define MOD(x) ((x) >> 6)
#define RM(x) ((x) & 0x7)
#define REG(x) (((x) >> 3) & 0x7)
#define BASE(x) ((x) & 0x7)
#define INDEX(x) (((x) >> 3) & 0x7)

/* Search for instruction that moves data form GS to a register */
/* either with displacement mov gs:($1234). reg      */
/* or with mov gs:$1234(rip), reg                    */
/*             GS RX MOV MOD.RM SIB DISPLACEMENT     */
/* For example 65 48 8b  3c     25  80 41 01 00      */
/*                       ^                           */
/*                   mod  reg rm                     */
/*                   0b00 111 100                    */
/*             mov    r15,QWORD PTR gs:0x14180       */
/* Or                                                */
/*             GS RX MOV MOD.RM  DISPLACEMENT        */
/*             65 48 8b  3d      0c 00 00 00         */
/* 1f 44 00 00 65 48 8b  3d      e3 b1 50 3f         */
/*                       ^                           */
/*                   mod  reg rm                     */
/*                   0b00 111 101                    */
/*             mov    0xc(%rip),%rdi                 */
u64 search_relevant_prefix(void *c, int size, bool *found)
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
	*found = false;
	for (;;) {
		u8 *p;
		u8 *rip;
		c = memchr(c, GS_SEG_OVERRIDE, end - c);
		if (c == NULL)
			return -1;
		rip = c;
		c++;
		p = c;
		if (!IS_RX_W(*p))
			continue;
		p++;
		if (*p != MOV_M_TO_R_OPCODE)
			continue;
		/* We need direct access to memory with displacement */
		/* Don't care which registers are used */
		p++;
		if (MOD(*p) != 0)
			continue;
		if (RM(*p) == 0b101) {
			int instruction_len;
			s32 displacement;
			p++;
			/* return rip+displacement32 value+instruction_len */
			displacement = *(s32 *)p;
			/* add displacement length of 4 bytes */
			instruction_len = (p + 4) - rip;
			*found = true;
			return (u64)rip + displacement + instruction_len;
		} else if (RM(*p) == 0b100) {
			// in case of MOD/RM 0/100 we need the SIB byte
			p++;
			if (BASE(*p) != 0b101 || INDEX(*p) != 0b100)
				continue;
			p++;
			/* grab displacement32 value */
			*found = true;
			return *(u32 *)p;
		}
	}
}

void kvm_before_handle_nmi(struct kvm_vcpu *vcpu);

int init_module(void)
{
	u64 offset;
	bool found;
	struct kvm_vcpu *vcpu;
	struct kvm_vcpu *vcpu_stencil = (struct kvm_vcpu *)0xDEAD1BEEF;

	/* Find out the offset of the current_vcpu percpu variable */
	char *kvm_is_user_mode_ptr =
	    (char *)kallsyms_lookup_name("kvm_is_user_mode");
	if (kvm_is_user_mode_ptr == 0) {
		pr_err("Cannot find kvm_is_user_mode symbol. Is kvm "
		       "module loaded?");
		return -2;
	}

	offset = search_relevant_prefix(kvm_is_user_mode_ptr, 20, &found);
	if (!found) {
		pr_err("Cannot find current_vcpu offset: %*ph\n", 20,
		       kvm_is_user_mode_ptr);
		return -2;
	}
	vcpu_offset = offset;

	/* sanity check, hope no PMI interrupt would occur */
	kvm_before_handle_nmi(vcpu_stencil);
	/* compiler can't know __get_current_vcpu snatches what
	 * kvm_before_handle_nmi writes */
	barrier();
	vcpu = __get_current_vcpu();
	barrier();
	kvm_before_handle_nmi(NULL);
	if (vcpu != vcpu_stencil) {
		pr_err("Couldn't read current_vcpu percpu variable (expected "
		       "%p have %p)\n"
		       "Try reloading without PMI (e.g., perf running)\n"
		       "Maybe your kernel is not supported\n",
		       vcpu_stencil, vcpu);
		return -1;
	}

	register_nmi_handler(NMI_LOCAL, perf_event_nmi_handler, 0,
			     "gueststack_pmi");
	if (!init_debugfs())
		goto unregister;
	pr_info("Loaded\n");
	return 0;
unregister:
	unregister_nmi_handler(NMI_LOCAL, "gueststack_pmi");
	return -1;
}

void cleanup_module(void)
{
	unregister_nmi_handler(NMI_LOCAL, "gueststack_pmi");
	remove_debugfs();
	pr_info("Removed\n");
}
