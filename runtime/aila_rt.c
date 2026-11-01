// AILA 语言运行时（M1 薄层）
//
// 依据 designs/0001 §12.3「薄层四件套」：本文件提供其中两件的宿主实现 +
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
#include <time.h>      /* time / clock_gettime / nanosleep */
#include <process.h>   /* _getpid（熵源混合） */

#if defined(_WIN32)
/* 随包 MinGW 的 windows.h 头不全（psdk_inc/intrin-impl.h 缺失），
   这里直接声明用到的 WinAPI，runtime 保持自包含不依赖系统头。 */
typedef union _AILA_LARGE_INTEGER {
    long long QuadPart;
} AILA_LARGE_INTEGER;
__attribute__((dllimport)) int __stdcall QueryPerformanceCounter(AILA_LARGE_INTEGER*);
__attribute__((dllimport)) int __stdcall QueryPerformanceFrequency(AILA_LARGE_INTEGER*);
__attribute__((dllimport)) void __stdcall Sleep(unsigned long ms);
typedef void* AILA_HANDLE;
__attribute__((dllimport)) int __stdcall CloseHandle(AILA_HANDLE);
#else
#include <pthread.h>   /* pthread_create / pthread_detach */
#endif

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
// --------------------------------------------------------------------------- //
// 标准输出原语：**唯一**的 OS 边界
//
// 纪律（docs/rfcs/0005）：标准库必须用 AILA 自己实现，stdlib 里不出现任何 C；
// 这里只留两个语言原语（`#[intrinsic]`）：数字格式化 / UTF-8 编码 / 浮点打印全在 stdlib/io.aila。
// --------------------------------------------------------------------------- //
void aila_rt_write(const uint8_t* p, size_t n) {
    if (p && n > 0) fwrite(p, 1, (size_t)n, stdout);
}

void aila_rt_flush(void) { fflush(stdout); }

// --------------------------------------------------------------------------- //
// std.os 边界原语（designs/0003 §4.1）。只在宿主链接得到满足；嵌入式工程
// 不 import std.os，缺符号会在链接期报「std.os 需要 OS target」的映射提示。
// --------------------------------------------------------------------------- //

// ---- 命令行参数：C 入口（codegen 生成的 @main）转存，AILA 侧只读 ----

static int      g_argc = 0;
static char**   g_argv = NULL;

void aila_rt_args_save(int argc, char** argv) {
    g_argc = argc;
    g_argv = argv;
}

int64_t aila_rt_args_count(void) {
    return (int64_t)g_argc;
}

// 返回该参数的完整长度，并把 min(完整长度, cap) 字节拷进 out（-1 = 越界）。
int64_t aila_rt_args_get(size_t i, uint8_t* out, size_t cap) {
    if (i >= (size_t)g_argc || !g_argv) return -1;
    const char* s = g_argv[i];
    size_t n = 0;
    while (s[n]) n++;
    size_t take = n < cap ? n : cap;
    for (size_t k = 0; k < take; k++) out[k] = (uint8_t)s[k];
    return (int64_t)n;
}

// ---- 环境：name/val 传视图（ptr, len），C 侧补 NUL ----

// 返回值的完整长度并截断拷进 out；-1 = 未设置。
int64_t aila_rt_env_get(const uint8_t* name, size_t nlen, uint8_t* out, size_t ocap) {
    char buf[512];
    char* nb = buf;
    char* heap = NULL;
    if ((size_t)nlen + 1 > sizeof(buf)) {
        heap = (char*)malloc((size_t)nlen + 1);
        if (!heap) return -1;
        nb = heap;
    }
    memcpy(nb, name, (size_t)nlen);
    nb[nlen] = '\0';
    const char* v = getenv(nb);
    free(heap);
    if (!v) return -1;
    int64_t n = 0;
    while (v[n]) n++;
    int64_t take = n < ocap ? n : ocap;
    for (int64_t k = 0; k < take; k++) out[k] = (uint8_t)v[k];
    return n;
}

