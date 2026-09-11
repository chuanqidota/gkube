# 代码质量工具链 — 设计文档与实施方案

> **版本**: v1.0 | **日期**: 2026-09-09 | **状态**: Draft

---

## 第一部分：设计文档

### 1. 背景与目标

#### 1.1 现状

项目当前**零代码质量工具**：

- 无 ESLint / Prettier / golangci-lint 配置
- 无 `.editorconfig`
- 无 pre-commit hooks
- 唯一的检查是 `vue-tsc -b`（TypeScript 编译器检查），已启用 `strict`、`noUnusedLocals`、`noUnusedParameters`、`noFallthroughCasesInSwitch`
- `.gitignore` 缺少 `node_modules/`、`dist/` 等前端条目

现有代码风格（人工审查）：

| 维度 | 前端 (TS/Vue) | 后端 (Go) |
|------|--------------|-----------|
| 缩进 | 2 空格 | Tab (gofmt) |
| 分号 | 无 | N/A |
| 引号 | 单引号 | N/A |
| 尾逗号 | 有 | N/A |
| 接口属性终止符 | 无分号，裸换行 | N/A |
| Vue SFC | `<script setup lang="ts">` | N/A |

#### 1.2 目标

| # | 目标 | 衡量标准 |
|---|------|---------|
| 1 | 自动化格式统一 | `npm run format:check` 和 `make lint` 零报错 |
| 2 | 静态缺陷检测 | pre-commit 拦截 errcheck/bodyclose/no-explicit-any 等问题 |
| 3 | 渐进式落地 | 存量问题通过 warn 过渡，不阻塞日常开发 |
| 4 | IDE 即时反馈 | VS Code 保存时自动 lint + format |
| 5 | 不引入 CI/CD | 本次只做本地工具链，CI 后续单独规划 |

---

### 2. 工具选型

#### 2.1 前端：ESLint 9 + Prettier

| 工具 | 版本 | 职责 |
|------|------|------|
| `eslint` | ^9.x | 代码质量规则（flat config） |
| `@eslint/js` | ^9.x | ESLint 推荐规则集 |
| `typescript-eslint` | ^8.x | TypeScript 类型感知规则（需确认与 TS 6.0 兼容性，见风险章节） |
| `eslint-plugin-vue` | ^10.x | Vue SFC 规则 |
| `prettier` | ^3.x | 代码格式化 |
| `eslint-config-prettier` | ^10.x | 关闭与 Prettier 冲突的 ESLint 规则 |

**选型理由**：

- ESLint 9 flat config 是当前标准，`.eslintrc` 已废弃
- `typescript-eslint` 的类型感知规则能捕获 `vue-tsc` 无法发现的运行时 bug
- Prettier 管格式，ESLint 管逻辑，职责分离
- 不选 Biome：Vue SFC 支持不成熟，插件生态不足

#### 2.2 后端：golangci-lint

| 工具 | 版本 | 职责 |
|------|------|------|
| `golangci-lint` | ^1.62.x | Go 静态分析聚合器（集成 govet/staticcheck/errcheck/gofmt/goimports 等） |

**选型理由**：

- Go 社区事实标准，单一二进制集成 100+ linter
- 支持 `--fix` 自动修复
- 不需要单独装 `gofmt`/`goimports`

#### 2.3 通用工具

| 工具 | 版本 | 职责 |
|------|------|------|
| `.editorconfig` | — | 编辑器基础格式对齐 |
| `husky` | ^9.x | Git hooks 管理 |
| `lint-staged` | ^15.x | 仅对暂存文件运行 lint |

---

### 3. 配置设计

#### 3.1 EditorConfig

**文件**: `.editorconfig`（项目根目录）

```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{ts,vue,js,json,yaml,yml,css,html,md}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false
```

#### 3.2 ESLint

**文件**: `frontend/eslint.config.js`

```js
import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import pluginVue from 'eslint-plugin-vue'
import prettier from 'eslint-config-prettier'

export default [
  { ignores: ['dist/**', 'node_modules/**', 'src/**/*.d.ts'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs['flat/recommended'],
  prettier,
  {
    files: ['**/*.{ts,vue}'],
    rules: {
      // ---- TypeScript ----
      // 注意：noUnusedLocals/noUnusedParameters 已在 tsconfig 中启用（vue-tsc 检查）
      // ESLint 版本更灵活（支持 _ 前缀忽略），建议关闭 tsconfig 中的对应检查，改用 ESLint
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': ['error', {
        argsIgnorePattern: '^_', varsIgnorePattern: '^_',
      }],
      '@typescript-eslint/consistent-type-imports': ['error', { prefer: 'type-imports' }],

      // ---- Vue ----
      'vue/multi-word-component-names': 'off',
      'vue/define-macros-order': ['error', { order: ['defineProps', 'defineEmits'] }],
      'vue/block-order': ['error', { order: ['script', 'template', 'style'] }],

      // ---- 通用质量 ----
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      eqeqeq: ['error', 'always'],
      'no-var': 'error',
      'prefer-const': 'error',
    },
  },
  {
    files: ['**/*.vue'],
    languageOptions: { parserOptions: { parser: tseslint.parser } },
  },
]
```

