# GKube UI 重构设计文档

## 概述

将 GKube 前端从 Element Plus 默认蓝色主题升级为 Rancher 风格的专业 K8s 多集群管理界面，支持浅色/深色双主题。

## 目标

- 采用 Rancher 风格：蓝色主色调、清爽简洁、卡片化布局、圆角较多、现代感强
- 支持浅色 + 深色双主题切换
- 基于 Element Plus 深度定制，不更换 UI 库
- 建立完整的设计 Token 系统，消除硬编码颜色

## 当前状态

- Element Plus 默认蓝色主题，零定制
- 颜色全部硬编码，没有设计 token 系统
- 深色侧边栏 `#001529` + 浅色内容区 `#f0f2f5`
- 94 个 Vue 组件，全部用 scoped CSS，无预处理器
- 无 CSS 变量定义，无主题切换机制

## 技术方案

### 方案选择：Element Plus CSS 变量覆盖

**理由**：
- 风险最低，不需要更换 UI 库
- 渐进式迁移，可以逐页改造
- Element Plus 2.x 原生支持 CSS 变量覆盖
- 生态成熟，组件丰富

**不选择的方案**：
- SCSS 主题定制：引入预处理器复杂度
- 换用 Naive UI：需要重写所有组件引用
- UnoCSS 原子化：引入新工具，学习成本高

## 设计 Token 系统

### 文件结构

```
src/styles/
├── tokens.css              # 设计 token 定义（CSS 自定义属性）
├── themes/
│   ├── light.css           # 浅色主题变量
│   └── dark.css            # 深色主题变量
├── element-overrides.css   # Element Plus 变量覆盖
├── theme-switcher.ts       # 主题切换逻辑
└── index.css               # 统一入口
```

### Token 定义（Rancher 风格）

```css
:root {
  /* === 颜色系统 === */
  /* 主色 - Rancher 蓝 */
  --gk-color-primary: #3b82f6;
  --gk-color-primary-light: #60a5fa;
  --gk-color-primary-dark: #2563eb;
  --gk-color-primary-bg: #eff6ff;

  /* 语义色 */
  --gk-color-success: #22c55e;
  --gk-color-warning: #f59e0b;
  --gk-color-danger: #ef4444;
  --gk-color-info: #6366f1;

  /* 中性色 - 由主题覆盖 */
  --gk-color-text-primary: var(--gk-neutral-900);
  --gk-color-text-secondary: var(--gk-neutral-500);
  --gk-color-text-placeholder: var(--gk-neutral-400);
  --gk-color-bg-page: var(--gk-neutral-50);
  --gk-color-bg-card: var(--gk-white);
  --gk-color-bg-sidebar: var(--gk-neutral-900);
  --gk-color-border: var(--gk-neutral-200);

  /* === 间距系统（4px 基准） === */
  --gk-space-1: 4px;
  --gk-space-2: 8px;
  --gk-space-3: 12px;
  --gk-space-4: 16px;
  --gk-space-5: 20px;
  --gk-space-6: 24px;
  --gk-space-8: 32px;
  --gk-space-10: 40px;
  --gk-space-12: 48px;

  /* === 圆角系统 === */
  --gk-radius-sm: 4px;
  --gk-radius-md: 8px;
  --gk-radius-lg: 12px;
  --gk-radius-xl: 16px;
  --gk-radius-full: 9999px;

  /* === 阴影系统 === */
  --gk-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --gk-shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  --gk-shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  --gk-shadow-card: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.06);

  /* === 字体系统 === */
  --gk-font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --gk-font-mono: 'JetBrains Mono', 'Fira Code', monospace;
  --gk-font-size-xs: 12px;
  --gk-font-size-sm: 13px;
  --gk-font-size-base: 14px;
  --gk-font-size-lg: 16px;
  --gk-font-size-xl: 18px;
  --gk-font-size-2xl: 20px;
  --gk-font-size-3xl: 24px;

  /* === 布局 === */
  --gk-sidebar-width: 240px;
  --gk-sidebar-collapsed-width: 64px;
  --gk-header-height: 56px;
  --gk-content-max-width: 1400px;
}
```

