package sema

import (
	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/types"
)

// 本文件处理「壳内前向声明 + 块内实现」的配对检查（designs/0001 §2.1 壳/块分工）。
//
// 为什么需要它：按可见性阶梯，只有壳里的 `public` 才能「包外可见」，
// 而壳里不允许带函数体——所以对外导出的 AILA 函数写成
//
//	public func f(a: I32) -> I32;      // 壳：接口 + 可见性（包外）
//	module {
//	    func f(a: I32) -> I32 { ... }   // 块：实现
//	}
//
// 配对规则：签名只登记一次（取声明），函数体取实现；声明必须有实现。

// sigForDecl 按名字取签名（方法带前缀）。
func (c *Checker) sigForDecl(fn *ast.FuncDecl) *FuncSig {
	if fn.OwnerType != "" {
		return c.sigs.Methods[fn.OwnerType+"."+fn.Name]
	}
	return c.sigs.Funcs[fn.Name]
}

// checkForwardDeclarations 检查壳内前向声明与块内实现的配对。
func (c *Checker) checkForwardDeclarations() {
	for _, d := range c.mod.AllDecls() {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Body != nil || fn.IsExtern || !fn.InShell {
			continue // 只关心壳里的非 #[c] 前向声明
		}
		sig := c.sigForDecl(fn)
		if sig == nil || sig.Decl != fn {
			continue
		}
		if sig.Impl == nil {
			c.err("E7031", fn.Span,
				"`"+qualifiedName(fn)+"` 在壳里声明了但没有实现",
				"壳只放接口；带函数体的实现必须落在 `module { ... }` 块内",
				"在 `module` 块内补 `func "+fn.Name+"(...) { ... }`，或删掉这条声明")
			continue
		}
		c.checkImplMatchesDecl(sig, sig.Impl)
	}
}

// checkImplMatchesDecl 校验块内实现与壳内声明的签名一致。
func (c *Checker) checkImplMatchesDecl(decl *FuncSig, impl *ast.FuncDecl) {
	if impl.Variadic != decl.Variadic {
		c.err("E3201", impl.Span, "`"+qualifiedName(impl)+"` 的变参修饰与壳内声明不一致", "签名必须一致", "")
		return
	}
	if len(impl.Params) != len(decl.Params) {
		c.err("E3201", impl.Span,
			"`"+qualifiedName(impl)+"` 的参数个数与壳内声明不一致",
			"声明 "+itoa(len(decl.Params))+" 个，实现 "+itoa(len(impl.Params))+" 个",
			"两边写成同一个签名")
		return
	}
	for i, p := range impl.Params {
		if i >= len(decl.Params) {
			break
		}
		if p.Type == nil && !p.IsSelf {
			continue
		}
		it := c.table.Resolve(p.Type, c.bag)
		if !sameType(decl.Params[i].Type, it) {
			c.err("E3201", p.Span,
				"`"+qualifiedName(impl)+"` 的参数 `"+decl.Params[i].Name+"` 类型与壳内声明不一致",
				"声明 "+types.TypeName(decl.Params[i].Type)+"，实现 "+types.TypeName(it),
				"两边写成同一个签名")
		}
	}
	if impl.Ret == nil && decl.RetGroup == nil {
		if !isVoid(decl.Ret) {
			c.err("E3201", impl.Span,
				"`"+qualifiedName(impl)+"` 缺少返回类型（壳内声明是 `"+types.TypeName(decl.Ret)+"`）",
				"签名必须一致", "")
		}
		return
	}
	retT := c.table.Resolve(impl.Ret, c.bag)
	if decl.Err != nil {
		if r, ok := retT.(types.Result); ok {
			if !sameType(decl.Err, r.Err) {
				c.err("E3201", impl.Span, "`"+qualifiedName(impl)+"` 的错误类型与壳内声明不一致",
					"声明 "+types.TypeName(decl.Err)+"，实现 "+types.TypeName(r.Err), "")
			}
			retT = r.Ok
		}
	}
	if decl.RetGroup == nil && !sameType(decl.Ret, retT) {
		c.err("E3201", impl.Span, "`"+qualifiedName(impl)+"` 的返回类型与壳内声明不一致",
			"声明 "+types.TypeName(decl.Ret)+"，实现 "+types.TypeName(retT),
			"两边写成同一个签名")
	}
}

func qualifiedName(fn *ast.FuncDecl) string {
	if fn.OwnerType != "" {
		return fn.OwnerType + "." + fn.Name
	}
	return fn.Name
}
