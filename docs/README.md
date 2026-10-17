# AILA 文档索引

## 编号规则（强制）

- **每个目录独立编号，一律从 `0001` 起**，形如 `NNNN-kebab-case-title.md`；
- 编号一经分配不复用、不重排；新增文档 = 该目录当前最大号 + 1；
- 跨文档引用一律写「`目录/NNNN`」（如 `designs/0001`），不写文件名，避免改名断链；
- 设计稿 → RFC → SPEC → 计划是四类文档，各自独立编号，**同一主题在四类中的编号不必相同**。

## 目录与流程

开发流程固定四级，缺一级不得进入下一级：

| 阶段 | 目录 | 回答的问题 | 产物 |
|---|---|---|---|
| **设计（Design）** | `designs/` | 语言/系统「是什么、为什么」——语法、语义、动机与弃选 | 权威设计规范 |
| **RFC（Request for Comments）** | `rfcs/` | 「这一批要做成什么」——范围、方案对比、弃选理由、风险 | 待评审提案 |
| **SPEC（Specification）** | `specs/` | 「精确到可验收」——接口、数据格式、错误码、边界行为 | 可实现、可测试的规格 |
| **开发计划（Plan）** | `plans/` | 「按什么顺序做、怎么算做完」——里程碑、验收命令、批次划分 | 里程碑表 + 验收清单 |

```text
designs/0001（语言设计，唯一权威）
        │
        ▼
rfcs/0001（管理器 + 前端这一批做什么） ──► rfcs/0002（设计稿缺陷裁决）
        │                                   └► rfcs/0003（Go 重写决定与分批）
        ├──► specs/0001（管理器：CLI / 清单 / 工程模型 / bindgen）
        ├──► specs/0002（前端：词法 / 语法 / AST / 语义 / IR）
        └──► specs/0003（分发形态 / 随包工具链 / CI 打包）
                    │
                    ▼
             plans/0001（M1–M4 里程碑与验收）
```

## 现有文档

| 编号 | 标题 | 状态 |
|---|---|---|
| [designs/0001](designs/0001-language-spec.md) | AILA 语言设计规范（v0 语法设计定稿） | 权威、自包含 |
| [rfcs/0001](rfcs/0001-manager-and-frontend.md) | 管理器与编译器前端：范围、方案与弃选 | 历史记录（被 rfcs/0003 取代） |
| [rfcs/0002](rfcs/0002-design-defects.md) | 实现期发现的 `designs/0001` 内部不一致（缺陷登记 + 裁决） | 已裁决，待设计稿修订 |
| [rfcs/0003](rfcs/0003-go-rewrite.md) | 用 Go 重写管理器与前端的决定与分批计划（G1–G4） | 已完成（G1–G4 全绿，Python 前端已删） |
| [rfcs/0004](rfcs/0004-package-dependencies.md) | 跨项目依赖（包管理）：路径 / git / 锁文件 / vendor / 跨包可见性 | 已落地 |
| [rfcs/0005](rfcs/0005-pure-aila-stdlib.md) | 标准库必须用 AILA 自己实现（`#[intrinsic]` 与运行时边界） | 已落地 |
| [specs/0001](specs/0001-manager-cli.md) | 管理器：工程模型、清单、CLI、bindgen、工具链供给 | 已定稿 |
| [specs/0002](specs/0002-compiler-frontend.md) | 编译器前端：流水线、语法、AST、语义、LLVM IR | 已定稿 |
| [specs/0003](specs/0003-distribution-and-toolchain.md) | 分发形态、随包工具链解析、CI 打包流水线 | 已定稿（M2-D 已落地） |
| [plans/0001](plans/0001-manager-frontend-dev-plan.md) | 管理器 + 前端开发计划（M1–M4） | M1 完成、Go 重写 G1–G4 完成，M2+ 执行中 |

## 冲突裁决

`designs/0001` 是语言语义的唯一权威。任何 RFC / SPEC / 计划与它冲突，一律以 `designs/0001` 为准，并必须回改冲突文档；实现与文档冲突时，报缺陷而不是就地发明语义。
