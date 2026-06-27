# Vaultr · Neo Theme — Neo-Brutalist Design Spec

**Version:** 0.1 (draft)
**Scope:** `data-theme="neo"` — a standalone third theme parallel to `light` and `dark`. Always light-based, no dark variant.

---

## 1. 设计哲学 Design Philosophy

Neo 主题的核心风格是 **新野兽派（Neo-Brutalism）**：不是复古像素游戏，而是一种当代的图形化 UI 语言。它将极简主义的克制与粗犷的结构感结合，拒绝软化、模糊和过度优雅。

**三条核心原则：**

1. **结构即装饰** — 粗边框、硬阴影、0 圆角本身就是视觉语言，不需要额外的装饰元素。
2. **色彩高度克制** — 白底、黑线、一个高饱和品牌色。品牌色只作为结构色（nav 背景）和主操作色（primary CTA），绝不泛滥。
3. **交互有重量感** — 每个可点击元素都有"物理存在感"：阴影让它像一个可以按下的实体，hover/active 的 translate 是按下去的反馈。

**参考美学坐标：** Linear / Pieter Levels 早期项目 / 复古计算机手册印刷风格。

---

## 2. 色彩系统 Color System

### 2.1 基础色板

| Token            | 值        | 用途                       |
| ---------------- | --------- | -------------------------- |
| `--bg`           | `#ffffff` | 页面底色，纯白，无暖色偏移 |
| `--fg`           | `#111111` | 主文本，近黑               |
| `--muted`        | `#6b7280` | 次级文本                   |
| `--card-bg`      | `#f5f5f5` | 卡片背景                   |
| `--surface-soft` | `#f8f9fa` | 次级面板背景               |

### 2.2 品牌色 Brand Color

| Token          | 值          | 说明                                    |
| -------------- | ----------- | --------------------------------------- |
| `--accent`     | `#eab308`   | 主品牌色，清正黄（Tailwind yellow-500） |
| `--accent-hov` | `#ca9f07`   | hover 状态，暗一档                      |
| `--accent-rgb` | `234,179,8` | 用于 rgba() 半透明场景                  |

**品牌色使用规则：**
- ✅ Nav 侧边栏背景（最重要的结构色用途）
- ✅ Primary CTA 按钮填充色（choice-card、主操作按钮）
- ✅ Active 链接文字、Tab 激活指示
- ✅ 引用块左边框（低透明度）
- ✅ Knowledge Unit 卡片（`.graph-node-card.is-featured`）填充色
- ❌ 不用于普通卡片背景
- ❌ 不用于 Nav active 状态（黄底上的 active 用黑色填充）

### 2.6 语义分类色 Semantic Category Colors

少数特定语义元素允许使用固定的第三方色彩，不属于品牌色系但符合 Neo-Brutalist 饱和度要求：

| 元素                        | 色值      | 说明                                |
| --------------------------- | --------- | ----------------------------------- |
| `.graph-node-panel-type` chip | `#7c3aed` | 知识图谱实体类型标签，紫色（Violet-600） |

规则：分类色只用于**纯填充背景**（文字改为 `#ffffff`，边框仍为 `#000000`）。不用于边框或阴影。

### 2.3 边框色 Border Colors

| Token       | 值        | 用途                                  |
| ----------- | --------- | ------------------------------------- |
| `--card-bd` | `#000000` | 卡片、按钮、输入框边框，纯黑          |
| `--hr`      | `#333333` | 结构分割线（topbar、nav、panel 边界） |
| `--h2-rule` | `#333333` | Prose 中 h2 下方分隔线                |

> **设计决策：** 边框用纯黑 `#000`，不用半透明灰。这是 Neo-Brutalism 的核心特征——线条是线条，不是暗示。

### 2.4 阴影色 Shadow Color

| Token         | 值              | 说明                                |
| ------------- | --------------- | ----------------------------------- |
| `--px-shadow` | `rgba(0,0,0,1)` | **100% 不透明纯黑**，零模糊，硬偏移 |

> **为什么必须不透明：** 半透明阴影（如 0.82）在白色背景上会变成"灰色模糊"，失去印刷感。纯黑阴影让元素像被印在纸上的橡皮章。