**关键决策**：

| 规则 | 设置 | 原因 |
|------|------|------|
| `no-explicit-any` | `warn` | 存量约 806 处（预估，以实际扫描为准），先告警不阻塞 |
| `multi-word-component-names` | `off` | Header/Sidebar 等单词组件名大量使用 |
| `consistent-type-imports` | `error` | 强制 `import type`，利于 tree-shaking |
| `no-unused-vars` | ESLint 版本 | tsconfig 的 `noUnusedLocals`/`noUnusedParameters` 不支持 `_` 前缀忽略，建议关闭 tsconfig 检查改用 ESLint 版本 |

#### 3.3 Prettier

**文件**: `frontend/.prettierrc`

```json
{
  "semi": false,
  "singleQuote": true,
  "trailingComma": "all",
  "printWidth": 100,
  "tabWidth": 2,
  "endOfLine": "lf",
  "vueIndentScriptAndStyle": false
}
```

**文件**: `frontend/.prettierignore`

```
dist
node_modules
*.d.ts
package.json
package-lock.json
```

#### 3.4 golangci-lint

**文件**: `backend/.golangci.yml`

```yaml
run:
  timeout: 5m
  go: '1.26'

linters:
  enable:
    - govet           # Go 官方 vet
    - staticcheck     # 全面静态分析
    - errcheck        # 未检查的错误返回值
    - gosimple        # 代码简化
    - ineffassign     # 无效赋值
    - unused          # 未使用的代码
    - gofmt           # 格式检查
    - goimports       # import 排序
    - misspell        # 拼写检查
    - gocritic        # 代码风格和最佳实践
    - bodyclose       # HTTP Response Body 未关闭
    - noctx           # HTTP 请求未带 context
    # - exportloopref # 已废弃：Go 1.22 已修复循环变量捕获问题，Go 1.26 不再需要
    - prealloc        # slice 预分配
    - unconvert       # 不必要的类型转换
    - errorlint       # error wrapping 最佳实践

  disable:
    - depguard        # 不限制依赖
    - funlen          # handler wrapper 闭包天然长
    - lll             # 不限制行宽（gofmt 控制）
    - godox           # 允许 TODO/FIXME
    - wsl             # 不强制空行风格

linters-settings:
  govet:
    enable-all: true
    disable:
      - fieldalignment  # 改字段顺序会破坏序列化兼容

  gocritic:
    enabled-tags: [diagnostic, style, performance]
    disabled-checks:
      - hugeParam       # K8s 对象天然大
      - rangeValCopy

  errcheck:
    check-type-assertions: true
    exclude-functions:
      - (io.Closer).Close

  misspell:
    locale: US
    ignore-words: [k8s, kubeconfig]

  errorlint:
    errorf: false     # 项目大量用 %v，不强制切 %w

issues:
  exclude-rules:
    - path: _test\.go$
      linters: [errcheck, gocritic]
    - path: cmd/
      linters: [errcheck]

  # 渐进式：首次只检查新增代码
  new-from-rev: HEAD
  max-issues-per-linter: 50
  max-same-issues: 5
```

**关键决策**：

| 设置 | 值 | 原因 |
|------|---|------|
| `new-from-rev: HEAD` | 渐进式 | 只查新增代码，不强制修存量 |
| `fieldalignment` | 关闭 | 改结构体字段顺序影响序列化兼容 |
| `errorlint.errorf` | false | 项目大量 `%v`，不强制切 `%w` |
| `funlen` | 关闭 | handler wrapper 闭包天然长函数 |
| `bodyclose` + `noctx` | 启用 | 捕获 HTTP 资源泄漏和缺 context |
| `exportloopref` | 已移除 | Go 1.22 已修复循环变量捕获，Go 1.26 不需要 |

---

### 4. npm Scripts 与 Makefile

#### 4.1 前端 `package.json` 新增脚本

