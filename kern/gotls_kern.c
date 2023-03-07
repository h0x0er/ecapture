// Copyright 2022 CFC4N <cfc4n.cs@gmail.com>. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/* Copyright © 2022 Hengqi Chen */
#include "ecapture.h"
#include "gotls.h"
#include "tc.h"

#define GOTLS_RANDOM_SIZE 32

// max length is "CLIENT_HANDSHAKE_TRAFFIC_SECRET"=31
#define MASTER_SECRET_KEY_LEN 32
#define EVP_MAX_MD_SIZE 64

struct go_tls_event {
    u64 ts_ns;
    u32 pid;
    u32 tid;
    s32 data_len;
    char comm[TASK_COMM_LEN];
    char data[MAX_DATA_SIZE_OPENSSL];
};

struct mastersecret_gotls_t {
    u8 lable[MASTER_SECRET_KEY_LEN];
    u8 client_random[GOTLS_RANDOM_SIZE];
    u8 secret_[EVP_MAX_MD_SIZE];
};

/////////////////////////BPF MAPS ////////////////////////////////

// bpf map
struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} mastersecret_go_events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, u64);
    __type(value, struct go_tls_event);
    __uint(max_entries, 2048);
} gte_context SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, u32);
    __type(value, struct go_tls_event);
    __uint(max_entries, 1);
} gte_context_gen SEC(".maps");

static __always_inline struct go_tls_event *get_gotls_event() {
    u32 zero = 0;
    struct go_tls_event *event = bpf_map_lookup_elem(&gte_context_gen, &zero);
    if (!event) return 0;

    u64 id = bpf_get_current_pid_tgid();
    event->ts_ns = bpf_ktime_get_ns();
    event->pid = id >> 32;
    event->tid = (__u32)id;
    bpf_get_current_comm(event->comm, sizeof(event->comm));
    bpf_map_update_elem(&gte_context, &id, event, BPF_ANY);
    return bpf_map_lookup_elem(&gte_context, &id);
}

static __always_inline int gotls_text(struct pt_regs *ctx, bool is_register_abi) {
    s32 record_type, len;
    const char *str;
    void * record_type_ptr;
    void * len_ptr;
    record_type_ptr = (void *)go_get_argument(ctx, is_register_abi, 2);
    bpf_probe_read_kernel(&record_type, sizeof(record_type), (void *)&record_type_ptr);
    str = (void *)go_get_argument(ctx, is_register_abi, 3);
    len_ptr = (void *)go_get_argument(ctx, is_register_abi, 4);
    bpf_probe_read_kernel(&len, sizeof(len), (void *)&len_ptr);

    debug_bpf_printk("gotls_text record_type:%d\n", record_type);
    if (record_type != recordTypeApplicationData) {
        return 0;
    }

    struct go_tls_event *event = get_gotls_event();
    if (!event) {
        return 0;
    }

    event->data_len = len;
    int ret = bpf_probe_read_user(&event->data, sizeof(event->data), (void*)str);
    if (ret < 0) {
        debug_bpf_printk(
            "gotls_text bpf_probe_read_user_str failed, ret:%d, str:%d\n", ret,
            str);
        return 0;
    }
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, event,
                          sizeof(struct go_tls_event));
    return 0;
}

// capture golang tls plaintext, supported golang stack-based ABI (go version >= 1.17)
// type recordType uint8
// writeRecordLocked(typ recordType, data []byte)
SEC("uprobe/gotls_text_register")
int gotls_text_register(struct pt_regs *ctx) {
    return gotls_text(ctx, true);
}

// capture golang tls plaintext, supported golang stack-based ABI (go version < 1.17)
// type recordType uint8
// writeRecordLocked(typ recordType, data []byte)
SEC("uprobe/gotls_text_stack")
int gotls_text_stack(struct pt_regs *ctx) {
    return gotls_text(ctx, false);
}


/*
* crypto/tls/common.go
* func (c *Config) writeKeyLog(label string, clientRandom, secret []byte) error
*/
static __always_inline int gotls_masterkey(struct pt_regs *ctx, bool is_register_abi) {
    const char *label, *clientrandom, *secret;
    void *lab_ptr, *cr_ptr, *secret_ptr;
    lab_ptr = (void *)go_get_argument(ctx, is_register_abi, 2);
    cr_ptr = (void *)go_get_argument(ctx, is_register_abi, 3);
    secret_ptr = (void *)go_get_argument(ctx, is_register_abi, 4);
    struct mastersecret_gotls_t mastersecret_gotls={};
    int ret = bpf_probe_read_user(&mastersecret_gotls.lable, sizeof(mastersecret_gotls.lable), (void*)lab_ptr);
    if (ret < 0) {
        debug_bpf_printk(
            "gotls_masterkey read mastersecret lable failed, ret:%d, str:%d\n", ret,
            str);
        return 0;
    }

    ret = bpf_probe_read_user(&mastersecret_gotls.client_random, sizeof(mastersecret_gotls.client_random), (void*)cr_ptr);
    if (ret < 0) {
        debug_bpf_printk(
            "gotls_masterkey read mastersecret client_random failed, ret:%d, str:%d\n", ret,
            str);
        return 0;
    }

    ret = bpf_probe_read_user(&mastersecret_gotls.secret_, sizeof(mastersecret_gotls.secret_), (void*)secret_ptr);
    if (ret < 0) {
        debug_bpf_printk(
            "gotls_masterkey read mastersecret client_random failed, ret:%d, str:%d\n", ret,
            str);
        return 0;
    }

    bpf_perf_event_output(ctx, &mastersecret_go_events, BPF_F_CURRENT_CPU,
                              &mastersecret_gotls, sizeof(struct mastersecret_gotls_t));
    return 0;
}

SEC("uprobe/gotls_masterkey_register")
int gotls_masterkey_register(struct pt_regs *ctx) {
    return gotls_masterkey(ctx, true);
}

SEC("uprobe/gotls_masterkey_stack")
int gotls_masterkey_stack(struct pt_regs *ctx) {
    return gotls_masterkey(ctx, false);
}