### 2.5 导航色 Navigation Colors

| Token       | 值                | 说明                        |
| ----------- | ----------------- | --------------------------- |
| `--nav-bg`  | `#eab308`         | Nav 侧边栏背景，品牌黄      |
| `--nav-act` | `rgba(0,0,0,0.1)` | 黄色底上的 hover/press 叠层 |

---

## 3. 阴影系统 Shadow Tiers

硬阴影是 Neo-Brutalism 的核心视觉元素。使用三档，对应不同层级的 UI 元素。

```
--px-d1: 2px 2px 0    →  小型元素（小按钮、chip、状态指示）
--px-d2: 3px 3px 0    →  主要交互元素（按钮、输入框、消息气泡）
--px-d3: 4px 4px 0    →  浮层面板（modal、search panel、confirm）
```

### 3.1 三种交互模式

阴影不是统一的"所有卡片都有阴影"，而是根据元素的**语义**和**场景密度**决定：

**模式 A：按钮式（静止有阴影 → hover 按下）**

适用于稀疏场景的主操作元素（按钮、CTA chip、小型可点击 chip）。

```
静止:  box-shadow: var(--px-dN) var(--px-shadow)   ← 漂浮感
Hover: box-shadow: none; transform: translate(Npx, Npx)  ← 阴影消失 = 被按下
```

这个反转是关键：**阴影=漂浮，无阴影+位移=按下**，模拟物理按钮。

| 档位 | 元素                                                                  | Hover 位移          |
| ---- | --------------------------------------------------------------------- | ------------------- |
| d1   | 小按钮（`notif-play-btn`、`drawer-more-btn`、`toolbar-toggle`）       | translate(2px, 2px) |
| d2   | 主按钮（`chat-send-btn`、`cfg-save-btn`、`choice-card`、`mate-chip`） | translate(3px, 3px) |

---

**模式 B：密集内容卡片（静止无阴影 → hover 仅加阴影，无运动）**

适用于同屏存在多张的内容卡片（笔记格、图片格、文件夹格）。

密集场景规则：**阴影大小 ∝ 1 / 同屏密度**。当所有卡片都有静止阴影时，阴影失去区分意义变为视觉噪声。因此静止态靠 2px 黑边框定义"这是一张卡片"，hover 时阴影出现表示"这张卡片被注意到"。

```
静止:  box-shadow: none          ← 平放在桌面，边框定义卡片
Hover: box-shadow: var(--px-d3) var(--px-shadow)   ← 原地浮起，无位移
```

⚠️ **hover 时不改变背景色、不改变边框色、不产生位移**。只加阴影。

适用：`.card`、`.folder-card`（Home）、`.note-card`（Library）、`.img-card`（Images）

---

**模式 C：浮层面板（始终有阴影，无 hover 变化）**

适用于漂浮在页面上方的覆盖层，需要持续表达"与页面存在深度差"。

```
静止:  box-shadow: var(--px-d3) var(--px-shadow)   ← 始终漂浮
Hover: 无变化
```

适用：`.srch-panel`、`.settings-modal-panel`、`.confirm-card`、`.info-card`、`.drawer-more-menu`

---

### 3.2 不使用阴影的场景

| 场景                                     | 原因                                                                   |
| ---------------------------------------- | ---------------------------------------------------------------------- |
| **切换控件**（segmented control 激活项） | 黑色填充本身即足够强的选中信号                                         |
| **Icon-only 按钮**（T5）                 | 纯平面操作，hover 只改背景色                                           |
| **结构性容器**（`.cfg-section`）         | 容器内的子元素（输入框、按钮）已有各自的阴影，容器再加阴影造成叠加噪声 |
| **密集列表卡片静止态**                   | 见模式 B                                                               |
| **Agent / Mate 卡片列表**                | 密集设置列表，hover 改背景色即可                                       |

### 3.3 特殊规则

**Nav active item（黑色方块）**：虽然是 icon-only，但在黄色背景上黑色填充需要 d1 阴影来增强立体感。这是唯一不遵循"icon-only 无阴影"规则的例外。

