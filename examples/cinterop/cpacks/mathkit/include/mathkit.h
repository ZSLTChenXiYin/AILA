/* 示例 6 的 C 包：导入即用（共置即关联，designs/0001 §11.8）
 *
 * 头文件给 bindgen 生成 `.ail` 绑定；同目录的 .c 由 cc 编译并自动链接。
 */
#ifndef MATHKIT_H
#define MATHKIT_H

#define MATHKIT_MAX 1000
#define MATHKIT_VERSION "0.1.0"

typedef int mathkit_err_t;

/* 纯加法：返回值错误码按 C 惯例直译，AILA 侧自己判码（§11.6） */
int mathkit_add(int a, int b);

int mathkit_clamp(int v, int lo, int hi);

unsigned int mathkit_hash(unsigned int seed, unsigned int v);

#endif
