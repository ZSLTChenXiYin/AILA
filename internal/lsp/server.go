package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// Server 是 stdio 版语言服务器。
type Server struct {
	in   *bufio.Reader
	out  io.Writer
	mu   sync.Mutex // 写回包时串行（诊断推送可能来自分析）
	ws   *Workspace
	log  io.Writer
	shut bool
}

// NewServer 构造服务器；log 为 nil 时不打日志（建议接 stderr）。
func NewServer(in io.Reader, out io.Writer, log io.Writer) *Server {
	return &Server{in: bufio.NewReader(in), out: out, ws: NewWorkspace(), log: log}
}

// Serve 跑主循环，直到 exit 或输入结束。
func (s *Server) Serve() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if msg == nil {
			continue
		}
		if err := s.handle(msg); err != nil {
			s.logf("处理 %s 失败：%v", msg.Method, err)
			if msg.isRequest() {
				s.replyError(msg.ID, -32603, err.Error())
			}
		}
		if s.shut && msg.Method == "exit" {
			return nil
		}
	}
}

// readMessage 读一条 Content-Length 分帧的 JSON-RPC 报文。
func (s *Server) readMessage() (*rpcMessage, error) {
	length := -1
	for {
		line, err := s.in.ReadString('\n')
		if err != nil {
			if err == io.EOF && line == "" {
				return nil, io.EOF
			}
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if i := strings.IndexByte(line, ':'); i > 0 {
			name := strings.ToLower(strings.TrimSpace(line[:i]))
			value := strings.TrimSpace(line[i+1:])
			if name == "content-length" {
				n, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("Content-Length 非法：%q", value)
				}
				length = n
			}
		}
	}
	if length < 0 {
		return nil, fmt.Errorf("缺少 Content-Length")
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(s.in, buf); err != nil {
		return nil, err
	}
	var msg rpcMessage
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, fmt.Errorf("JSON-RPC 解析失败：%w", err)
	}
	return &msg, nil
}

// write 发一条报文（自动加 Content-Length 头）。
func (s *Server) write(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	_, err = s.out.Write(data)
	return err
}

func (s *Server) reply(id json.RawMessage, result any) {
	if len(id) == 0 {
		return
	}
	m := map[string]any{"jsonrpc": "2.0", "id": id}
	if result == nil {
		m["result"] = nil
	} else {
		m["result"] = result
	}
	_ = s.write(m)
}

func (s *Server) replyError(id json.RawMessage, code int, message string) {
	if len(id) == 0 {
		return
	}
	_ = s.write(map[string]any{
		"jsonrpc": "2.0", "id": id,
		"error": map[string]any{"code": code, "message": message},
	})
}

func (s *Server) notify(method string, params any) {
	_ = s.write(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func (s *Server) logf(format string, args ...any) {
	if s.log == nil {
		return
	}
	fmt.Fprintf(s.log, "[aila-lsp] "+format+"\n", args...)
}

// handle 分发一条报文。
func (s *Server) handle(msg *rpcMessage) error {
	switch msg.Method {
	case "initialize":
		var p initializeParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		root := p.RootURI
		if root == "" {
			root = p.RootPath
		}
		if root == "" && len(p.WorkspaceFolders) > 0 {
			root = p.WorkspaceFolders[0].URI
		}
		s.ws.SetRoot(URI2Path(root))
		res := initializeResult{}
		res.Capabilities = serverCapabilities{
			TextDocumentSync: 1, // 全量同步
			DocumentSymbolProvider:    true,
			DefinitionProvider:        true,
			HoverProvider:             true,
			ReferencesProvider:        true,
			DocumentHighlightProvider: true,
			WorkspaceSymbolProvider:   true,
		}
		res.ServerInfo.Name = "aila-lsp"
		res.ServerInfo.Version = "0.1.0"
		s.reply(msg.ID, res)
		return nil

	case "initialized":
		return nil

	case "shutdown":
		s.shut = true
		s.reply(msg.ID, nil)
		return nil

	case "exit":
		s.ws.Close()
		return nil

	case "textDocument/didOpen":
		var p didOpenParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.ws.Open(p.TextDocument.URI, p.TextDocument.Text)
		s.publishDiagnostics(p.TextDocument.URI)
		return nil

	case "textDocument/didChange":
		var p didChangeParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		if n := len(p.ContentChanges); n > 0 {
			// 全量同步：只取最后一份文本
			s.ws.Open(p.TextDocument.URI, p.ContentChanges[n-1].Text)
		}
		s.publishDiagnostics(p.TextDocument.URI)
		return nil

	case "textDocument/didClose":
		var p textDocumentPositionParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.ws.CloseDoc(p.TextDocument.URI)
		s.notify("textDocument/publishDiagnostics", map[string]any{
			"uri": p.TextDocument.URI, "diagnostics": []Diagnostic{},
		})
		return nil

	case "textDocument/documentSymbol":
		var p textDocumentPositionParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.reply(msg.ID, s.ws.DocumentSymbols(p.TextDocument.URI))
		return nil

	case "textDocument/definition":
		var p textDocumentPositionParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.reply(msg.ID, s.ws.Definition(p.TextDocument.URI, p.Position))
		return nil

	case "textDocument/hover":
		var p textDocumentPositionParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.reply(msg.ID, s.ws.Hover(p.TextDocument.URI, p.Position))
		return nil

	case "textDocument/references":
		var p textDocumentPositionParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.reply(msg.ID, s.ws.References(p.TextDocument.URI, p.Position))
		return nil

	case "textDocument/documentHighlight":
		var p textDocumentPositionParams
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.reply(msg.ID, s.ws.Highlights(p.TextDocument.URI, p.Position))
		return nil

	case "workspace/symbol":
		var p struct {
			Query string `json:"query"`
		}
		if err := msg.decode(&p); err != nil {
			return err
		}
		s.reply(msg.ID, s.ws.WorkspaceSymbols(p.Query))
		return nil
	}
	if msg.isRequest() {
		s.replyError(msg.ID, -32601, "未实现的方法："+msg.Method)
	}
	return nil
}

// publishDiagnostics 分析当前文件所属工程并把诊断推给编辑器。
func (s *Server) publishDiagnostics(uri string) {
	diags := s.ws.Diagnostics(uri)
	s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri": uri, "diagnostics": diags,
	})
	for u, ds := range s.ws.OtherDiagnostics(uri) {
		s.notify("textDocument/publishDiagnostics", map[string]any{"uri": u, "diagnostics": ds})
	}
}