**侧边栏列表项**（`.tags-col .tag-card`、`.settings-sidebar-item`、`.graph-index-item`）：hover 和 active 状态加 d1 阴影，表示"当前位置"。这是状态指示性阴影，不遵循密集卡片规则（因为这里的阴影是状态语义，不是内容强调）。

---

## 4. 边框系统 Border System

### 4.1 元素边框

所有有边框的 UI 元素在 neo 主题下一律使用 **2px** 宽度，配合 `--card-bd: #000`（纯黑）。

```css
/* 覆盖所有带边框的 card/button/input */
border-width: 2px !important;
```

覆盖范围：card、folder-card、note-card、tag-card、agent-card、mate-card、choice-card、img-card、cfg-section、confirm-card、short-card、info-card、srch-panel、lb-panel、chat-input-card、msg-user-bubble、所有 action buttons、所有 form inputs。

### 4.2 边框色统一规则

**所有元素在任何状态下（静止、hover、active、selected）的可见边框均使用 `#000000` 纯黑。**

这意味着：
- 选中态（`.is-open`、`.is-active`）：黑色边框，不随状态改变颜色
- Hover 态：黑色边框不变（状态通过背景色或阴影传达）
- 禁止使用半透明品牌色边框（`rgba(accent-rgb, X.XX)`）——那是光暗主题的过渡风格

> **核心逻辑：** 在 Neo-Brutalism 中，边框是结构线，不是状态指示器。状态（选中、激活）通过 **背景色** 和 **阴影** 传达，边框颜色始终保持纯黑，与背景形成最高对比度。

### 4.3 结构分割线

结构性边界（panel 之间的分隔）使用 **2px**，颜色 `--hr: #333`：

| 元素                     | 方向   |
| ------------------------ | ------ |
| `.lib-nav`               | 右边框 |
| `.lib-topbar`            | 下边框 |
| `.col-head`, `.col-body` | 右边框 |
| `.drawer-panel`          | 左边框 |
| `.img-toolbar`           | 下边框 |
| `.lb-sidebar`            | 左边框 |
| `.shorts-month-rail`     | 左边框 |
| `.graph-index-col`       | 右边框 |

### 4.4 圆角

```css
html[data-theme="neo"] * { border-radius: 0 !important; }
```

全局清零，无例外。Neo-Brutalism 没有圆角。

---

## 5. 导航系统 Navigation

### 5.1 Nav 侧边栏

黄色背景是整个 neo 主题视觉识别度最高的单个元素。

```
背景色:    --nav-bg = #eab308 (品牌黄)
图标色:    rgba(0,0,0,0.52)   (暗色，在黄底上清晰可见)
Hover:     color: #000;  background: rgba(0,0,0,0.1)
Active:    background: #000;  color: #fff  (黑色填充，高对比)
```

**Active 用黑色而非品牌色的原因：** 黄底+黄色高亮=不可见。黑色是黄色背景上唯一的高对比选项。黑色方块在黄条纹上也是极强的 Neo-Brutalist 视觉语言。

### 5.2 Nav 徽章（Run Badge）

```
背景: #000000 (黑色)
文字: #eab308 (品牌黄)
无渐变动画 (去掉彩虹渐变，保持图形感)
```

---

## 6. 按钮系统 Button System

### 6.1 Primary Button（主操作）

```
背景色:    --btn-primary-bg = #eab308
文字色:    --btn-primary-fg = #111111
边框:      2px solid #000
阴影:      var(--px-d2) var(--px-shadow) = 3px 3px 0 #000
Hover:     box-shadow: none; transform: translate(3px, 3px)
```

黄色填充 + 黑字 + 黑边框 + 黑硬阴影，是整个界面中视觉重量最重的操作元素。

**适用组件：** `choice-card`（首页快捷操作）、任何使用 `--btn-primary-bg` 的 filled button。

### 6.2 Secondary Button（次级操作）

```
背景色:    transparent
文字色:    --fg (#111111)
边框:      2px solid #000  (--card-bd)
阴影:      var(--px-d2) var(--px-shadow)
Hover:     box-shadow: none; transform: translate(3px, 3px)
```

