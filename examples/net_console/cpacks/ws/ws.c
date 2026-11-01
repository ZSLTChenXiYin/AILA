// cpacks/ws/ws.c —— 「控制台 TCP 服务端」的 Windows 薄边界层
//
// AILA 的纪律是「OS 边界一律走 C」（designs/0002 §1 纯度三线），socket 与控制台
// 这类平台 API 不进语言。本文件只做系统调用翻译，不含任何业务逻辑：
//
//   ws_console_utf8()          控制台输入输出切 UTF-8（AILA 的 io.println 写 UTF-8 字节）
//   ws_start(port)             建监听 socket；返回 >=0 句柄，负值 = 失败
//   ws_accept(lfd)             等一个入站连接（阻塞）；负值 = 失败
//   ws_send(fd, p, n)          把 n 字节全部发出；返回发出的字节数，负值 = 连接断了
//   ws_close(fd)               关 socket（重复关闭无害）
//   ws_line(dst, cap)          从控制台读一行，UTF-16 → UTF-8，去掉行尾 CR/LF；
//                              >0 = 字节数，0 = 空行，<0 = EOF / 读失败
//
// 句柄宽度：Windows 的 SOCKET 是 UINT_PTR（x64 上 64 位），所以用 long long 往返，
// 不能用 int（会被截断，而且截断后 closesocket 会误伤别的句柄）。
//
// 编译：随 aila 构建（`import cc _.cpacks.ws;` 只编 C、不生成绑定），
// 链接：manifest 的 `[build] libs = ["ws2_32"]` 提供 Winsock。

#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#include <stdio.h>
#include <string.h>
#include <fcntl.h>
#include <io.h>

static int g_wsa_ready = 0;

static int ws_ensure(void) {
    if (g_wsa_ready) {
        return 0;
    }
    WSADATA d;
    if (WSAStartup(MAKEWORD(2, 2), &d) != 0) {
        return -1;
    }
    g_wsa_ready = 1;
    return 0;
}

void ws_console_utf8(void) {
    /* 输入也切 UTF-8：ReadConsoleW 走宽字符，这里主要让「AILA 写的 UTF-8 字节」被正确显示 */
    SetConsoleOutputCP(CP_UTF8);
    SetConsoleCP(CP_UTF8);
    /* 管道 / 重定向输入走字节通道：不设二进制模式的话 0x1A 会被当 EOF、
       CRLF 会被翻译，中文（多字节 >=0x80）第一个字节就可能踩雷 */
    _setmode(_fileno(stdin), _O_BINARY);
}

long long ws_start(unsigned int port) {
    if (ws_ensure() != 0) {
        return -1;
    }
    SOCKET s = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (s == INVALID_SOCKET) {
        return -2;
    }
    BOOL yes = TRUE;
    setsockopt(s, SOL_SOCKET, SO_REUSEADDR, (const char *)&yes, sizeof yes);

    struct sockaddr_in a;
    memset(&a, 0, sizeof a);
    a.sin_family = AF_INET;
    a.sin_addr.s_addr = htonl(INADDR_ANY);
    a.sin_port = htons((unsigned short)port);
    if (bind(s, (struct sockaddr *)&a, (int)sizeof a) != 0) {
        closesocket(s);
        return -3;
    }
    if (listen(s, 1) != 0) {
        closesocket(s);
        return -4;
    }
    return (long long)s;
}

long long ws_accept(long long lfd) {
    SOCKET l = (SOCKET)lfd;
    SOCKET c = accept(l, NULL, NULL);
    if (c == INVALID_SOCKET) {
        return -1;
    }
    /* 发小包不要 Nagle 攒着，OLED 那种「打一行显示一行」的节奏不需要合并 */
    BOOL nodelay = TRUE;
    setsockopt(c, IPPROTO_TCP, TCP_NODELAY, (const char *)&nodelay, sizeof nodelay);
    return (long long)c;
}

long long ws_send(long long fd, const unsigned char *p, unsigned long n) {
    SOCKET s = (SOCKET)fd;
    unsigned long sent = 0;
    while (sent < n) {
        int k = send(s, (const char *)(p + sent), (int)(n - sent), 0);
        if (k <= 0) {
            return -1;
        }
        sent += (unsigned long)k;
    }
    return (long long)sent;
}

void ws_close(long long fd) {
    if (fd < 0) {
        return;
    }
    closesocket((SOCKET)fd);
}

int ws_line(unsigned char *dst, unsigned long cap) {
    HANDLE in = GetStdHandle(STD_INPUT_HANDLE);
    if (in != INVALID_HANDLE_VALUE && in != NULL) {
        DWORD mode = 0;
        /* 真控制台：用宽字符读，中文输入与代码页无关 */
        if (GetConsoleMode(in, &mode)) {
            wchar_t w[512];
            DWORD got = 0;
            if (!ReadConsoleW(in, w, (DWORD)(sizeof w / sizeof w[0] - 1), &got, NULL)) {
                return -1; /* EOF（Ctrl+Z）或读失败 */
            }
            if (got == 0) {
                return -1;
            }
            while (got > 0 && (w[got - 1] == L'\n' || w[got - 1] == L'\r')) {
                got--;
            }
            if (got == 0) {
                return 0; /* 空行 */
            }
            int n = WideCharToMultiByte(CP_UTF8, 0, w, (int)got, (char *)dst, (int)cap, NULL, NULL);
            return n > 0 ? n : -1;
        }
    }
    /* 管道 / 重定向：按 UTF-8 字节读一行（ws_console_utf8 已把 stdin 设成二进制） */
    char tmp[512];
    if (!fgets(tmp, (int)sizeof tmp, stdin)) {
        return -1;
    }
    size_t n = strlen(tmp);
    while (n > 0 && (tmp[n - 1] == '\n' || tmp[n - 1] == '\r')) {
        n--;
    }
    if (n > cap) {
        n = cap;
    }
    memcpy(dst, tmp, n);
    return (int)n;
}
