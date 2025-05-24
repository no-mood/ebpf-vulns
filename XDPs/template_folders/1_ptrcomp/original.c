// SPDX-License-Identifier: (GPL-2.0 OR BSD-2-Clause
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
//#include <linux/types.h>
#include "tools/jhash.h"
//#include <linux/if_ether.h>
#define MAX_SERVERS 512
#define IP_FRAGMENTED 0x3FFF

#ifndef ETH_P_IP
#define ETH_P_IP 0x0800
#endif

char LICENSE[] SEC("license") = "Dual BSD/GPL";

struct pkt_meta {
    __be32 src;
    __be32 dst;
    union {
        __u32 ports;
        __u16 port16[2];
    };
};

struct dest_info {
    __u32 saddr;
    __u32 daddr;
    __u64 bytes;
    __u64 pkts;
    __u8 dmac[6];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_SERVERS);
    __type(key, __u32);
    __type(value, struct dest_info);
} servers SEC(".maps");

static __always_inline struct dest_info *hash_get_dest(struct pkt_meta *pkt)
{
    __u32 key = jhash_2words(pkt->src, pkt->ports, MAX_SERVERS) % MAX_SERVERS;
    struct dest_info *tnl = bpf_map_lookup_elem(&servers, &key);

    if (!tnl) {
        key = 0;
        tnl = bpf_map_lookup_elem(&servers, &key);
    }
    return tnl;
}

static __always_inline bool parse_udp(void *data, __u64 off, void *data_end,
                                      struct pkt_meta *pkt)
{
    struct udphdr *udp = data + off;
    if ((void *)(udp + 1) > data_end)
        return false;

    pkt->port16[0] = udp->source;
    pkt->port16[1] = udp->dest;
    return true;
}

static __always_inline bool parse_tcp(void *data, __u64 off, void *data_end,
                                      struct pkt_meta *pkt)
{
    struct tcphdr *tcp = data + off;
    if ((void *)(tcp + 1) > data_end)
        return false;

    pkt->port16[0] = tcp->source;
    pkt->port16[1] = tcp->dest;
    return true;
}

static __always_inline void set_ethhdr(struct ethhdr *new_eth,
                                       const struct ethhdr *old_eth,
                                       const struct dest_info *tnl,
                                       __be16 h_proto)
{
    __builtin_memcpy(new_eth->h_source, old_eth->h_dest, sizeof(new_eth->h_source));
    __builtin_memcpy(new_eth->h_dest, tnl->dmac, sizeof(new_eth->h_dest));
    new_eth->h_proto = h_proto;
}

static __always_inline int process_packet(struct xdp_md *ctx, __u64 off)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct pkt_meta pkt = {};
    struct ethhdr *new_eth, *old_eth;
    struct iphdr *iph, iph_tnl;
    struct dest_info *tnl;
    __u16 pkt_size, payload_len;
    __u8 protocol;
    __u32 csum = 0;

    iph = data + off;
    if ((void *)(iph + 1) > data_end || iph->ihl != 5)
        return XDP_DROP;

    protocol = iph->protocol;
    payload_len = bpf_ntohs(iph->tot_len);
    off += sizeof(struct iphdr);

    if (iph->frag_off & bpf_htons(IP_FRAGMENTED))
        return XDP_DROP;

    pkt.src = iph->saddr;
    pkt.dst = iph->daddr;

    if (protocol == IPPROTO_TCP) {
        if (!parse_tcp(data, off, data_end, &pkt))
            return XDP_DROP;
    } else if (protocol == IPPROTO_UDP) {
        if (!parse_udp(data, off, data_end, &pkt))
            return XDP_DROP;
    } else {
        return XDP_PASS;
    }

    tnl = hash_get_dest(&pkt);
    if (!tnl)
        return XDP_DROP;

    if (bpf_xdp_adjust_head(ctx, 0 - (int)sizeof(struct iphdr)))
        return XDP_DROP;

    data = (void *)(long)ctx->data;
    data_end = (void *)(long)ctx->data_end;

    new_eth = data;
    old_eth = data + sizeof(struct iphdr);
    if ((void *)(new_eth + 1) > data_end || (void *)(old_eth + 1) > data_end)
        return XDP_DROP;

    set_ethhdr(new_eth, old_eth, tnl, bpf_htons(ETH_P_IP));

    iph_tnl.version = 4;
    iph_tnl.ihl = sizeof(struct iphdr) >> 2;
    iph_tnl.frag_off = 0;
    iph_tnl.protocol = IPPROTO_IPIP;
    iph_tnl.check = 0;
    iph_tnl.id = 0;
    iph_tnl.tos = 0;
    iph_tnl.tot_len = bpf_htons(payload_len + sizeof(struct iphdr));
    iph_tnl.daddr = tnl->daddr;
    iph_tnl.saddr = tnl->saddr;
    iph_tnl.ttl = 8;

    __u16 *next_iph_u16 = (__u16 *)&iph_tnl;
#pragma clang loop unroll(full)
    for (int i = 0; i < (int)sizeof(struct iphdr) >> 1; i++)
        csum += *next_iph_u16++;
    iph_tnl.check = ~((csum & 0xffff) + (csum >> 16));

    iph = data + sizeof(struct ethhdr);
    *iph = iph_tnl;

    pkt_size = (__u16)(data_end - data);
    __sync_fetch_and_add(&tnl->pkts, 1);
    __sync_fetch_and_add(&tnl->bytes, pkt_size);

    return XDP_TX;
}

SEC("xdp")
int loadbal(struct xdp_md *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct ethhdr *eth = data;

    if ((void *)(eth + 1) > data_end)
        return XDP_DROP;

    __u16 eth_proto = bpf_ntohs(eth->h_proto);
    if (eth_proto == ETH_P_IP)
        return process_packet(ctx, sizeof(struct ethhdr));

    return XDP_PASS;
}

