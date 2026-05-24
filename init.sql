-- 数据库初始化脚本
-- 字符集: utf8mb4
-- 排序规则: utf8mb4_unicode_ci

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 1. 用户表 (Users)
-- 存储微信授权后的基础信息
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `openid` VARCHAR(64) NOT NULL COMMENT '微信OpenID',
  `nickname` VARCHAR(64) DEFAULT NULL COMMENT '微信昵称',
  `avatar_url` VARCHAR(255) DEFAULT NULL COMMENT '微信头像URL',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_openid` (`openid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户基础信息表';

-- 2. 家庭表 (Families)
-- 家庭主体信息
CREATE TABLE IF NOT EXISTS `families` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `name` VARCHAR(64) NOT NULL COMMENT '家庭名称',
  `creator_id` BIGINT UNSIGNED NOT NULL COMMENT '创建者用户ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_creator` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='家庭信息表';

-- 3. 家庭成员表 (Family Members)
-- 核心表：连接用户与家庭，定义角色与权限
CREATE TABLE IF NOT EXISTS `family_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '所属家庭ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联真实用户ID (虚拟账号此字段为NULL)',
  `role_type` ENUM('OWNER', 'ADMIN', 'PARENT', 'CHILD') NOT NULL COMMENT '角色: OWNER-家主, ADMIN-超管, PARENT-普通家长, CHILD-孩子',
  `status` ENUM('ACTIVE', 'REMOVED') NOT NULL DEFAULT 'ACTIVE' COMMENT '成员状态: ACTIVE-正常, REMOVED-已移除',
  `nickname` VARCHAR(64) NOT NULL COMMENT '家庭内昵称 (如: 爸爸, 大宝)',
  `current_points` INT NOT NULL DEFAULT 0 COMMENT '当前积分余额 (仅孩子有效)',
  `total_earned_points` INT NOT NULL DEFAULT 0 COMMENT '累计获得积分 (仅孩子有效)',
  `is_virtual` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否虚拟账号: 0-否, 1-是',
  `created_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人ID (通常用于标记虚拟账号是谁建的)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_family_user` (`family_id`, `user_id`),
  KEY `idx_family_role` (`family_id`, `role_type`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='家庭成员关系表';

-- 4. 任务库表 (Tasks)
-- 定义家庭内的任务模板
CREATE TABLE IF NOT EXISTS `tasks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '所属家庭ID',
  `title` VARCHAR(128) NOT NULL COMMENT '任务标题',
  `points` INT NOT NULL DEFAULT 1 COMMENT '任务奖励积分',
  `cycle_type` ENUM('ONCE', 'DAILY', 'WEEKLY') NOT NULL DEFAULT 'ONCE' COMMENT '周期类型: ONCE-一次性, DAILY-每日, WEEKLY-每周',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-生效中, 0-已归档/删除',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建人成员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_family` (`family_id`),
  KEY `idx_family_status` (`family_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务库表';

-- 5. 任务记录表 (Task Records)
-- 记录孩子完成任务的流水
CREATE TABLE IF NOT EXISTS `task_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID (冗余字段，方便查询)',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '关联任务ID',
  `member_id` BIGINT UNSIGNED NOT NULL COMMENT '执行任务的成员ID (孩子)',
  `status` ENUM('CLAIMED', 'PENDING', 'APPROVED', 'REJECTED') NOT NULL DEFAULT 'PENDING' COMMENT '状态: CLAIMED-已领取, PENDING-待审核, APPROVED-已通过, REJECTED-已驳回',
  `submit_remark` VARCHAR(255) DEFAULT NULL COMMENT '提交说明',
  `submit_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
  `audit_time` DATETIME DEFAULT NULL COMMENT '审核时间',
  `audit_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '审核人成员ID',
  `audit_remark` VARCHAR(255) DEFAULT NULL COMMENT '审核备注/驳回理由',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_family_member` (`family_id`, `member_id`),
  KEY `idx_status` (`status`),
  KEY `idx_task_member_status` (`task_id`, `member_id`, `status`),
  KEY `idx_family_status` (`family_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行记录表';

-- 5.1. 循环任务领取关系表 (Task Claims)
-- 记录孩子对每日/每周循环任务的持续领取状态
CREATE TABLE IF NOT EXISTS `task_claims` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '领取关系ID',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
  `member_id` BIGINT UNSIGNED NOT NULL COMMENT '孩子成员ID',
  `status` ENUM('ACTIVE', 'STOPPED') NOT NULL DEFAULT 'ACTIVE' COMMENT '领取状态: ACTIVE-持续领取中, STOPPED-已停止领取',
  `claimed_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
  `stopped_at` DATETIME DEFAULT NULL COMMENT '停止领取时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_claim_member` (`task_id`, `member_id`),
  KEY `idx_claim_family_member_status` (`family_id`, `member_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='循环任务领取关系表';

-- 6. 奖品库表 (Rewards)
-- 定义家庭内的奖品
CREATE TABLE IF NOT EXISTS `rewards` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '所属家庭ID',
  `name` VARCHAR(128) NOT NULL COMMENT '奖品名称',
  `points_cost` INT NOT NULL DEFAULT 0 COMMENT '兑换消耗积分',
  `stock` INT NOT NULL DEFAULT -1 COMMENT '库存数量: -1代表无限, 0代表无货',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-上架, 0-下架',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建人成员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_family` (`family_id`),
  KEY `idx_family_status` (`family_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='奖品库表';

-- 7. 奖品兑换记录表 (Reward Records)
-- 记录孩子兑换奖品的流水
CREATE TABLE IF NOT EXISTS `reward_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID',
  `reward_id` BIGINT UNSIGNED NOT NULL COMMENT '关联奖品ID',
  `member_id` BIGINT UNSIGNED NOT NULL COMMENT '兑换成员ID (孩子)',
  `points_cost` INT NOT NULL COMMENT '消耗积分快照 (防止奖品价格变动影响历史)',
  `status` ENUM('APPLIED', 'DELIVERED', 'RECEIVED', 'REJECTED') NOT NULL DEFAULT 'APPLIED' COMMENT '状态: APPLIED-申请中, DELIVERED-已发放, RECEIVED-已确认, REJECTED-已拒绝/回退',
  `apply_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
  `operate_time` DATETIME DEFAULT NULL COMMENT '家长发放/拒绝时间',
  `operate_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '操作家长成员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_family_member` (`family_id`, `member_id`),
  KEY `idx_status` (`status`),
  KEY `idx_reward_member_status` (`reward_id`, `member_id`, `status`),
  KEY `idx_family_status` (`family_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='奖品兑换记录表';

-- 8. 积分流水表 (Point Logs)
-- 记录每一次积分变动，用于对账和展示明细
CREATE TABLE IF NOT EXISTS `point_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID',
  `member_id` BIGINT UNSIGNED NOT NULL COMMENT '积分变动对象ID (孩子)',
  `points` INT NOT NULL COMMENT '变动积分 (正数为加, 负数为减)',
  `source_type` ENUM('TASK', 'REWARD', 'ADJUST') NOT NULL COMMENT '来源类型: TASK-任务, REWARD-兑换, ADJUST-人工调整',
  `source_id` BIGINT UNSIGNED NOT NULL COMMENT '关联源ID (task_records.id 或 reward_records.id)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
  PRIMARY KEY (`id`),
  KEY `idx_member` (`member_id`),
  KEY `idx_family_member` (`family_id`, `member_id`),
  KEY `idx_source` (`source_type`, `source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分流水表';

CREATE TABLE IF NOT EXISTS `family_invites` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID',
  `inviter_member_id` BIGINT UNSIGNED NOT NULL COMMENT '邀请人成员ID',
  `target_role` ENUM('ADMIN', 'PARENT', 'CHILD') NOT NULL COMMENT '受邀加入角色',
  `token` VARCHAR(128) NOT NULL COMMENT '邀请令牌',
  `status` ENUM('ACTIVE', 'ACCEPTED', 'EXPIRED', 'CANCELED') NOT NULL DEFAULT 'ACTIVE' COMMENT '邀请状态',
  `expires_at` DATETIME NOT NULL COMMENT '过期时间',
  `accepted_by_user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '接受邀请的用户ID',
  `accepted_at` DATETIME DEFAULT NULL COMMENT '接受时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_family_status` (`family_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='家庭邀请表';

CREATE TABLE IF NOT EXISTS `family_child_bind_invites` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID',
  `child_member_id` BIGINT UNSIGNED NOT NULL COMMENT '待绑定的虚拟孩子成员ID',
  `inviter_member_id` BIGINT UNSIGNED NOT NULL COMMENT '邀请人成员ID',
  `token` VARCHAR(128) NOT NULL COMMENT '绑定邀请令牌',
  `status` ENUM('ACTIVE', 'ACCEPTED', 'EXPIRED', 'CANCELED') NOT NULL DEFAULT 'ACTIVE' COMMENT '邀请状态',
  `expires_at` DATETIME NOT NULL COMMENT '过期时间',
  `accepted_by_user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '接受绑定的用户ID',
  `accepted_at` DATETIME DEFAULT NULL COMMENT '接受时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_family_status` (`family_id`, `status`),
  KEY `idx_child_status` (`child_member_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='虚拟孩子绑定邀请表';

SET FOREIGN_KEY_CHECKS = 1;
