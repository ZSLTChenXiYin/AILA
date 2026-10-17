// Command aila 是 AILA 的管理器 + 编译器前端——**一份 Go 单文件**（rfcs/0003 的 G1–G4）。
//
// 无 cgo：`CGO_ENABLED=0 go build ./cmd/aila` 即可得到可分发产物；
// 语言前端（词法/语法/语义/IR）、C 绑定生成与构建编排全部在本进程内完成，
// 只有 IR → 目标文件与最终链接调用随包/系统的 clang、llc、cc。
package main

import (
	"os"

	"github.com/aila-lang/aila/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