### 浅色主题

```css
[data-theme="light"], :root {
  --gk-white: #ffffff;
  --gk-neutral-50: #f8fafc;
  --gk-neutral-100: #f1f5f9;
  --gk-neutral-200: #e2e8f0;
  --gk-neutral-300: #cbd5e1;
  --gk-neutral-400: #94a3b8;
  --gk-neutral-500: #64748b;
  --gk-neutral-600: #475569;
  --gk-neutral-700: #334155;
  --gk-neutral-800: #1e293b;
  --gk-neutral-900: #0f172a;

  --gk-color-bg-page: #f8fafc;
  --gk-color-bg-card: #ffffff;
  --gk-color-bg-sidebar: #0f172a;
  --gk-color-bg-header: #ffffff;
  --gk-color-text-primary: #0f172a;
  --gk-color-text-secondary: #64748b;
  --gk-color-border: #e2e8f0;
}
```

### 深色主题

```css
[data-theme="dark"] {
  --gk-white: #0f172a;
  --gk-neutral-50: #1e293b;
  --gk-neutral-100: #1e293b;
  --gk-neutral-200: #334155;
  --gk-neutral-300: #475569;
  --gk-neutral-400: #64748b;
  --gk-neutral-500: #94a3b8;
  --gk-neutral-600: #cbd5e1;
  --gk-neutral-700: #e2e8f0;
  --gk-neutral-800: #f1f5f9;
  --gk-neutral-900: #f8fafc;

  --gk-color-bg-page: #0f172a;
  --gk-color-bg-card: #1e293b;
  --gk-color-bg-sidebar: #020617;
  --gk-color-bg-header: #1e293b;
  --gk-color-text-primary: #f1f5f9;
  --gk-color-text-secondary: #94a3b8;
  --gk-color-border: #334155;
}
```

## Element Plus 变量覆盖

```css
/* element-overrides.css */
:root {
  /* 覆盖 Element Plus 主色 */
  --el-color-primary: var(--gk-color-primary);
  --el-color-primary-light-3: var(--gk-color-primary-light);
  --el-color-primary-light-5: var(--gk-color-primary-bg);
  --el-color-primary-dark-2: var(--gk-color-primary-dark);

  /* 覆盖 Element Plus 背景色 */
  --el-bg-color: var(--gk-color-bg-card);
  --el-bg-color-page: var(--gk-color-bg-page);
  --el-fill-color-blank: var(--gk-color-bg-card);
  --el-fill-color-light: var(--gk-neutral-100);
  --el-fill-color-lighter: var(--gk-neutral-50);

  /* 覆盖 Element Plus 文字色 */
  --el-text-color-primary: var(--gk-color-text-primary);
  --el-text-color-regular: var(--gk-color-text-primary);
  --el-text-color-secondary: var(--gk-color-text-secondary);
  --el-text-color-placeholder: var(--gk-color-text-placeholder);

  /* 覆盖 Element Plus 边框 */
  --el-border-color: var(--gk-color-border);
  --el-border-color-light: var(--gk-color-border);
  --el-border-color-lighter: var(--gk-color-border);
  --el-border-color-extra-light: var(--gk-neutral-100);

  /* 覆盖 Element Plus 圆角 */
  --el-border-radius-base: var(--gk-radius-md);
  --el-border-radius-small: var(--gk-radius-sm);
  --el-border-radius-round: var(--gk-radius-full);

  /* 覆盖 Element Plus 阴影 */
  --el-box-shadow: var(--gk-shadow-md);
  --el-box-shadow-light: var(--gk-shadow-sm);

  /* 覆盖 Element Plus 字体 */
  --el-font-family: var(--gk-font-sans);
  --el-font-size-base: var(--gk-font-size-base);
}
```

## 布局重设计

### 侧边栏

