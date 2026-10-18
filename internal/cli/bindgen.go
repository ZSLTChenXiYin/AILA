package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ailang/aila/internal/bindgen"
	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/manifest"
	"github.com/ailang/aila/internal/project"
	"github.com/ailang/aila/internal/resolve"
)

// newBindgenCmd 是 `aila bindgen`：把 cpack 的 C 头文件翻成 `.ail` 绑定模块。
//
// 两种文件不要混：
//   - `.aila` = AILA **源码**（人写）
//   - `.ail`  = bindgen 生成的**绑定声明**（机器生成，随构建重生成，不进版本库）
//
// 平时不需要手跑：`aila build/run/check` 遇到 `import ch _.cpacks.<名字>` 会自动生成并缓存。
// 这条命令是给「看产物 / 排错 / 预生成」用的（docs/specs/0001 §3.2 §6）。
func newBindgenCmd() *cobra.Command {
	var backend, board, out string
	cmd := &cobra.Command{
		Use:   "bindgen [目标]",
		Short: "生成 C 绑定（`.ail`）：把 cpack 的头文件翻成 AILA 声明模块",
		Long: "把 cpack 的 C 头文件翻成 `.ail` 绑定文件（机器生成，`.aila` 才是人写的源码）。\n" +
			"不给目标 = 生成项目 `cpacks/` 下的全部 cpack；给目标 = cpack 名（`math`）、相对路径（`cpacks/math`）或含头文件的目录。",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot(".")
			bag := &diag.Bag{}
			man := manifest.Load(root, bag)
			if man == nil {
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			if board != "" {
				man.Board = board
			}
			prj := project.Discover(root, bag)
			triple := man.EffectiveTriple(resolve.HostTriple())
			target := bindgen.Target{
				Board: man.Board, Triple: triple, PtrSize: man.PtrSize(),
			}

			type job struct {
				name string
				dir  string
			}
			var jobs []job
			switch {
			case len(args) == 0:
				names := make([]string, 0, len(prj.Cpacks))
				for name := range prj.Cpacks {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					jobs = append(jobs, job{name, prj.Cpacks[name]})
				}
				if len(jobs) == 0 {
					bag.Error("E3001", diag.Empty, "项目里没有 `cpacks/` 目录或其中没有 cpack",
						"没东西可绑定", "把 C 包放到 `cpacks/<名字>/`（头文件放 `include/`）")
				}
			default:
				spec := args[0]
				dir := spec
				if !filepath.IsAbs(dir) {
					dir = filepath.Join(root, dir)
				}
				// 允许写 `math`（cpack 名）或 `cpacks/math`
				if !dirExists(dir) {
					if d, ok := prj.Cpacks[spec]; ok {
						dir = d
					} else if d, ok := prj.Cpacks[strings.TrimPrefix(spec, "cpacks/")]; ok {
						dir = d
					}
				}
				if !dirExists(dir) {
					bag.Error("E3001", diag.Empty, "找不到 cpack `"+spec+"`", dir,
						"参数可以是 cpack 名（`math`）、相对路径（`cpacks/math`）或含头文件的目录")
				} else {
					abs, _ := filepath.Abs(dir)
					base := filepath.Base(abs)
					if strings.HasPrefix(abs, prj.Root) {
						rel, _ := filepath.Rel(prj.Root, abs)
						parts := strings.Split(filepath.ToSlash(rel), "/")
						if len(parts) >= 2 && parts[0] == "cpacks" {
							base = parts[1]
						}
					}
					jobs = append(jobs, job{base, abs})
				}
			}
			if len(jobs) > 1 && out != "" {
				bag.Error("E7001", diag.Empty, "`--out` 只能配单个 cpack 使用", "--out 是单文件",
					"要么指定一个 cpack，要么去掉 --out")
			}
			if bag.HasErrors() {
				printDiags(bag)
				return &exitError{ExitDiag}
			}

			outDir := filepath.Join(root, "build", "bindgen")
			if err := os.MkdirAll(outDir, 0o777); err != nil {
				bag.Error("E7001", diag.Empty, "建不了输出目录", err.Error(), "")
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			fail := false
			for _, j := range jobs {
				res := bindgen.Run(j.name, j.dir, target, bag, backend)
				for _, note := range res.Notes {
					if isBindgenWarning(note) {
						bag.Warning("W3001", diag.Empty, j.name+": "+note, "bindgen 边界", "")
					}
				}
				if bag.HasErrors() {
					fail = true
					continue
				}
				text := res.Header(j.name, target) + "\n\n" + res.Text
				path := filepath.Join(outDir, j.name+".ail")
				if out != "" {
					path = out
				}
				if err := os.WriteFile(path, []byte(text), 0o666); err != nil {
					bag.Error("E7001", diag.Empty, "写不了 `"+path+"`", err.Error(), "")
					fail = true
					continue
				}
				shown := path
				if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
					shown = filepath.ToSlash(rel)
				}
				fmt.Fprintf(stdout, "%s  →  %s（%s 档，%d 行）\n", j.name, shown, res.Backend,
					strings.Count(text, "\n"))
				if res.Text == "" {
					bag.Warning("W3001", diag.Empty, j.name+": 头文件里没有可映射的声明", "空绑定", "")
				}
			}
			if len(bag.Items) > 0 {
				printDiags(bag)
			}
			if fail || bag.HasErrors() {
				return &exitError{ExitDiag}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&backend, "backend", "auto", "bindgen 档位：auto | clang | builtin")
	cmd.Flags().StringVar(&board, "target", "", "目标 board（临时覆盖清单，影响 long/size_t 的映射）")
	cmd.Flags().StringVar(&out, "out", "", "输出文件（单 cpack 时可用）")
	return cmd
}

// isBindgenWarning 判断一条 bindgen 提示是否需要以 W3001 报给用户（未映射 / 无法求值 / 未解析）。
func isBindgenWarning(note string) bool {
	for _, kw := range []string{"未映射", "无法求值", "未解析条件编译", "无法求值成常量"} {
		if strings.Contains(note, kw) {
			return true
		}
	}
	return false
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}