// 0 = 成功，-1 = 失败。
int64_t aila_rt_env_set(const uint8_t* name, size_t nlen, const uint8_t* val, size_t vlen) {
    if (nlen < 0 || vlen < 0) return -1;
    char* nb = (char*)malloc((size_t)nlen + 1);
    char* vb = (char*)malloc((size_t)vlen + 1);
    if (!nb || !vb) { free(nb); free(vb); return -1; }
    memcpy(nb, name, (size_t)nlen); nb[nlen] = '\0';
    memcpy(vb, val, (size_t)vlen); vb[vlen] = '\0';
#if defined(_WIN32)
    int rc = _putenv_s(nb, vb);
#else
    int rc = setenv(nb, vb, 1);
#endif
    free(nb);
    free(vb);
    return rc == 0 ? 0 : -1;
}

// ---- 时间与睡眠（单调钟） ----

uint64_t aila_rt_time_ms(void) {
#if defined(_WIN32)
    AILA_LARGE_INTEGER f, c;
    if (!QueryPerformanceFrequency(&f) || !QueryPerformanceCounter(&c)) return 0;
    return (uint64_t)(c.QuadPart / (f.QuadPart / 1000));
#else
    struct timespec ts;
    if (clock_gettime(CLOCK_MONOTONIC, &ts) != 0) return 0;
    return (uint64_t)ts.tv_sec * 1000u + (uint64_t)ts.tv_nsec / 1000000u;
#endif
}

void aila_rt_sleep_ms(uint64_t ms) {
#if defined(_WIN32)
    while (ms > 0xFFFFFFFFu) {   /* Sleep 收 DWORD */
        Sleep(0xFFFFFFFFu);
        ms -= 0xFFFFFFFFu;
    }
    Sleep((unsigned long)ms);
#else
    struct timespec ts;
    ts.tv_sec = (time_t)(ms / 1000u);
    ts.tv_nsec = (long)(ms % 1000u) * 1000000L;
    while (nanosleep(&ts, &ts) != 0) {
    } /* 被信号打断就接着睡剩余量 */
#endif
}

AILA_NORETURN void aila_rt_exit(uint32_t code) {
    fflush(stdout);
    exit((int)code);
}

// ---- 文件：句柄 = FILE* 的整数值（0 = 无效） ----

uint64_t aila_rt_file_open(const uint8_t* path, size_t n, uint32_t mode) {
    if (n < 0) return 0;
    char* pb = (char*)malloc((size_t)n + 1);
    if (!pb) return 0;
    memcpy(pb, path, (size_t)n);
    pb[n] = '\0';
    const char* m = (mode == 1u) ? "wb" : (mode == 2u) ? "ab" : "rb";
    FILE* f = fopen(pb, m);
    free(pb);
    return f ? (uint64_t)(uintptr_t)f : 0;
}

// 返回读到的字节数；0 = EOF；-1 = 读错误。
int64_t aila_rt_file_read(uint64_t h, uint8_t* out, size_t cap) {
    FILE* f = (FILE*)(uintptr_t)h;
    if (!f || cap < 0) return -1;
    if (cap == 0) return 0;
    size_t n = fread(out, 1, (size_t)cap, f);
    if (n == 0 && ferror(f)) return -1;
    return (int64_t)n;
}

// 返回写入的字节数（短写 = 调用侧判错）。
int64_t aila_rt_file_write(uint64_t h, const uint8_t* p, size_t n) {
    FILE* f = (FILE*)(uintptr_t)h;
    if (!f || n < 0) return -1;
    if (n == 0) return 0;
    return (int64_t)fwrite(p, 1, (size_t)n, f);
}

// ---- chan：SPSC 环形缓冲（designs/0001 §10.6） ----
//
// 控制块 + 缓冲在创建时一次性区域分配；无 close、无 select（v0）。
// head/tail 是 free-running 计数（模 2^64 回绕），count = head - tail：
// 生产者只写 head、消费者只写 tail，SPSC 下天然无竞争；原子序全 seq_cst
// （v0 零参数，designs/0001 §10.10 同款精神）。
typedef struct {
    uint64_t cap;
    uint64_t esize;
    uint64_t head;
    uint64_t tail;
} AilaChan;

static uint64_t aila_chan_buf_off(void) {
    return (uint64_t)((sizeof(AilaChan) + 7u) & ~(size_t)7u);
}

