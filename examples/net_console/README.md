# net_console —— Windows 控制台 TCP 服务端（配 esp8266_oled）

等一个 ESP8266 连上来，你在控制台敲中文回车，它就把这一行（UTF-8）发过去，
板子把收到的字显示在 4 针 I2C OLED 上。

```
┌─────────────┐   Wi-Fi   ┌──────────┐   I2C   ┌────────────┐
│ 这个程序(PC) │ ────────▶ │ ESP8266  │ ──────▶ │ SSD1306 128x64 │
└─────────────┘  TCP 8266 └──────────┘  SCL/SDA └────────────┘
```

## 跑起来

```bash
aila run examples/net_console
# 监听 0.0.0.0:8266 …
# 等待 ESP8266 连接（同一 Wi-Fi 网段）…
# 已连上。输入中文后回车发送；空行退出。
# > 你好，世界
# [已发送 15 字节]
```

固件侧要填的是**本机局域网 IP**（`ipconfig` 看一眼），端口固定 8266。

## 协议

- 一条消息 = 一行 UTF-8 字节 + `\n`（对端只靠换行切行，不做长度帧）。
- 发送一行 `CLEAR` = 清屏（固件约定）。
- 空行（直接回车）= 结束会话并退出。

## 为什么有一个 C 文件

`cpacks/ws/ws.c` 是 Windows 边界层：Winsock（listen/accept/send/close）和控制台读行
（`ReadConsoleW` → UTF-8，非控制台时退回按字节读）都在这里。AILA 侧只管流程与提示。

- `import cc _.cpacks.ws;`：只编译 C、不生成绑定（声明在 `main.aila` 里手写，
  这样才能把 `[U8]` 视图按「只传指针」的约定接到 C 上）。
- `[build] libs = ["ws2_32"]`：链 Winsock 需要的系统库（manifest 的系统库名清单）。

## 自测

```bash
pwsh -File ci/smoke-net-console.ps1
```

它构建服务端、把 4 行中文喂进控制台输入、再起一个 TCP 客户端当假 ESP8266，
断言四行都按「一行一消息」收到了。