**适用组件：** `cfg-save-btn`, `cfg-discard-btn`, `settings-apply-btn`, `agents-toolbar-btn`, `home-nav-btn`, `mate-save-btn`。

### 6.3 Small Chip Button（行内小按钮）

```
边框:      2px solid #000
阴影:      var(--px-d1) var(--px-shadow) = 2px 2px 0 #000
Hover:     box-shadow: none; transform: translate(2px, 2px)
```

**适用组件：** `note-read-btn`, `note-knowledge-btn`, `mate-act-btn`, `cal-nav-btn`, `toolbar-toggle`。

---

### 6.4 3D 按钮 Hover 背景规则（Mode A 通用约束）

**所有带像素阴影的 3D 按钮（§3.1 Mode A），hover 时绝不改变背景色。**

hover 的唯一视觉变化是阴影缩小 + `transform: translate` 下压（"半按下"状态）。

**实现要求：** base 组件 CSS 通常有 `:hover { background: var(--card-hov) }` 等规则。neo.css 的 hover 覆盖必须显式写 `background: <静止值>` 来取消 base 的背景色变化：

```css
/* ✅ 正确 */
html[data-theme="neo"] .some-3d-btn:hover {
  background: var(--bg);  /* 取消 base 的 background: var(--card-hov) */
  box-shadow: var(--px-d1) var(--px-shadow);
  transform: translate(1px, 1px);
}
```

| 按钮静止背景 | hover 应写 |
|---|---|
| `transparent` | `background: transparent` |
| `var(--bg)` | `background: var(--bg)` |
| `var(--btn-primary-bg)` 黄色 | `background: var(--btn-primary-bg)` |
| `var(--p2)` 紫色 | `background: var(--p2)` |
| `var(--accent)` 品牌黄 | `background: var(--accent)` |

**例外：** 带固定填充色的 danger 按钮（如 `.mate-act-btn.del`）在 hover 时需显式保持填充色，同样不得引入背景色变化。

---

### 6.5 Icon-only Button（纯图标按钮）

```
hover:     background: rgba(0,0,0,0.06)  (平面背景色变化，无阴影、无位移)
无 translate，无 box-shadow
```

**适用组件：** `lib-action-btn`（topbar 图标）、`lib-back-btn`、drawer 工具按钮。

---

## 6.6 侧边栏列表项 Sidebar List Items

侧边栏导航列表（标签列、图谱索引列、设置导航）的统一交互语言：

```
hover:   background: var(--bg) (#fff)；box-shadow: var(--px-d1) var(--px-shadow)
         — 白色卡片从灰色面板背景中"抬起"，轻浮感

active:  background: var(--accent) (#eab308)；color: #111111
         box-shadow: var(--px-d1) var(--px-shadow)
         — 品牌黄填充 + 黑色文字 + 黑色硬阴影，标记当前位置
```

**适用组件：**
- `.tags-col .tag-card`（Library 左侧标签/索引列）
- `.settings-sidebar-item`（Settings 弹窗左侧导航）
- `.graph-index-item`（Graph 页索引列）

**区别于切换控件：** 列表项用品牌黄作为 active 状态（表达"当前位置"），切换控件用黑色填充（表达"当前选项"）。

---

## 6.7 切换控件 Segmented / Toggle Control

所有多选一切换场景使用统一视觉语言：

```
容器:      border: 2px solid #000; background: transparent;
激活项:    background: #000; color: #fff; box-shadow: none;
未激活 hover: background: rgba(0,0,0,0.06); color: var(--fg);
```

**禁止使用凹陷效果（`inset box-shadow`）作为激活状态**。凹陷感是物理按压反馈，不适合表达"已选中"状态。

**统一适用：**
- `.conv-seg` / `.conv-seg-btn`（Chat / Trigger 切换）
- `.theme-seg` / `.theme-seg-btn`（Light / Dark / Neo 主题切换）
- `.mate-et-pill`（事件类型选择）
- `.tags-col .col-mode-tab`（Library 侧边栏模式切换）
- `.graph-index-item.active`（Graph 索引侧边栏选中项）
- `.effect-card.active`（视觉效果卡选中）
- `.settings-sidebar-item.active`（设置侧边栏选中项）

