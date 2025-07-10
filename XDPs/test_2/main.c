#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

SEC("xdp")
int  xdp_prog(struct xdp_md *ctx)
{
	int x = 0;
	int y = 4;
	while((x=y)){
		x++;
	}
	return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