```jsonc
{
  "scripts": {
    // 已有
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "preview": "vite preview",

    // 新增
    "lint": "eslint .",
    "lint:fix": "eslint . --fix",
    "format": "prettier --write \"src/**/*.{ts,vue,css,json}\"",
    "format:check": "prettier --check \"src/**/*.{ts,vue,css,json}\"",
    "check": "vue-tsc -b && eslint . && prettier --check \"src/**/*.{ts,vue,css,json}\""
    // vue-tsc 放最前：类型检查耗时最长但错误最严重，fail-fast
  }
}
```

| 命令 | 用途 |
|------|------|
| `npm run lint` | 检查所有文件 |
| `npm run lint:fix` | 自动修复 |
| `npm run format` | Prettier 格式化 |
| `npm run format:check` | 仅检查格式（不修改） |
| `npm run check` | 全量检查入口 |

#### 4.2 后端 `Makefile`

**文件**: `backend/Makefile`

```makefile
.PHONY: lint lint-fix build run test check

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

build:
	go build -o gkube .

run:
	go run main.go

test:
	go test ./...

check: lint build
```

---

### 5. Git Hooks 设计

#### 5.1 架构

```
git commit
    ↓
pre-commit hook (Husky)
    ↓
lint-staged（仅暂存文件）
    ├── frontend/**/*.{ts,vue}  → eslint --fix + prettier --write
    ├── frontend/**/*.{css,json,md} → prettier --write
    └── backend/**/*.go         → golangci-lint run --fix --new-from-rev=HEAD
    ↓
自动修复的文件重新 git add
    ↓
仍有 error → commit 失败
全部通过 → commit 成功
```

#### 5.2 配置

**根目录 `package.json`**：

```jsonc
{
  "private": true,
  "type": "module",
  "scripts": { "prepare": "husky" },
  "devDependencies": { "husky": "^9.x", "lint-staged": "^15.x" }
}
```

**`.husky/pre-commit`**：

```bash
#!/bin/sh
. "$(dirname "$0")/_/husky.sh"
npx lint-staged
```

> 注意：`lint-staged.config.js` 使用 ESM 语法（`export default`），需在根目录 `package.json` 中设置 `"type": "module"`。

**`lint-staged.config.js`**：

```js
export default {
  'frontend/**/*.{ts,vue}': ['eslint --fix', 'prettier --write'],
  'frontend/**/*.{css,json,md}': ['prettier --write'],
  'backend/**/*.go': ['golangci-lint run --fix --new-from-rev=HEAD'],
}
```

---

### 6. VS Code 集成

**文件**: `.vscode/settings.json`（提交到仓库）

```jsonc
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.lintOnSave": "package",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.codeActionsOnSave": { "source.organizeImports": "explicit" }
  },
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "[typescript]": { "editor.defaultFormatter": "esbenp.prettier-vscode" },
  "[vue]": { "editor.defaultFormatter": "esbenp.prettier-vscode" },
  "editor.codeActionsOnSave": { "source.fixAll.eslint": "explicit" },
  "editor.rulers": [100],
  "files.eol": "\n",
  "files.trimTrailingWhitespace": true
}
```

---

### 7. 预期效果

#### 7.1 可自动修复（`--fix`）

| 来源 | 问题 | 预估数量 |
|------|------|---------|
| Prettier | 格式不统一（缩进、空格、换行） | 全量格式化 |
| ESLint | `prefer-const` / `no-var` / `eqeqeq` | ~50 处 |
| ESLint | `consistent-type-imports` | ~100 处 |
| golangci-lint | `gofmt` / `goimports` / `gosimple` | ~20 处 |

#### 7.2 需人工处理

| 来源 | 问题 | 预估数量 | 级别 |
|------|------|---------|------|
| ESLint | `no-explicit-any` | ~806 处（预估） | warn |
| ESLint | `no-console` | ~30 处 | warn |
| golangci-lint | `errcheck` | ~100 处 | error |
| golangci-lint | `bodyclose` | ~10 处 | error |
| golangci-lint | `gocritic` | ~50 处 | warn |

#### 7.3 开发者体验

- IDE 保存时自动 lint + format
- commit 时自动修复格式，阻止逻辑错误
- PR 审查不再讨论风格，只关注逻辑

---

### 8. 风险与缓解