---

## 7. 表单输入 Form Inputs

```
边框:      2px solid #000
阴影:      var(--px-d2) var(--px-shadow) = 3px 3px 0 #000
圆角:      0 (全局清零)
```

输入框的硬阴影给它"印章感"，像一个物理上可以操作的区域。这是参考设计（Raft.build 登录页）中最显著的表单处理方式。

**适用元素：** `settings-input`, `cfg-input`, `cfg-textarea`, `mate-form-input`, `mate-form-select`, `mate-form-textarea`。

---

## 8. 排版 Typography

### 8.1 字体变量架构

三个 CSS 变量控制全局字体分工，均定义在 `appTokensNeo`：

| 变量          | 字体               | 用途                                 |
| ------------- | ------------------ | ------------------------------------ |
| `--font-ui`   | **Space Grotesk**  | UI chrome（body 继承，覆盖全部界面） |
| `--font-sans` | **Inter**          | 阅读内容区（prose、ProseMirror）     |
| `--font-mono` | **JetBrains Mono** | 代码块、路径、等宽场景               |

**级联逻辑：**

```css
body { font-family: var(--font-ui); }          /* Space Grotesk 覆盖全局 */
.prose, .ProseMirror { font-family: var(--font-sans); }  /* 内容区回退 Inter */
```

UI chrome 元素（按钮、导航、表单、搜索框、标签、工具栏）无需单独声明字体，直接从 `body` 继承 Space Grotesk。只有明确需要 Inter 的内容区才显式覆盖。

### 8.2 例外：像素点缀字体

"Press Start 2P" 仅用于 `.home-hero-title`（页面大标题），作为像素风的点睛之笔，不扩散到其他元素。

### 8.3 过去架构的问题（已修复）

旧方案将 `body` 设为 Inter，然后在 `neo.css` 里用 ~80 条选择器逐一覆盖为 Space Grotesk。这是维护负担：每新增一个 UI 组件都需要手动添加选择器。现行方案翻转级联方向，在 `body` 层一次到位，prose 区域按需回退。

---

## 9. 动效原则 Motion Principles

Neo-Brutalism 的核心动效语言是**物理位移反馈**：`transform: translate(Xpx, Ypx)` 配合 `box-shadow: none`，模拟按钮被按下的物理感。这是状态切换，而非装饰性动画。

Transition 的使用由各组件按需自行决定，规范不强制开启或关闭。

---

## 10. 内容元素 Content Elements

### Prose（阅读内容区）

| 元素     | Neo 处理                                  |
| -------- | ----------------------------------------- |
| 无序列表 | `list-style-type: square`（方形而非圆形） |
| 引用块   | 2px 左边框，品牌黄低透明度                |
| 代码块   | d2 硬阴影                                 |
| 表格     | d2 硬阴影                                 |
| 复选框   | 方形，无圆角，像素风对勾                  |

### Graph Node Panel（知识图谱节点面板）

右侧滑出面板包含两类卡片，交互模式不同：

| 卡片                            | 模式   | 静止阴影             | Hover                              | 背景            |
| ------------------------------- | ------ | -------------------- | ---------------------------------- | --------------- |
| `.graph-node-card.is-featured`  | Mode A | d2 硬阴影            | `box-shadow: none; translate(3,3)` | 品牌黄 `#eab308` |
| `.graph-node-card`（connected） | Mode B | 无阴影               | d3 阴影，无位移，无背景变化        | `--card-bg`     |

**实体类型 chip（`.graph-node-panel-type`）：** 紫色填充（`#7c3aed`）+ 白色文字 + 2px 黑边框。见 §2.6。

### Search Overlay

搜索面板使用 d3 阴影（最大档），像一个漂浮在页面上方的实体盖板。Keyboard 高亮状态：d1 阴影 + 品牌黄左侧 3px inset 边框。

---

## 11. 禁止事项 Don'ts

