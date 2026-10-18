// Package skill 定位并读取「给 AI 编程工具的技能包」。
//
// 技能是一棵纯文本文档树（`SKILL.md` + `reference/*.md`），随发布包分发在
// `<AILA_HOME>/share/aila/skills/aila/`；仓库里也是同一个相对路径，
// 所以「开发时」和「解压即用后」用的是同一套定位逻辑（与 home 的相对目录铁律一致）。
package skill

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aila/internal/home"
)

// Name 是技能名（目录名，也是 SKILL.md frontmatter 里的 name）。
const Name = "aila"

// Dir 返回技能目录。
func Dir() string {
	return filepath.Join(home.Root(), "share", "aila", "skills", Name)
}

// Entry 是技能入口文件（AI 工具读这一个就够了）。
const Entry = "SKILL.md"

// File 是一个技能文件。
type File struct {
	Rel   string `json:"rel"`   // 相对技能目录的路径（始终用 / 分隔）
	Bytes int    `json:"bytes"` // 字节数
}

// Files 列出技能里的全部文件（按路径升序；SKILL.md 在最前）。
func Files() ([]File, error) {
	root := Dir()
	var out []File
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		st, statErr := d.Info()
		size := 0
		if statErr == nil {
			size = int(st.Size())
		}
		out = append(out, File{Rel: rel, Bytes: size})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Rel == Entry {
			return true
		}
		if out[j].Rel == Entry {
			return false
		}
		return out[i].Rel < out[j].Rel
	})
	return out, nil
}

// Read 读技能里的一个文件（rel 用 / 或 \ 都行）。
func Read(rel string) (string, error) {
	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	data, err := os.ReadFile(filepath.Join(Dir(), filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Install 把技能整棵树复制到 dst（dst 是「技能仓库目录」，会在其下建 aila/）。
// 返回实际写入的技能目录。
func Install(dst string) (string, error) {
	target := filepath.Join(dst, Name)
	files, err := Files()
	if err != nil {
		return "", err
	}
	for _, f := range files {
		text, err := Read(f.Rel)
		if err != nil {
			return "", err
		}
		out := filepath.Join(target, filepath.FromSlash(f.Rel))
		if err := os.MkdirAll(filepath.Dir(out), 0o777); err != nil {
			return "", err
		}
		if err := os.WriteFile(out, []byte(text), 0o666); err != nil {
			return "", err
		}
	}
	return target, nil
}
