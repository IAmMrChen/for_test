# 任务入口与首页快捷处理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 增加底部任务入口，迁移任务管理到任务页，首页保留待办和代孩子快捷处理，并让家长可代真实孩子和虚拟孩子操作任务。

**Architecture:** 后端只调整代孩子任务权限，不改变任务表结构。小程序新增 `pages/tasks/index.vue` 作为任务主入口，首页移除完整任务管理但保留代孩子快捷区，`pages.json` 增加任务 tab。前端同名任务只弹窗提醒并允许继续。

**Tech Stack:** Go 原生 SQL service 测试、uni-app Vue3、现有 task/member/dashboard/reward API、Node 源码行为测试、`node --check`、`go test`。

---

### Task 1: 后端允许家长代真实孩子任务操作

**Files:**
- Modify: `parent-child-api/srv/member_service_test.go`
- Modify: `parent-child-api/srv/member_service.go`
- Test: `go test ./srv -run TestMemberRoleHelpers -count=1`

- [ ] **Step 1: 写失败测试**

修改 `TestMemberRoleHelpers`，把真实孩子代操作期望改为允许：

```go
func TestMemberRoleHelpers(t *testing.T) {
	parent := model.FamilyMember{RoleType: model.FamilyRoleParent}
	child := model.FamilyMember{RoleType: model.FamilyRoleChild}

	if !memberCanSubmitForChild(parent, true) {
		t.Fatal("parent should submit for virtual child")
	}
	if !memberCanSubmitForChild(parent, false) {
		t.Fatal("parent should submit for real child")
	}
	if memberCanSubmitForChild(child, true) {
		t.Fatal("child should not submit for virtual child through parent helper")
	}
	if memberCanSubmitForChild(child, false) {
		t.Fatal("child should not submit for real child through parent helper")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
go test ./srv -run TestMemberRoleHelpers -count=1
```

Expected: FAIL，提示 `parent should submit for real child`。

- [ ] **Step 3: 修改权限 helper**

把 `memberCanSubmitForChild` 改为：

```go
func memberCanSubmitForChild(operator model.FamilyMember, targetIsVirtual bool) bool {
	return operator.RoleType.IsParentRole()
}
```

`targetIsVirtual` 参数保留，避免本次扩大改动面。

- [ ] **Step 4: 运行测试确认通过**

Run:

```powershell
go test ./srv -run TestMemberRoleHelpers -count=1
```

Expected: PASS。

- [ ] **Step 5: 提交后端权限调整**

```powershell
git add parent-child-api/srv/member_service.go parent-child-api/srv/member_service_test.go
git commit -m "feat: 允许家长代真实孩子操作任务"
```

### Task 2: 增加任务页源码行为测试

**Files:**
- Create: `parent-child-miniprogram/pages/tasks/index.behavior.test.mjs`
- Test: `node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs`

- [ ] **Step 1: 写失败测试**

创建 `parent-child-miniprogram/pages/tasks/index.behavior.test.mjs`：

```js
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const pageSource = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')
const pagesJson = fs.readFileSync(path.resolve(__dirname, '../../pages.json'), 'utf8')

assert.ok(pagesJson.includes('"path": "pages/tasks/index"'), 'pages.json 应注册任务页')
assert.ok(pagesJson.includes('"pagePath": "pages/tasks/index"'), '底部菜单应包含任务入口')
assert.ok(pageSource.includes('发布任务'), '任务页应提供发布任务入口')
assert.ok(pageSource.includes('任务管理'), '任务页应提供任务管理区域')
assert.ok(pageSource.includes('已存在同名任务，仍要继续吗？'), '任务页应对同名任务做前端提醒')
assert.ok(pageSource.includes('confirmDuplicateTaskTitle'), '任务页应通过确认函数允许同名任务继续提交')
assert.ok(pageSource.includes('archiveExistingTask'), '任务页应支持归档任务')
assert.ok(pageSource.includes('editTask'), '任务页应支持编辑任务')
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
```

Expected: FAIL，因为 `pages/tasks/index.vue` 还不存在。

- [ ] **Step 3: 提交失败测试**

```powershell
git add parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
git commit -m "test: 增加任务页行为约束"
```

### Task 3: 新增任务页和底部任务入口

**Files:**
- Create: `parent-child-miniprogram/pages/tasks/index.vue`
- Modify: `parent-child-miniprogram/pages.json`
- Test: `parent-child-miniprogram/pages/tasks/index.behavior.test.mjs`

- [ ] **Step 1: 注册任务页和 tab**

在 `pages.json` 中增加 `pages/tasks/index` 页面，并在 tabBar 中放到首页之后、奖励之前。

- [ ] **Step 2: 创建任务页**

任务页复用当前首页中的任务相关逻辑：

1. 家长视角展示发布任务、任务管理列表、编辑、归档。
2. 孩子视角展示可领取、已领取、审核中任务。
3. 创建和编辑任务时调用 `confirmDuplicateTaskTitle`，只提示，不强制阻断。
4. 成功后刷新任务列表。