| ❌ 禁止                                         | ✅ 替代                                              |
| ---------------------------------------------- | --------------------------------------------------- |
| 任何 `border-radius`（已全局清零）             | 直角                                                |
| 半透明阴影（rgba alpha < 1）                   | 纯黑 `rgba(0,0,0,1)`                                |
| `box-shadow` 带 blur radius                    | 0 blur，纯偏移                                      |
| `inset 2px 2px 0`（凹陷效果）作为选中/激活状态 | 黑色填充 `background: #000; color: #fff`            |
| `inset` 对角阴影作为 hover 反馈                | `background: rgba(0,0,0,0.06)` 背景色变化           |
| 渐变色背景                                     | 纯色平铺                                            |
| 品牌黄用于 active 状态（黄底上）               | 黑色填充                                            |
| 像素风图标变体（lib-ai-px）                    | 统一使用标准 smooth 图标                            |
| 多个品牌色叠加使用                             | 品牌黄只做 nav 背景和 primary CTA                   |
| 密集卡片全部加静止阴影                         | 静止无阴影，hover 时才加阴影（见 §3.1 模式 B）      |
| hover 时改变卡片背景色或边框色                 | hover 只加阴影，其余属性保持静止值                  |
| hover 时对密集卡片使用 `transform: translate`  | 位移仅用于按钮式（模式 A），内容卡片 hover 原地不动 |
| 给结构容器（`cfg-section`）加阴影              | 结构容器无阴影，内部交互元素各有自己的阴影          |
| `rgba(--accent-rgb, 0.X)` 半透明品牌色边框     | `#000000` 纯黑边框，状态通过背景色和阴影传达        |

> **关于 `inset 3px 0 0`（水平色条）**：左侧/右侧纯水平方向的 inset 色条（如 `inset 3px 0 0 var(--link)`）是**装饰线**而非凹陷效果，允许保留。仅禁止带斜向偏移的 `inset X Y 0` 形式。

---

## 12. CSS Token 速查 Token Reference

```css
/* neo 主题在 shared_tokens.go 中定义（appTokensNeo），唯一主题，无继承： */

--px-shadow:      rgba(0,0,0,1)       /* 阴影色：纯黑不透明 */
--px-d1:          2px 2px 0           /* 阴影偏移 tier-1：小元素 */
--px-d2:          3px 3px 0           /* 阴影偏移 tier-2：主元素 */
--px-d3:          4px 4px 0           /* 阴影偏移 tier-3：大容器 */

--hr:             #333333             /* 结构分割线 */
--h2-rule:        #333333             /* prose h2 分割线 */
--card-bd:        #000000             /* 元素边框，纯黑 */

--nav-bg:         #eab308             /* nav 背景，品牌黄 */
--nav-act:        rgba(0,0,0,0.1)     /* nav hover 叠层 */

--accent:         #eab308             /* 品牌色 */
--accent-rgb:     234,179,8
--accent-hov:     #ca9f07

--btn-primary-bg: #eab308             /* 主按钮背景，品牌黄 */
--btn-primary-fg: #111111             /* 主按钮文字，近黑 */
--btn-primary-hover: #ca9f07
--btn-primary-active: #b88a10

--font-ui:   "Space Grotesk", system-ui, sans-serif   /* UI chrome 字体 */
--font-sans: "Inter", system-ui, sans-serif            /* 阅读内容字体 */
--font-mono: "JetBrains Mono", ui-monospace, monospace /* 代码字体 */
```

---

## 13. 文件对照 File Reference

| 职责                             | 文件                                                     |
| -------------------------------- | -------------------------------------------------------- |
| Neo token 定义                   | `internal/server/view/shared_tokens.go` → `appTokensNeo` |
| Neo 结构样式（阴影、边框、动效） | `internal/server/view/assets/neo.css`                    |
| Nav 颜色规则                     | `internal/server/view/shared_chrome.go` → `navCSS`       |
| 主题切换逻辑（Alpine store）     | `internal/server/view/shared_theme.go`                   |
| 设置弹窗 Neo 按钮                | `internal/server/view/shared_settings_modal.go`          |
| Electron 跨视图同步              | `desktop-app/src/main.js`                                |

---

## 14. CSS 分层架构 CSS Layering Architecture

### 核心判断标准

> **差异能否用 token 表达？**
> - 能 → 规则写入 **base 组件 CSS**
> - 不能 → 规则写入 **neo.css 覆盖层**