| 风险 | 缓解措施 |
|------|---------|
| 格式化污染 `git blame` | `.git-blame-ignore-revs` 记录格式化 commit |
| golangci-lint 过严阻塞开发 | `new-from-rev: HEAD` 渐进式 + 关闭主观规则 |
| `any` 警告疲劳 | 先 `warn` 不阻塞，按模块分配清理 |
| lint-staged 执行慢 | 仅检查暂存文件 + golangci-lint 增量模式 |
| `typescript-eslint` ^8.x 与 TS 6.0 兼容性 | 安装时确认版本兼容；如不支持需升级 `typescript-eslint` 或降级 TS |

**`.git-blame-ignore-revs`**：

```
# 格式化 commit（git blame 时忽略）
# <commit-hash>  chore: format all files with prettier
```

格式化 commit 后填入 hash，开发者执行：

```bash
git config blame.ignoreRevsFile .git-blame-ignore-revs
```

---

## 第二部分：实施方案

### 实施总览

```
Phase 1  基础设施搭建     1 天
Phase 2  Git Hooks 集成   0.5 天
Phase 3  存量问题标记     1-2 天
```

总计 **2.5-3.5 天**，不含 CI/CD。

---

### Phase 1：基础设施搭建（1 天）

#### 1.1 创建 EditorConfig

```bash
# 项目根目录创建 .editorconfig（内容见 3.1）
```

#### 1.2 前端：安装依赖 + 创建配置

```bash
cd frontend

# 安装 ESLint + Prettier
# 注意：项目使用 TypeScript ~6.0.2，安装时确认 typescript-eslint 版本兼容
# 如遇兼容问题，尝试 npm install -D typescript-eslint@canary
npm install -D eslint @eslint/js typescript-eslint eslint-plugin-vue \
  prettier eslint-config-prettier

# 创建配置文件
# - eslint.config.js（内容见 3.2）
# - .prettierrc（内容见 3.3）
# - .prettierignore（内容见 3.3）

# 添加 scripts 到 package.json（内容见 4.1）
```

**验证**：

```bash
npm run lint           # 应能运行，输出 warning/error 列表
npm run format:check   # 应能运行，报告格式不一致的文件
npm run format         # 格式化所有文件
```

**可选优化**：关闭 `tsconfig.app.json` 中的 `noUnusedLocals` 和 `noUnusedParameters`，改由 ESLint 的 `@typescript-eslint/no-unused-vars` 统一管理（支持 `_` 前缀忽略）。

#### 1.3 后端：安装 golangci-lint + 创建配置

```bash
# 安装 golangci-lint（推荐 latest，与 .golangci.yml 中 go 版本配合）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 或 macOS
brew install golangci-lint

# 创建配置文件
cd backend
# - .golangci.yml（内容见 3.4）
# - Makefile（内容见 4.2）
```

**验证**：

```bash
cd backend
make lint    # 应能运行，检查新增代码
make build   # 确认编译不受影响
```

#### 1.4 创建 VS Code 配置

```bash
# 创建 .vscode/settings.json（内容见 6）
```

#### 1.5 更新 .gitignore

当前 `.gitignore` **缺少**以下前端条目，需新增：

```gitignore
# 新增
node_modules/
dist/
.eslintcache
.husky/_
```

#### 1.6 阶段产出

- [ ] `.editorconfig` 可用
- [ ] `npm run lint` / `npm run format:check` 可运行
- [ ] `make lint` / `make build` 可运行
- [ ] VS Code 保存时自动 lint + format
- [ ] `.gitignore` 已更新

---

### Phase 2：Git Hooks 集成（0.5 天）

#### 2.1 根目录初始化 Husky

```bash
# 项目根目录
cd /data/gkube
npm init -y
npm install -D husky lint-staged
npx husky init
```

#### 2.2 配置 pre-commit hook

```bash
# 编辑 .husky/pre-commit（内容见 5.2）
```

#### 2.3 创建 lint-staged 配置

```bash
# 创建 lint-staged.config.js（内容见 5.2）
```

#### 2.4 验证

```bash
# 前端：修改一个 .ts 文件
echo "const x = 1" >> frontend/src/test-lint.ts
git add frontend/src/test-lint.ts
git commit -m "test: lint hook"
# 应触发 eslint + prettier

# 后端：修改一个 .go 文件
echo "// test" >> backend/main.go
git add backend/main.go
git commit -m "test: lint hook"
# 应触发 golangci-lint

# 清理测试文件
git checkout -- frontend/src/test-lint.ts backend/main.go
rm -f frontend/src/test-lint.ts
```

#### 2.5 阶段产出

- [ ] `git commit` 自动触发 lint-staged
- [ ] 前端文件触发 ESLint + Prettier
- [ ] 后端文件触发 golangci-lint
- [ ] 自动修复的文件被重新暂存

