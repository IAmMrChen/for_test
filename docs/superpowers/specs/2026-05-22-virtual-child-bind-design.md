# 虚拟孩子关联真实账号设计

## 背景

PRD 明确要求虚拟孩子未来可以通过邀请关联为真实孩子账号，并要求关联后成员 ID 不变、历史任务记录、积分流水和兑换记录继续保留。当前系统已经支持创建虚拟孩子，也支持邀请真实微信用户作为新成员加入家庭，但还缺少“把已有虚拟孩子绑定到真实微信用户”的能力。

现有 `family_invites` 表表达的是“新成员加入家庭”。虚拟孩子关联表达的是“已有成员绑定用户”，两者业务语义不同，所以本次新增独立的绑定邀请模型，避免把两种流程混在同一张邀请表里。

## 目标

1. 家长可以为家庭内某个虚拟孩子生成绑定邀请 token。
2. 被邀请微信用户接受绑定邀请后，系统把该虚拟孩子绑定到当前用户。
3. 绑定后 `family_members.id` 不变。
4. 绑定后 `family_members.user_id` 写入真实用户 ID，`is_virtual` 变为 false。
5. 绑定后历史任务记录、奖品兑换记录、积分流水不迁移，因为它们仍指向同一个 member id。
6. 如果当前用户已经是该家庭成员，必须拒绝绑定。
7. 如果目标成员不是虚拟孩子、已绑定用户、已移除或不属于该家庭，必须拒绝生成或接受绑定邀请。

## 不做范围

- 不做微信卡片真实分享能力，只生成可复制 token。
- 不做绑定邀请撤销。
- 不做绑定后解绑。
- 不做成员详情页独立页面；MVP 先在个人中心成员列表上提供轻量入口。

## 数据模型

新增表 `family_child_bind_invites`：

- `id`
- `family_id`
- `child_member_id`
- `inviter_member_id`
- `token`
- `status`
- `expires_at`
- `accepted_by_user_id`
- `accepted_at`
- `created_at`
- `updated_at`

状态复用邀请语义：

- `ACTIVE`
- `ACCEPTED`
- `EXPIRED`
- `CANCELED`

本次只使用 `ACTIVE` 和 `ACCEPTED`，预留其余状态。

## 后端接口

新增两个接口：

- `POST /api/member/createVirtualChildBindInvite`
- `POST /api/member/acceptVirtualChildBindInvite`

创建请求：

```json
{
  "familyId": 1,
  "memberId": 10
}
```

接受请求：

```json
{
  "token": "..."
}
```

创建成功返回绑定邀请对象。接受成功返回绑定后的家庭成员对象。

## 服务规则

创建绑定邀请：

1. 操作者必须是该家庭家长角色。
2. `memberId` 必须是该家庭内 active 的 child 成员。
3. 目标成员必须 `is_virtual=true` 且 `user_id IS NULL`。
4. 生成 7 天有效 token。

接受绑定邀请：

1. token 必须存在且处于 `ACTIVE`。
2. token 未过期。
3. 当前用户在该家庭内不能已有 active 身份。
4. 目标虚拟孩子必须仍是 active、child、virtual、未绑定用户。
5. 同一事务内更新邀请状态和成员绑定状态。

错误文案保持清晰：

- `bind invite not found`
- `bind invite expired`
- `user already has a family identity`
- `virtual child not found`
- `bind invite has been changed`

## 小程序入口

个人中心成员列表中，家长看到虚拟孩子时显示“邀请关联”按钮。

点击后：

1. 调用创建绑定邀请接口。
2. 展示 token、过期时间。
3. 提供复制 token 按钮。

家庭选择页增加“关联虚拟孩子”表单：

1. 用户粘贴绑定 token。
2. 调用接受绑定邀请接口。
3. 成功后将返回成员所在家庭写入当前家庭缓存，并进入首页。

## 测试策略

1. 单元测试覆盖创建绑定邀请缺少 member id。
2. 单元测试覆盖 token insert 参数。
3. 集成测试覆盖：
   - 创建虚拟孩子后生成绑定邀请。
   - 接受绑定邀请后 member id 不变。
   - 当前积分不变。
   - `is_virtual=false`，`user_id` 写入接受用户。
   - 同一用户再次接受普通家庭邀请被单身份规则阻止。
   - 已是家庭成员的用户接受绑定邀请被拒绝。
4. 前端做 JS 语法检查和 Vue 乱码扫描。

## 自查

- 使用独立绑定邀请表，避免污染普通家庭邀请语义。
- 绑定更新原成员，不创建新成员，满足历史数据保留要求。
- 单家庭单身份规则在接受绑定时强校验。
- MVP 只做 token 复制和粘贴，不依赖微信分享能力。
