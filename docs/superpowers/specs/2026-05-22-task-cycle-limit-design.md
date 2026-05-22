# 周期任务提交限制设计

## 背景

PRD 在“周期任务规则”中明确提出：

- 每日任务按自然日限制提交次数。
- 每周任务按自然周限制提交次数。
- 同一个孩子、同一个任务、同一个周期内只能存在一条有效提交记录。
- 任务归档后保留历史记录，但不再出现在可执行任务列表中。

当前代码已经有 `ONCE`、`DAILY`、`WEEKLY` 三种任务周期，也能阻止同一个任务同时存在 `CLAIMED` 或 `PENDING` 记录。但当记录被审核通过后，同一个孩子可以在同一天或同一周再次领取/提交同一个任务，这会导致每日/每周积分被重复发放。

本次设计补齐周期限制，让任务周期真正参与领取和提交校验。

## 目标

1. `ONCE` 任务：同一个孩子同一个任务只能存在一条有效记录。
2. `DAILY` 任务：同一个孩子同一个任务在同一自然日只能存在一条有效记录。
3. `WEEKLY` 任务：同一个孩子同一个任务在同一自然周只能存在一条有效记录。
4. 有效记录包含 `CLAIMED`、`PENDING`、`APPROVED`。
5. `REJECTED` 不算有效记录，孩子被驳回后可以重新领取或提交。
6. 领取任务和直接提交任务都执行同一套周期限制。

## 不做范围

- 不增加“下一次可提交时间”字段。
- 不改小程序按钮展示状态；如果后续需要，可以在任务列表中返回当前孩子的周期可操作状态。
- 不处理历史脏数据迁移。
- 不引入复杂日历设置；自然日和自然周以数据库当前时间为准。

## 规则细节

### 周期锚点

周期判断使用 `task_records.created_at`。

原因：

- 领取和直接提交都会创建任务记录。
- `CLAIMED` 记录还没有可靠的 `submit_time`。
- 使用创建记录时间可以让“本周期已经开始做过这项任务”成为统一判断。

### 有效记录

有效记录状态：

- `CLAIMED`
- `PENDING`
- `APPROVED`

无效记录状态：

- `REJECTED`

驳回记录保留历史，但不阻塞同周期重新提交，便于孩子按家长意见修正后再次完成。

### 周期 SQL

`ONCE`：

```sql
SELECT COUNT(*)
FROM task_records
WHERE family_id=@p1
  AND task_id=@p2
  AND member_id=@p3
  AND status IN ('CLAIMED', 'PENDING', 'APPROVED')
```

`DAILY`：

```sql
... AND DATE(created_at)=CURRENT_DATE()
```

`WEEKLY`：

```sql
... AND YEARWEEK(created_at, 1)=YEARWEEK(CURRENT_DATE(), 1)
```

`YEARWEEK(..., 1)` 使用周一作为一周开始，更符合家庭周计划语义。

## 后端改动

将当前 `requireNoOpenTaskRecord(familyId, taskId, memberId)` 替换为：

```go
requireNoEffectiveTaskRecord(familyId, task, memberId)
```

其中 `task` 包含 `CycleType`。该方法根据 `CycleType` 拼接周期条件，并在发现有效记录时抛出：

```text
task record already exists
```

保持错误文案不变，避免影响已有前端和测试。

## 测试策略

1. 单元测试覆盖周期 SQL 片段生成：
   - `ONCE` 不带日期条件。
   - `DAILY` 带 `DATE(created_at)=CURRENT_DATE()`。
   - `WEEKLY` 带 `YEARWEEK(created_at, 1)=YEARWEEK(CURRENT_DATE(), 1)`。
2. 集成测试覆盖：
   - 每日任务审核通过后，同一天再次领取失败。
   - 每日任务被驳回后，同一天可以再次领取。
   - 一次性任务审核通过后，再次直接提交失败。

## 自查

- 与 PRD 周期规则一致。
- 不新增表结构。
- 领取和直接提交共用同一套限制。
- 驳回后允许重试，避免家长驳回造成当天任务无法修正。