---

### Phase 3：存量问题标记（1-2 天）

#### 3.1 前端格式化（一次性）

```bash
cd frontend

# 1. 全量格式化
npm run format

# 2. 检查还有哪些 lint 问题
npm run lint 2>&1 | tee /tmp/eslint-baseline.txt

# 3. 统计
echo "=== ESLint baseline ==="
npm run lint 2>&1 | grep -c "warning" || echo "0 warnings"
npm run lint 2>&1 | grep -c "error" || echo "0 errors"
```

**Git 操作**：

```bash
# 单独一个 commit，只含格式变更
git add -A
git commit -m "style: format all files with prettier"

# 记录 commit hash，填入 .git-blame-ignore-revs
echo "# $(git log --oneline -1)" >> .git-blame-ignore-revs
git config blame.ignoreRevsFile .git-blame-ignore-revs
```

#### 3.2 后端存量扫描

```bash
cd backend

# 去掉 new-from-rev，检查全部代码
golangci-lint run ./... 2>&1 | tee /tmp/golangci-baseline.txt

# 统计（golangci-lint 结尾输出汇总行，如 "3 issues: 1 error, 2 warnings"）
echo "=== golangci-lint baseline ==="
golangci-lint run ./... 2>&1 | tail -1
```

#### 3.3 存量问题分类

将扫描结果分类，制定清理计划：

**前端（按优先级）**：

| 优先级 | 规则 | 预估数量 | 清理策略 |
|--------|------|---------|---------|
| P1 | `eqeqeq` / `no-var` / `prefer-const` | ~50 | `lint:fix` 自动修复 |
| P1 | `consistent-type-imports` | ~100 | `lint:fix` 自动修复 |
| P2 | `no-console` | ~30 | 手动删除或改为 `console.warn/error` |
| P3 | `no-explicit-any` | ~806（预估） | 按模块逐步替换为具体类型 |

**后端（按优先级）**：

| 优先级 | 规则 | 预估数量 | 清理策略 |
|--------|------|---------|---------|
| P1 | `errcheck`（未检查错误返回值） | ~100 | 手动添加 `_ =` 或 `if err != nil` |
| P1 | `bodyclose`（HTTP Body 未关闭） | ~10 | 添加 `defer resp.Body.Close()` |
| P2 | `gocritic` | ~50 | 按建议逐个修复 |
| P2 | `misspell` | ~10 | `lint-fix` 自动修复 |

#### 3.4 阶段产出

- [ ] 前端格式化 commit 完成（单独 commit）
- [ ] `.git-blame-ignore-revs` 已记录格式化 commit
- [ ] ESLint baseline 数量已记录
- [ ] golangci-lint baseline 数量已记录
- [ ] 存量问题分类表已建立
- [ ] `new-from-rev: HEAD` 已在 `.golangci.yml` 中（持续生效）

---

## 第三部分：文件清单

实施后项目新增/修改的文件：

```
gkube/
├── .editorconfig                    # 新增
├── .gitignore                       # 修改（新增 node_modules/dist/.eslintcache）
├── .git-blame-ignore-revs           # 新增（格式化 commit 后填写）
├── .vscode/
│   └── settings.json                # 新增
├── package.json                     # 新增（根目录，type:module + husky/lint-staged）
├── lint-staged.config.js            # 新增
├── .husky/
│   └── pre-commit                   # 新增
├── frontend/
│   ├── eslint.config.js             # 新增
│   ├── .prettierrc                  # 新增
│   ├── .prettierignore              # 新增
│   └── package.json                 # 修改（添加 scripts + devDeps）
└── backend/
    ├── .golangci.yml                # 新增
    └── Makefile                     # 新增
```

---

## 第四部分：后续规划（本次不做）

以下为后续可独立推进的事项，不在本次范围内：

| 事项 | 前置条件 | 说明 |
|------|---------|------|
| CI/CD 集成 | 有 CI 环境（GitHub Actions 等） | 在 PR/push 时自动跑 `npm run check` 和 `make lint` |
| 单元测试 | lint 工具链就绪 | 先补测试再收紧 lint 规则 |
| `no-explicit-any` → `error` | 存量 any 清理完成 | 逐步收紧 |
| `errorlint.errorf` → true | 存量 `%v` 切换为 `%w` | 后端 error wrapping 统一 |
| `fieldalignment` 启用 | 评估序列化兼容性 | 结构体内存优化 |
| `vuejs-accessibility` | 前端功能稳定 | 可访问性检查 |