```
┌─────────────────────────────────────┐
│  [Logo] GKube          [收起按钮]    │  ← Header 56px
├──────────┬──────────────────────────┤
│          │                          │
│  侧边栏   │      内容区域              │
│  240px   │      padding: 24px       │
│          │                          │
│  ▸ 集群   │                          │
│  ▸ 工作负载 │                          │
│  ▸ 网络   │                          │
│  ▸ 存储   │                          │
│  ▸ 配置   │                          │
│          │                          │
│──────────│                          │
│  用户信息  │                          │
└──────────┴──────────────────────────┘
```

**改进点**：
- 背景：`var(--gk-color-bg-sidebar)` — 深色主题下更深
- Logo 区域：更大的 padding，品牌色突出
- 菜单项：hover 时显示浅色背景条，active 时左侧 3px 蓝色指示条
- 子菜单：缩进 + 图标，展开/收起动画更平滑
- 底部：用户头像 + 快捷操作

### Header

**改进点**：
- 高度：60px → 56px，更紧凑
- 左侧：侧边栏收起按钮 + 面包屑
- 右侧：集群选择器（带状态指示的下拉）+ 通知图标 + 主题切换按钮 + 用户头像
- 底部：1px 边框 + 微妙阴影

### 卡片组件

```css
.gk-card {
  background: var(--gk-color-bg-card);
  border-radius: var(--gk-radius-lg);
  box-shadow: var(--gk-shadow-card);
  padding: var(--gk-space-6);
  border: 1px solid var(--gk-color-border);
  transition: box-shadow 0.2s ease;
}

.gk-card:hover {
  box-shadow: var(--gk-shadow-md);
}

.gk-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--gk-space-4);
  padding-left: var(--gk-space-3);
  border-left: 3px solid var(--gk-color-primary);
}
```

## 页面级组件设计

### 列表页统一模式

```
┌─────────────────────────────────────────────┐
│  页面标题                         [+ 创建]    │  ← 页面头部
├─────────────────────────────────────────────┤
│  [命名空间 ▾] [状态 ▾] [搜索...]   [筛选]    │  ← 筛选栏
├─────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────┐│
│  │  资源卡片 1                              ││  ← 卡片列表
│  │  名称 | 状态 | 副本 | 创建时间            ││
│  └─────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────┐│
│  │  资源卡片 2                              ││
│  └─────────────────────────────────────────┘│
│  ...                                        │
│                              [分页]          │  ← 分页
└─────────────────────────────────────────────┘
```

**改进点**：
- 页面标题区：大标题 + 描述文字 + 创建按钮（蓝色主色）
- 筛选栏：统一的 `.gk-filter-bar` 组件，圆角搜索框 + 下拉筛选器
- 卡片列表：每个资源一张卡片，显示关键信息
- 状态指示：彩色圆点 + 文字（绿色=Running，黄色=Pending，红色=Failed）
- 分页：Element Plus 分页组件，居右对齐

### 详情页统一模式

```
┌─────────────────────────────────────────────┐
│  ← 返回  |  资源名称            [编辑] [删除] │  ← 详情头部
├─────────────────────────────────────────────┤
│  [概览] [事件] [日志] [终端]                  │  ← Tab 导航
├─────────────────────────────────────────────┤
│  ┌─────────────────────┐ ┌─────────────────┐│
│  │  基本信息            │ │  状态信息        ││  ← 双栏布局
│  │  命名空间: default   │ │  就绪: 3/3      ││
│  │  创建时间: 2024-...  │ │  更新: 3        ││
│  │  标签: app=nginx     │ │  可用: 3        ││
│  └─────────────────────┘ └─────────────────┘│
│  ┌─────────────────────────────────────────┐│
│  │  容器列表                                ││  ← 扩展区
│  │  ┌─────┐ ┌─────┐ ┌─────┐               ││
│  │  │容器1│ │容器2│ │容器3│               ││
│  │  └─────┘ └─────┘ └─────┘               ││
│  └─────────────────────────────────────────┘│
└─────────────────────────────────────────────┘
```

### 状态指示组件