void* aila_rt_chan_create(void* region, int64_t cap, int64_t esize) {
    if (cap <= 0) aila_rt_trap("chan: capacity must be positive", 33, __FILE__, __LINE__);
    if (esize <= 0) aila_rt_trap("chan: element size must be positive", 34, __FILE__, __LINE__);
    uint64_t off = aila_chan_buf_off();
    uint8_t* base = (uint8_t*)aila_rt_region_alloc(
        region, (int64_t)(off + (uint64_t)cap * (uint64_t)esize), 8);
    AilaChan* c = (AilaChan*)base;
    c->cap = (uint64_t)cap;
    c->esize = (uint64_t)esize;
    c->head = 0;
    c->tail = 0;
    return base;
}

// 满 = 返回 0（调用侧挂起重试）；成功 = 1（值已入缓冲）。
int aila_rt_chan_try_send(void* cv, const void* src) {
    AilaChan* c = (AilaChan*)cv;
    uint64_t head = __atomic_load_n(&c->head, __ATOMIC_RELAXED);
    uint64_t tail = __atomic_load_n(&c->tail, __ATOMIC_ACQUIRE);
    if (head - tail >= c->cap) return 0;
    uint8_t* buf = (uint8_t*)cv + aila_chan_buf_off();
    memcpy(buf + (head % c->cap) * c->esize, src, (size_t)c->esize);
    __atomic_store_n(&c->head, head + 1, __ATOMIC_RELEASE);
    return 1;
}

// 空 = 返回 0（调用侧挂起重试）；成功 = 1（值已拷出）。
int aila_rt_chan_try_recv(void* cv, void* dst) {
    AilaChan* c = (AilaChan*)cv;
    uint64_t tail = __atomic_load_n(&c->tail, __ATOMIC_RELAXED);
    uint64_t head = __atomic_load_n(&c->head, __ATOMIC_ACQUIRE);
    if (tail == head) return 0;
    uint8_t* buf = (uint8_t*)cv + aila_chan_buf_off();
    memcpy(dst, buf + (tail % c->cap) * c->esize, (size_t)c->esize);
    __atomic_store_n(&c->tail, tail + 1, __ATOMIC_RELEASE);
    return 1;
}

// ---- thread：宿主线程创建（designs/0001 §10.8） ----
//
// fire-and-forget 无句柄：栈大小原样交给 OS 创建 API，生命周期归 OS。
// 参数块由 AILA 侧 thunk 用毕释放；kick 结构由 spawn 分配、trampoline 释放。
void* aila_rt_thread_args(int64_t n) {
    void* p = malloc((size_t)(n > 0 ? n : 1));
    if (!p) aila_rt_trap("thread: out of memory for args", 31, __FILE__, __LINE__);
    return p;
}

void aila_rt_thread_args_free(void* p) {
    free(p);
}

typedef struct {
    void* (*fn)(void*);
    void* arg;
} AilaThreadKick;

#if defined(_WIN32)
static unsigned __stdcall aila_thread_trampoline(void* p) {
    AilaThreadKick k = *(AilaThreadKick*)p;
    free(p);
    (void)k.fn(k.arg);
    return 0;
}

int aila_rt_thread_spawn(int64_t stack, void* (*fn)(void*), void* arg) {
    AilaThreadKick* k = (AilaThreadKick*)malloc(sizeof(AilaThreadKick));
    if (!k) return -1;
    k->fn = fn;
    k->arg = arg;
    uintptr_t h = _beginthreadex(NULL, (unsigned)(stack > 0 ? stack : 0),
                                 aila_thread_trampoline, k, 0, NULL);
    if (!h) {
        free(k);
        return -1;
    }
    CloseHandle((AILA_HANDLE)h);
    return 0;
}
#else
static void* aila_thread_trampoline(void* p) {
    AilaThreadKick k = *(AilaThreadKick*)p;
    free(p);
    (void)k.fn(k.arg);
    return NULL;
}

int aila_rt_thread_spawn(int64_t stack, void* (*fn)(void*), void* arg) {
    AilaThreadKick* k = (AilaThreadKick*)malloc(sizeof(AilaThreadKick));
    if (!k) return -1;
    k->fn = fn;
    k->arg = arg;
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    if (stack > 0) pthread_attr_setstacksize(&attr, (size_t)stack);
    pthread_t t;
    int rc = pthread_create(&t, &attr, aila_thread_trampoline, k);
    pthread_attr_destroy(&attr);
    if (rc != 0) {
        free(k);
        return -1;
    }
    pthread_detach(t);
    return 0;
}
#endif

