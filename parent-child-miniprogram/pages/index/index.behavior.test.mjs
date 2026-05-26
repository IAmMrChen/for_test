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
assert.ok(!source.includes('查看全部任务'), '首页代孩子快捷区不应提供查看全部任务入口')
assert.ok(source.includes('/pages/tasks/index'), '首页应能跳转到任务页')
assert.ok(!source.includes('发布任务'), '首页不应展示发布任务表单')
assert.ok(!source.includes('任务管理'), '首页不应展示完整任务管理列表')
assert.ok(source.includes('支持真实孩子和虚拟孩子'), '首页代孩子模块应明确支持真实孩子和虚拟孩子')
assert.ok(source.includes('loadError'), '首页应有加载错误态，避免接口异常白屏')
assert.ok(!source.includes('throw error'), '首页不应把加载错误直接抛到页面运行时')
assert.ok(source.includes('applyClaimedTaskRecord'), '首页领取任务后应更新本地状态而不是整页刷新')
assert.ok(source.includes('applySubmittedTaskRecord'), '首页提交任务后应更新本地状态而不是整页刷新')
assert.ok(source.includes('taskClaims'), '首页孩子视图应维护循环任务领取关系')
assert.ok(source.includes('startTaskClaim'), '首页孩子视图领取循环任务应调用开始领取接口')
assert.ok(source.includes('stopTaskClaim'), '首页孩子视图应支持停止领取循环任务')
assert.ok(source.includes('停止领取'), '首页孩子视图应展示停止领取按钮')
assert.ok(source.includes('directApproveProxyTask'), '首页代孩子完成任务应支持直接通过')
assert.ok(source.includes("task.viewStatus === 'recurringActive'"), '首页代孩子完成循环任务时应先提交当期记录再通过')
assert.ok(source.includes('applyApprovedTaskRecord'), '审核通过后应本地更新状态而不是整页刷新')
assert.ok(source.includes('removeRewardRecord'), '奖励发放或拒绝后应本地移除记录而不是整页刷新')
assert.ok(!source.includes('await loadHome()'), '首页操作成功后不应整页刷新')