Token 是主题机制的载体。把颜色、背景、字号等外观差异编码进 token，base 组件 CSS 使用 `var(--token)` 引用，token 值随主题自动切换，不需要覆盖层介入。

---

### Base 组件 CSS（颜色 / 外观层）

以下类型的规则属于 base，即使它们的值在 Neo 主题下有"品牌风格"：

| 规则类型 | 示例 |
|----------|------|
| Hover 背景 | `.cfg-card:hover { background: var(--card-hov); }` |
| Icon 点击区背景 | `.lib-action-btn:hover { background: var(--icon-hov); }` |
| 文字颜色 | `.lib-action-btn:hover { color: var(--fg); }` |
| 字体大小 | `.home-hero-title { font-size: clamp(1.2rem, 2.2vw, 1.8rem); }` |
| 边框颜色 | `.chip:hover { border-color: var(--card-bd); }` |

只要规则的所有值都是 `var(--token)` 或主题无关的固定值（如 `0`、`100%`），它就属于 base。

---

### neo.css 覆盖层（结构 / 行为层）

以下类型的规则无法通过 token 表达——它们改变元素的物理形状或动效行为：

| 规则类型 | 示例 |
|----------|------|
| 全局零圆角 | `* { border-radius: 0 !important; }` |
| 像素阴影（d1/d2/d3） | `box-shadow: var(--px-d2) var(--px-shadow)` |
| 硬边框宽度 | `border-width: 2px !important` |
| 禁用过渡动画 | `transition: none` |
| 像素位移交互 | `transform: translate(3px, 3px)` on hover/press |
| 像素字体 | `font-family: "Press Start 2P", monospace` |
| UI 字体覆盖 | `body { font-family: var(--font-ui); }` |
| 禁用动画 | `animation: none` |
| 去除模糊 | `backdrop-filter: none` |
| 自定义 SVG 图形 | pixel checkmark、pixel dropdown arrow |

---

### 反模式：死代码（Dead Code）

如果 base 中有一条规则，而 neo.css 中有一条更高优先级的规则覆盖了它的**每一个属性**，那么 base 中的规则是死代码，永远不会生效。

**识别方法**：在 neo.css 中搜索同一选择器，看它是否覆盖了 base 规则的全部属性。若是，则将 base 规则修改为正确值，并删除 neo.css 覆盖。

**示例**：
```css
/* base — 原本错误，被 neo.css 纠正，形成死代码 */
.settings-modal-close-btn:hover { background: var(--nav-act); }

/* neo.css — 纠正覆盖 */
html[data-theme="neo"] .settings-modal-close-btn:hover { background: var(--icon-hov); }
```
正确做法：直接在 base 写 `background: var(--icon-hov);`，删除 neo.css 覆盖。

---

### CSS 优先级陷阱

neo.css 使用 `html[data-theme="neo"] .selector` 选择器，其优先级高于 base 中的 `.selector`。因此：

- 修改 base 规则后，如果 neo.css 中存在相同属性的覆盖，base 的修改**不会生效**。
- 每次修改 base 的外观属性，都应先搜索 neo.css 中是否存在同选择器的覆盖规则。
- 若存在：先判断覆盖是否仍然必要（结构性？还是只是纠正错误 token？），再决定删除或保留。

---

### Token 速查：正确的 hover 背景选择

| 场景 | 正确 token |
|------|-----------|
| Icon-only 小按钮（关闭、刷新、更多） | `--icon-hov` |
| 可展开配置卡片 | `--card-hov` |
| Mate / Skill / Agent 列表卡 | `--card-hov` |
| 下拉列表行 | `--card-hov` |
| 标签页切换（未激活） | `--icon-hov` |
| 导航侧栏（黄色背景上） | `--nav-act`（黑色叠层，仅用于 nav） |

**规则**：`--nav-act` 仅用于黄色导航背景上的元素。白色背景上的所有 hover 背景一律使用 `--icon-hov` 或 `--card-hov`，绝不使用 `--nav-act` 或 `--surface-soft`。

---

*本文档随设计迭代持续更新。每次对 neo 主题做实质性调整时同步修订对应章节。*