// ---- std.atomics：原子原语（designs/0001 §10.10） ----
//
// 作用对象 = [Unsafe] 裸字节块 + 字节偏移，原子读写是与 C 侧共享内存
// （ISR / DMA / 其他线程所写）的唯一通道。全 seq_cst 零内存序参数（v0）。
// cas 返回操作前旧值，调用侧 old == expected 判成功（C11 弱 CAS 语义）。
// 偏移处自然对齐是硬件原子操作的物理事实：错对齐一律确定性 trap（§8.3
// 隐式陷阱安全网，检查在库内 C 层——[Unsafe] 不可解引用，AILA 层拿不到地址）。
#define AILA_DEFINE_ATOMIC_OPS(W, T)                                              \
    static T* aila_atomic_at##W(const void* base, int64_t off) {                  \
        const uint8_t* p = (const uint8_t*)base;                                  \
        if (off < 0)                                                              \
            aila_rt_trap("atomics: negative offset", 24, __FILE__, __LINE__);     \
        if ((uintptr_t)(p + off) % sizeof(T) != 0)                                \
            aila_rt_trap("atomics: misaligned atomic access", 33, __FILE__, __LINE__); \
        return (T*)(p + off);                                                     \
    }                                                                             \
    T aila_rt_atomic_load##W(const void* base, int64_t off) {                     \
        return __atomic_load_n(aila_atomic_at##W(base, off), __ATOMIC_SEQ_CST);   \
    }                                                                             \
    void aila_rt_atomic_store##W(void* base, int64_t off, T v) {                  \
        __atomic_store_n(aila_atomic_at##W(base, off), v, __ATOMIC_SEQ_CST);      \
    }                                                                             \
    T aila_rt_atomic_fetch_add##W(void* base, int64_t off, T v) {                 \
        return __atomic_fetch_add(aila_atomic_at##W(base, off), v, __ATOMIC_SEQ_CST); \
    }                                                                             \
    T aila_rt_atomic_fetch_sub##W(void* base, int64_t off, T v) {                 \
        return __atomic_fetch_sub(aila_atomic_at##W(base, off), v, __ATOMIC_SEQ_CST); \
    }                                                                             \
    T aila_rt_atomic_cas##W(void* base, int64_t off, T expected, T desired) {     \
        T old = expected;                                                         \
        __atomic_compare_exchange_n(aila_atomic_at##W(base, off), &old, desired,  \
                                    0, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST);       \
        return old;                                                               \
    }

AILA_DEFINE_ATOMIC_OPS(8, uint8_t)
AILA_DEFINE_ATOMIC_OPS(16, uint16_t)
AILA_DEFINE_ATOMIC_OPS(32, uint32_t)
AILA_DEFINE_ATOMIC_OPS(64, uint64_t)
#undef AILA_DEFINE_ATOMIC_OPS

void aila_rt_file_close(uint64_t h) {
    FILE* f = (FILE*)(uintptr_t)h;
    if (f) fclose(f);
}

void aila_rt_file_flush(uint64_t h) {
    FILE* f = (FILE*)(uintptr_t)h;
    if (f) fflush(f);
}

// ---- 随机种子：OS 熵源；拿不到就退回时间/地址混合 ----

uint64_t aila_rt_seed(void) {
#if defined(_WIN32)
    AILA_LARGE_INTEGER c;
    uint64_t v = 0;
    if (QueryPerformanceCounter(&c)) v = (uint64_t)c.QuadPart;
    v ^= (uint64_t)_getpid() * 0x9E3779B97F4A7C15ull;
    v ^= (uint64_t)time(NULL) * 0xBF58476D1CE4E5B9ull;
    return v;
#else
    FILE* f = fopen("/dev/urandom", "rb");
    if (f) {
        uint64_t v = 0;
        size_t n = fread(&v, 1, sizeof(v), f);
        fclose(f);
        if (n == sizeof(v)) return v;
    }
    uint64_t v = (uint64_t)time(NULL) * 0x9E3779B97F4A7C15ull;
    v ^= (uint64_t)(uintptr_t)&f;
    return v;
#endif
}

