package types

// Layout 按 **C ABI** 求 size / align / 字段偏移（docs/specs/0002 §5.2）。
//
// 判据：tests/corpus/layout.json 由 Python 参考实现生成，Go 侧逐项比对。
type Layout struct {
	PtrSize int
}

// NewLayout 构造布局器（PtrSize=8 为 64 位目标，4 为 32 位目标）。
func NewLayout(ptrSize int) *Layout {
	if ptrSize != 4 {
		ptrSize = 8
	}
	return &Layout{PtrSize: ptrSize}
}

func pad(off, align int) int {
	if align <= 1 {
		return off
	}
	return (off + align - 1) / align * align
}

// structLayout 按声明序摆放一组 (size, align) 部件。
func structLayout(parts [][2]int) (int, int) {
	off, align := 0, 1
	for _, p := range parts {
		off = pad(off, p[1])
		off += p[0]
		if p[1] > align {
			align = p[1]
		}
	}
	return pad(off, align), align
}

// SizeAlign 返回 (size, align)。
func (l *Layout) SizeAlign(t Type) (int, int) {
	ptr := l.PtrSize
	switch x := t.(type) {
	case Prim:
		if x.K == "text" {
			return 2 * ptr, ptr
		}
		if x.K == "usize" || x.K == "isize" {
			return ptr, ptr
		}
		if n, ok := primSize(x.K); ok {
			return n, n
		}
		return ptr, ptr
	case View:
		return 2 * ptr, ptr
	case UnsafeView:
		return 2 * ptr, ptr
	case Fixed:
		es, ea := l.SizeAlign(x.Elem)
		return es * int(x.N), ea
	case Optional:
		if isNicheOptional(x) {
			return l.SizeAlign(x.Inner)
		}
		s, a := l.SizeAlign(x.Inner)
		if a < 1 {
			a = 1
		}
		return s + 1, a
	case Result:
		os_, oa := l.SizeAlign(x.Ok)
		es, ea := l.SizeAlign(x.Err)
		payloadSize, payloadAlign := structLayout([][2]int{{os_, oa}, {es, ea}})
		return structLayout([][2]int{{1, 1}, {payloadSize, payloadAlign}})
	case Chan:
		return ptr, ptr
	case Coro:
		return ptr, ptr
	case FnPtr:
		return ptr, ptr
	case Group:
		parts := make([][2]int, 0, len(x.Items))
		for _, it := range x.Items {
			s, a := l.SizeAlign(it)
			parts = append(parts, [2]int{s, a})
		}
		return structLayout(parts)
	case Named:
		switch info := x.Info.(type) {
		case *StructInfo:
			if info.Opaque {
				return 0, 1
			}
			parts := make([][2]int, 0, len(info.Fields))
			for _, f := range info.Fields {
				s, a := l.SizeAlign(f.Type)
				parts = append(parts, [2]int{s, a})
			}
			var total, align int
			if info.Packed {
				total = 0
				for _, p := range parts {
					total += p[0]
				}
				align = 1
			} else {
				total, align = structLayout(parts)
				for _, p := range parts {
					if p[1] > align {
						align = p[1]
					}
				}
			}
			if info.Align > 0 {
				if info.Align > align {
					align = info.Align
				}
				total = pad(total, align)
			}
			return total, align
		case *EnumInfo:
			tagS, tagA := l.SizeAlign(info.TagType)
			if !info.Payload {
				return tagS, tagA
			}
			parts := [][2]int{{tagS, tagA}}
			best := [2]int{-1, -1}
			for _, v := range info.Variants {
				if len(v.Payload) == 0 {
					continue
				}
				fp := make([][2]int, 0, len(v.Payload))
				for _, f := range v.Payload {
					s, a := l.SizeAlign(f.Type)
					fp = append(fp, [2]int{s, a})
				}
				size, align := structLayout(fp)
				if size > best[0] || (size == best[0] && align > best[1]) {
					best = [2]int{size, align}
				}
			}
			if best[0] >= 0 {
				parts = append(parts, best)
			}
			return structLayout(parts)
		}
	case Void:
		return 0, 1
	case Never:
		return 0, 1
	}
	return ptr, ptr
}

func primSize(kind string) (int, bool) {
	switch kind {
	case "i8", "u8", "bool":
		return 1, true
	case "i16", "u16":
		return 2, true
	case "i32", "u32", "c32", "f32":
		return 4, true
	case "i64", "u64", "f64":
		return 8, true
	}
	return 0, false
}

// isNicheOptional：句柄类用 ptr=null 表示 None（designs/0001 §3.7 规则 6）。
func isNicheOptional(t Optional) bool {
	if IsViewFamily(t.Inner) {
		return true
	}
	switch t.Inner.(type) {
	case FnPtr, Chan, Coro:
		return true
	}
	return false
}

// Size 返回 size。
func (l *Layout) Size(t Type) int {
	s, _ := l.SizeAlign(t)
	return s
}

// Align 返回 align。
func (l *Layout) Align(t Type) int {
	_, a := l.SizeAlign(t)
	return a
}

// FieldOffsets 返回 struct 各字段偏移。
func (l *Layout) FieldOffsets(info *StructInfo) []int {
	offsets := make([]int, 0, len(info.Fields))
	off := 0
	for _, f := range info.Fields {
		s, a := l.SizeAlign(f.Type)
		if info.Packed {
			offsets = append(offsets, off)
			off += s
			continue
		}
		off = pad(off, a)
		offsets = append(offsets, off)
		off += s
	}
	return offsets
}

// VariantPayloadOffsets 返回某个 enum 变体各字段相对值起点的偏移。
func (l *Layout) VariantPayloadOffsets(info *EnumInfo, v VariantInfo) []int {
	tagS, _ := l.SizeAlign(info.TagType)
	offsets := make([]int, 0, len(v.Payload))
	off := pad(tagS, 1)
	for _, f := range v.Payload {
		s, a := l.SizeAlign(f.Type)
		off = pad(off, a)
		offsets = append(offsets, off)
		off += s
	}
	return offsets
}
