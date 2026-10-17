// AILA 语言运行时（M1 薄层）
//
// 依据 designs/0001 §12.3「薄层四件套」：本文件提供其中两件的宿主实现 +
// 最小 std.io 的打印出口。设计纪律：运行时不抢向量表、不含调度器、不含 GC。
//
// 注意（实现口径，写在这里防止误读）：
//   * 区域在**宿主侧**用「块链表 + bump 记账」实现：region_leave 会把本区域内
//     所有块一次性归还，语义等价于「区域退出整块回收」；嵌入式 target 的
//     O(1) bump 复位由平台层替换本文件实现（接口不变）。
//   * 配额检查（used + size > cap → trap）按 designs/0001 §5.6 的归属配额语义实现。

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#if defined(_MSC_VER)
#define AILA_NORETURN __declspec(noreturn)
#else
#define AILA_NORETURN __attribute__((noreturn))
#endif

typedef struct AilaBlock {
    struct AilaBlock* next;
    size_t usable;
} AilaBlock;

typedef struct AilaRegion {
    AilaBlock*   blocks;
    size_t       used;        // 归属配额计量（本区域累计接收的分配量）
    size_t       cap;         // 0 = 无上限
    struct AilaRegion* prev;
} AilaRegion;

static AilaRegion* g_current_region = NULL;

// --------------------------------------------------------------------------- //
// trap：显式 trap 与编译器插入的隐式陷阱共用同一条通道
// --------------------------------------------------------------------------- //
AILA_NORETURN void aila_rt_trap(const char* msg, int64_t len, const char* file, int64_t line) {
    fflush(stdout);
    fprintf(stderr, "\nAILA trap: ");
    if (msg && len > 0) fwrite(msg, 1, (size_t)len, stderr);
    else fprintf(stderr, "unknown");
    fprintf(stderr, "\n  at %s:%lld\n", file ? file : "<unknown>", (long long)line);
    fflush(stderr);
    exit(70);
}

// --------------------------------------------------------------------------- //
// 区域账本
// --------------------------------------------------------------------------- //
void* aila_rt_region_enter(int64_t cap_hint) {
    AilaRegion* r = (AilaRegion*)malloc(sizeof(AilaRegion));
    if (!r) aila_rt_trap("region: out of memory for region header", 41, __FILE__, __LINE__);
    r->blocks = NULL;
    r->used = 0;
    r->cap = (size_t)(cap_hint > 0 ? cap_hint : 0);
    r->prev = g_current_region;
    g_current_region = r;
    return (void*)r;
}

void* aila_rt_region_alloc(void* region, int64_t size, int64_t align) {
    AilaRegion* r = (AilaRegion*)(region ? region : (void*)g_current_region);
    if (!r) aila_rt_trap("allocation without an active region", 36, __FILE__, __LINE__);
    if (align <= 0) align = 8;
    size_t need = (size_t)(size > 0 ? size : 1);
    if (r->cap && r->used + need > r->cap) {
        aila_rt_trap("region quota exceeded", 23, __FILE__, __LINE__);
    }
    size_t header = sizeof(AilaBlock);
    size_t total = header + need + (size_t)align;
    AilaBlock* blk = (AilaBlock*)malloc(total);
    if (!blk) aila_rt_trap("region: physical memory exhausted", 35, __FILE__, __LINE__);
    blk->next = r->blocks;
    r->blocks = blk;
    uintptr_t base = (uintptr_t)(blk + 1);
    uintptr_t aligned = (base + (uintptr_t)(align - 1)) & ~(uintptr_t)(align - 1);
    r->used += need;
    memset((void*)aligned, 0, need);
    return (void*)aligned;
}

// `@static` = 预置区域裸名：寿命 = 模块生命周期。宿主侧实现为一个进程级常驻区域，
// 永不 leave（嵌入式 target 由链接期段落位实现，接口不变）。
void* aila_rt_static_region(void) {
    static AilaRegion* g_static = NULL;
    if (!g_static) {
        g_static = (AilaRegion*)malloc(sizeof(AilaRegion));
        if (!g_static) aila_rt_trap("static region: out of memory", 28, __FILE__, __LINE__);
        g_static->blocks = NULL;
        g_static->used = 0;
        g_static->cap = 0;
        g_static->prev = NULL;
    }
    return (void*)g_static;
}

void aila_rt_region_leave(void* region) {
    AilaRegion* r = (AilaRegion*)region;
    if (!r) return;
    AilaBlock* b = r->blocks;
    while (b) {
        AilaBlock* nxt = b->next;
        free(b);
        b = nxt;
    }
    if (g_current_region == r) g_current_region = r->prev;
    free(r);
}

// --------------------------------------------------------------------------- //
// 最小 std.io 的打印出口（AILA 侧以 `#[c]` 声明，stdlib/io.aila 里是普通包装函数）
// --------------------------------------------------------------------------- //
void aila_print_i32(int32_t v) { printf("%d", (int)v); }
void aila_print_u32(uint32_t v) { printf("%u", (unsigned)v); }
void aila_print_f64(double v) { printf("%g", v); }
void aila_print_bool(int8_t v) { printf("%s", v ? "true" : "false"); }
void aila_print_nl(void) { printf("\n"); fflush(stdout); }

void aila_print_char(uint32_t cp) {
    unsigned char buf[4];
    int n = 0;
    if (cp < 0x80) { buf[n++] = (unsigned char)cp; }
    else if (cp < 0x800) {
        buf[n++] = (unsigned char)(0xC0 | (cp >> 6));
        buf[n++] = (unsigned char)(0x80 | (cp & 0x3F));
    } else if (cp < 0x10000) {
        buf[n++] = (unsigned char)(0xE0 | (cp >> 12));
        buf[n++] = (unsigned char)(0x80 | ((cp >> 6) & 0x3F));
        buf[n++] = (unsigned char)(0x80 | (cp & 0x3F));
    } else {
        buf[n++] = (unsigned char)(0xF0 | (cp >> 18));
        buf[n++] = (unsigned char)(0x80 | ((cp >> 12) & 0x3F));
        buf[n++] = (unsigned char)(0x80 | ((cp >> 6) & 0x3F));
        buf[n++] = (unsigned char)(0x80 | (cp & 0x3F));
    }
    fwrite(buf, 1, (size_t)n, stdout);
}

void aila_print_bytes(const uint8_t* p, int64_t n) {
    if (p && n > 0) fwrite(p, 1, (size_t)n, stdout);
}

void aila_print_cstr(const uint8_t* p) {
    if (p) fputs((const char*)p, stdout);
}