```css
.gk-status {
  display: inline-flex;
  align-items: center;
  gap: var(--gk-space-2);
  font-size: var(--gk-font-size-sm);
}

.gk-status__dot {
  width: 8px;
  height: 8px;
  border-radius: var(--gk-radius-full);
}

.gk-status--running .gk-status__dot { background: var(--gk-color-success); }
.gk-status--pending .gk-status__dot { background: var(--gk-color-warning); }
.gk-status--failed  .gk-status__dot { background: var(--gk-color-danger); }
.gk-status--succeeded .gk-status__dot { background: var(--gk-color-primary); }
```

### 表格样式

Element Plus 表格的 Rancher 风格覆盖：
- 表头：浅灰背景，加粗文字，无边框
- 行：hover 时浅蓝背景，无斑马纹
- 单元格：更大的 padding（12px 16px）
- 圆角：表格整体 12px 圆角
- 边框：仅水平分隔线，无垂直边框

## 主题切换实现

### 主题切换器

```typescript
// src/styles/theme-switcher.ts
export type Theme = 'light' | 'dark'

const THEME_KEY = 'gk-theme'

export function getTheme(): Theme {
  return (localStorage.getItem(THEME_KEY) as Theme) || 'light'
}

export function setTheme(theme: Theme): void {
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
}

export function initTheme(): void {
  setTheme(getTheme())
}
```

### Element Plus 深色主题兼容

```typescript
// main.ts
import 'element-plus/theme-chalk/dark/css-vars.css'
```

### 主题切换按钮

```vue
<!-- Header.vue 中 -->
<el-button :icon="isDark ? Moon : Sunny" circle @click="toggleTheme" />
```

## 迁移策略

### 阶段 1：基础设施（第 1 周）

- 创建 `src/styles/` 目录结构
- 定义设计 token（tokens.css）
- 创建浅色/深色主题文件
- 配置 Element Plus 变量覆盖
- 实现主题切换器
- 在 main.ts 中初始化

### 阶段 2：布局组件（第 2 周）

- 重写 AppLayout.vue — 使用 token 变量
- 重写 Sidebar.vue — Rancher 风格导航
- 重写 Header.vue — 主题切换按钮 + 样式升级
- 创建 GkCard、GkStatusBadge 等基础组件

### 阶段 3：页面迁移（第 3-4 周）

- 按模块迁移：工作负载 → 网络 → 存储 → 配置 → 节点 → 命名空间
- 每个页面：替换硬编码颜色为 token 引用
- 统一列表页/详情页布局模式
- 测试浅色/深色主题在各页面的表现

### 阶段 4：细节打磨（第 5 周）

- 终端/日志查看器的主题适配
- 响应式布局优化
- 动画和过渡效果
- 跨浏览器测试

## 需要修改的文件清单

| 类别 | 文件 | 改动 |
|------|------|------|
| 新增 | `src/styles/tokens.css` | 设计 token 定义 |
| 新增 | `src/styles/themes/light.css` | 浅色主题变量 |
| 新增 | `src/styles/themes/dark.css` | 深色主题变量 |
| 新增 | `src/styles/element-overrides.css` | Element Plus 覆盖 |
| 新增 | `src/styles/theme-switcher.ts` | 主题切换逻辑 |
| 修改 | `src/main.ts` | 导入新样式，初始化主题 |
| 修改 | `src/components/Layout/AppLayout.vue` | 使用 token 变量 |
| 修改 | `src/components/Layout/Sidebar.vue` | Rancher 风格 |
| 修改 | `src/components/Layout/Header.vue` | 主题切换 + 样式 |
| 修改 | `src/style.css` | 移除重复样式，使用 token |
| 修改 | `src/App.vue` | 清理重复全局样式 |
| 修改 | 所有 94 个 `.vue` 文件 | 替换硬编码颜色 |

## 预期效果

- 专业的 Rancher 风格 K8s 管理界面
- 浅色/深色双主题支持
- 统一的设计语言和组件规范
- 更好的视觉层次和信息密度
- 现代化的交互体验