- [ ] **Step 3: 实现前端同名提醒**

新增函数：

```js
async function confirmDuplicateTaskTitle(title) {
  const normalized = title.trim()
  const duplicated = tasks.value.some((task) => {
    return task.id !== editingTaskId.value && task.title.trim() === normalized
  })
  if (!duplicated) {
    return true
  }
  return new Promise((resolve) => {
    uni.showModal({
      title: '任务名称重复',
      content: '已存在同名任务，仍要继续吗？',
      confirmText: '继续',
      cancelText: '取消',
      success: (res) => resolve(res.confirm)
    })
  })
}
```

- [ ] **Step 4: 运行任务页行为测试**

Run:

```powershell
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 5: 提交任务页**

```powershell
git add parent-child-miniprogram/pages.json parent-child-miniprogram/pages/tasks/index.vue
git commit -m "feat: 增加小程序任务入口"
```

### Task 4: 增加首页行为测试

**Files:**
- Create: `parent-child-miniprogram/pages/index/index.behavior.test.mjs`
- Test: `node parent-child-miniprogram/pages/index/index.behavior.test.mjs`

- [ ] **Step 1: 写失败测试**

创建 `parent-child-miniprogram/pages/index/index.behavior.test.mjs`：

```js
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('代孩子完成任务'), '首页应保留代孩子完成任务快捷区')
assert.ok(source.includes('待发放奖励'), '首页应保留待发放奖励区')
assert.ok(source.indexOf('待发放奖励') < source.indexOf('代孩子完成任务'), '代孩子完成任务应位于待发放奖励之后')
assert.ok(source.includes("member.roleType === 'CHILD'"), '首页代操作应支持所有孩子')
assert.ok(!source.includes('member.isVirtual'), '首页代操作不应只筛选虚拟孩子')
assert.ok(source.includes('查看全部任务'), '首页代孩子快捷区应提供查看全部任务入口')
assert.ok(source.includes('/pages/tasks/index'), '首页应能跳转到任务页')
assert.ok(!source.includes('发布任务'), '首页不应展示发布任务表单')
assert.ok(!source.includes('任务管理'), '首页不应展示完整任务管理列表')
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
```

Expected: FAIL，因为当前首页仍包含发布任务和任务管理，并且仍按虚拟孩子筛选。

- [ ] **Step 3: 提交失败测试**

```powershell
git add parent-child-miniprogram/pages/index/index.behavior.test.mjs
git commit -m "test: 增加首页任务入口行为约束"
```

### Task 5: 重构首页任务区域

**Files:**
- Modify: `parent-child-miniprogram/pages/index/index.vue`
- Test: `parent-child-miniprogram/pages/index/index.behavior.test.mjs`

- [ ] **Step 1: 移除首页发布和完整任务管理**

从家长视角首页删除发布任务表单、任务管理列表、编辑和归档逻辑。

- [ ] **Step 2: 顶部统计增加跳转**

给三个统计项加点击：

1. 待审核：滚动到待审核区。
2. 待奖励：滚动到待发放奖励区。
3. 可用任务：跳转 `/pages/tasks/index`。

- [ ] **Step 3: 代孩子快捷区移到待发放奖励后**

将代孩子区域放在待发放奖励区之后。

- [ ] **Step 4: 支持所有孩子**

将孩子筛选从：

```js
member.roleType === 'CHILD' && member.isVirtual
```

改为：

```js
member.roleType === 'CHILD'
```

- [ ] **Step 5: 限制首页展示数量并加查看全部任务**

首页代孩子任务最多展示 5 条，提供“查看全部任务”按钮跳转任务页。

- [ ] **Step 6: 运行首页行为测试**

Run:

```powershell
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 7: 提交首页重构**

```powershell
git add parent-child-miniprogram/pages/index/index.vue
git commit -m "feat: 调整首页任务快捷区"
```

### Task 6: 最终验证

**Files:**
- Verify: `parent-child-api/srv/member_service.go`
- Verify: `parent-child-miniprogram/pages.json`
- Verify: `parent-child-miniprogram/pages/tasks/index.vue`
- Verify: `parent-child-miniprogram/pages/index/index.vue`

- [ ] **Step 1: 后端测试**

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: PASS。

- [ ] **Step 2: 前端行为测试**

```powershell
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 3: 前端 JS 语法检查**

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 4: 目标页面乱码检查**

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue parent-child-miniprogram\pages\tasks\index.vue parent-child-miniprogram\pages.json
```

Expected: exit 1。

## 自检

1. 设计要求均有任务覆盖：任务 tab、首页统计跳转、首页代孩子保留、代真实孩子、任务页管理、同名提醒且允许继续。
2. 服务端只改代操作权限，不新增任务名唯一限制。
3. 首页行为测试避免任务管理重新回到首页。
4. 任务页行为测试避免任务入口缺失或同名提醒缺失。
