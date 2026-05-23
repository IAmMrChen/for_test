# 家庭入口与邀请流重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 调整小程序选择家庭页的信息架构，让创建家庭入口常驻，邀请码入口降权，移除选择家庭页的关联孩子入口，并把创建人昵称默认传为“家主”。

**Architecture:** 只修改小程序前端，不改服务端接口。新增一个轻量 Node 源码行为测试，直接检查 `family-select/index.vue` 的关键产品约束，避免这类纯页面结构问题缺少自动化保护。

**Tech Stack:** uni-app Vue3、现有 `api/family.js`/`api/invite.js`、Node 静态测试、`node --check`。

---

### Task 1: 增加选择家庭页行为测试

**Files:**
- Create: `parent-child-miniprogram/pages/family-select/index.behavior.test.mjs`
- Test: `node parent-child-miniprogram/pages/family-select/index.behavior.test.mjs`

- [ ] **Step 1: 写失败测试**

创建 `parent-child-miniprogram/pages/family-select/index.behavior.test.mjs`：

```js
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('已有邀请码？'), '选择家庭页应保留低权重的邀请码入口')
assert.ok(source.includes("nickname: '家主'"), '创建家庭应固定传入家主昵称')
assert.ok(!source.includes('acceptVirtualChildBindInvite'), '选择家庭页不应处理虚拟孩子绑定')
assert.ok(!source.includes('showBindForm'), '选择家庭页不应展示绑定表单')
assert.ok(!source.includes('toggleBindForm'), '选择家庭页不应有绑定入口切换逻辑')
assert.ok(!source.includes('bindForm'), '选择家庭页不应维护绑定码表单状态')
assert.ok(!source.includes('关联孩子'), '选择家庭页不应展示关联孩子入口')
assert.ok(!source.includes('我的昵称'), '创建家庭表单不应要求输入创建人昵称')
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
node parent-child-miniprogram/pages/family-select/index.behavior.test.mjs
```

Expected: FAIL，至少因为当前页面仍包含 `acceptVirtualChildBindInvite`、`showBindForm`、`关联孩子` 或未固定传 `nickname: '家主'`。

- [ ] **Step 3: 提交失败测试**

```powershell
git add parent-child-miniprogram/pages/family-select/index.behavior.test.mjs
git commit -m "test: 增加家庭入口行为约束"
```

### Task 2: 重构选择家庭页

**Files:**
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`
- Test: `parent-child-miniprogram/pages/family-select/index.behavior.test.mjs`

- [ ] **Step 1: 改造模板结构**

将三个并列按钮改为：

1. 标题区常驻“创建家庭”按钮。
2. 家庭列表或空状态作为页面主体。
3. “已有邀请码？”作为低权重文本按钮。
4. 删除“关联孩子”按钮和绑定码表单。
5. 创建家庭表单只保留家庭名称输入。

- [ ] **Step 2: 改造脚本状态**

删除 `acceptVirtualChildBindInvite` 导入，删除 `showBindForm`、`submittingBind`、`bindForm`、`defaultBindForm`、`toggleBindForm`、`cancelAcceptBind`、`submitAcceptBind`。

`submitCreateFamily` 改为：

```js
const response = await createFamily({ name, nickname: '家主' })
```

- [ ] **Step 3: 改造样式**

删除三按钮 tab 化样式，新增或调整：

1. 标题区按钮布局。
2. 低权重邀请码入口样式。
3. 空状态中主按钮样式。
4. 表单样式保持与现有页面一致。

- [ ] **Step 4: 运行行为测试确认通过**

Run:

```powershell
node parent-child-miniprogram/pages/family-select/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 5: 运行小程序 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 6: 检查目标页面没有明显乱码片段**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\family-select\index.vue
```

Expected: exit 1。

- [ ] **Step 7: 提交页面重构**

```powershell
git add parent-child-miniprogram/pages/family-select/index.vue
git commit -m "feat: 重构家庭选择入口"
```

### Task 3: 最终验证

**Files:**
- Verify: `parent-child-miniprogram/pages/family-select/index.vue`
- Verify: `parent-child-miniprogram/pages/family-select/index.behavior.test.mjs`

- [ ] **Step 1: 运行行为测试**

```powershell
node parent-child-miniprogram/pages/family-select/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 2: 运行后端全量测试**

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: PASS。

- [ ] **Step 3: 确认未提交内容只剩用户本地配置或为空**

```powershell
git status --short
```

Expected: 如果 `parent-child-miniprogram/manifest.json` 仍显示修改，应确认它是用户本地 AppID 配置，不纳入本次提交。

## 自检

1. 设计文档要求均有任务覆盖：创建家庭常驻、昵称默认“家主”、邀请码降权、移除选择家庭页关联孩子、保留成员页绑定能力。
2. 本计划不修改服务端，也不实现微信分享卡片，符合非目标。
3. 行为测试覆盖核心产品约束，页面语法和后端测试作为最终验证。
