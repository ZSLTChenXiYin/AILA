// Package lsp 是 AILA 的**语言服务器**（LSP over stdio）。
//
// 定位：编辑器插件只做渲染，语义全部来自编译器——不做「插件里再写一遍语法树」这种第二事实源。
// 覆盖的基础能力（M1）：
//
//	initialize / shutdown / exit     生命周期
//	textDocument/didOpen|didChange|didClose  文档同步（全量）+ 诊断推送
//	textDocument/documentSymbol      大纲（模块 / 类型 / 函数 / 方法 / 枚举变体 / 局部绑定）
//	textDocument/definition          跳转定义（含 import 别名、字段、方法、变体）
//	textDocument/hover               悬停：种类 + 名字 + 类型
//	textDocument/references          查找引用（工程内按名字 + 模块解析）
//	textDocument/documentHighlight   同文件高亮
//	workspace/symbol                 跨模块符号搜索
package lsp

import (
	"encoding/json"
	"fmt"
)

// --------------------------------------------------------------------------- //
// JSON-RPC / LSP 基础类型（只定义用到的字段）
// --------------------------------------------------------------------------- //

// Position 是零基行列。
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range 是文档区间（左闭右开）。
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location 是带 URI 的区间。
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic 是推给编辑器的诊断。
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=Error 2=Warning 3=Info 4=Hint
	Code     string `json:"code,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

// SymbolKind 用 LSP 的枚举值（只需要几个）。
const (
	KindFile     = 1
	KindModule   = 2
	KindNamespace = 3
	KindStruct   = 23
	KindEnum     = 10
	KindEnumMember = 22
	KindFunction = 12
	KindMethod   = 6
	KindConstant = 14
	KindVariable = 13
	KindField    = 8
	KindTypeParam = 26
)

// DocumentSymbol 是大纲节点（支持层级）。
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation 是 workspace/symbol 的结果项。
type SymbolInformation struct {
	Name     string   `json:"name"`
	Kind     int      `json:"kind"`
	Location Location `json:"location"`
}

// MarkupContent 是 hover 的内容载体。
type MarkupContent struct {
	Kind  string `json:"kind"` // plaintext | markdown
	Value string `json:"value"`
}

// Hover 是 hover 响应。
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// --------------------------------------------------------------------------- //
// JSON-RPC 报文
// --------------------------------------------------------------------------- //

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// isRequest 判断是否是需要应答的请求（有 id 且没有 method 的是响应）。
func (m *rpcMessage) isRequest() bool { return m.Method != "" && len(m.ID) > 0 }

// isNotification 判断是否是无应答的通知。
func (m *rpcMessage) isNotification() bool { return m.Method != "" && len(m.ID) == 0 }

func (m *rpcMessage) decode(v any) error {
	if len(m.Params) == 0 {
		return nil
	}
	if err := json.Unmarshal(m.Params, v); err != nil {
		return fmt.Errorf("参数解析失败：%w", err)
	}
	return nil
}

// --------------------------------------------------------------------------- //
// 常用参数/结果结构
// --------------------------------------------------------------------------- //

type textDocumentItem struct {
	URI     string `json:"uri"`
	Language string `json:"languageId,omitempty"`
	Version int    `json:"version,omitempty"`
	Text    string `json:"text"`
}

type textDocumentPositionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type initializeParams struct {
	RootURI string `json:"rootUri"`
	RootPath string `json:"rootPath"`
	WorkspaceFolders []struct {
		URI string `json:"uri"`
	} `json:"workspaceFolders"`
}

type serverCapabilities struct {
	TextDocumentSync         int  `json:"textDocumentSync"`
	DocumentSymbolProvider   bool `json:"documentSymbolProvider"`
	DefinitionProvider       bool `json:"definitionProvider"`
	HoverProvider            bool `json:"hoverProvider"`
	ReferencesProvider       bool `json:"referencesProvider"`
	DocumentHighlightProvider bool `json:"documentHighlightProvider"`
	WorkspaceSymbolProvider  bool `json:"workspaceSymbolProvider"`
}

type initializeResult struct {
	Capabilities serverCapabilities `json:"capabilities"`
	ServerInfo   struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}
