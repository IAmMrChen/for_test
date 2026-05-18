package srv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestIntegrationFamilyMemberInviteFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed
	invitedUserId := seed + 1

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("integration-family-%d", seed),
		Nickname: "owner",
	})
	familyId := family.Family.Id
	t.Cleanup(func() {
		cleanupIntegrationFamily(t, familyId)
	})

	if family.Family.CreatorId != ownerUserId {
		t.Fatalf("creatorId = %d, want %d", family.Family.CreatorId, ownerUserId)
	}
	if family.Member.RoleType != model.FamilyRoleOwner {
		t.Fatalf("owner role = %s, want %s", family.Member.RoleType, model.FamilyRoleOwner)
	}

	list := FamilyService.ListFamilies(ownerUserId)
	if !familyListContains(list, familyId, model.FamilyRoleOwner) {
		t.Fatalf("owner family list does not contain created family %d: %+v", familyId, list)
	}

	virtualChild := MemberService.CreateVirtualChild(ownerUserId, model.VirtualChildCreateRequest{
		FamilyId: familyId,
		Nickname: "virtual-child",
	})
	if virtualChild.UserId != nil || !virtualChild.IsVirtual || virtualChild.RoleType != model.FamilyRoleChild {
		t.Fatalf("virtual child = %+v, want nil userId, virtual child role", virtualChild)
	}

	invite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleParent,
	})
	if invite.Token == "" || invite.TargetRole != model.FamilyRoleParent {
		t.Fatalf("invite = %+v, want parent invite with token", invite)
	}

	accepted := InviteService.AcceptInvite(invitedUserId, invite.Token)
	if accepted.UserId == nil || *accepted.UserId != invitedUserId {
		t.Fatalf("accepted userId = %v, want %d", accepted.UserId, invitedUserId)
	}
	if accepted.RoleType != model.FamilyRoleParent || accepted.IsVirtual {
		t.Fatalf("accepted member = %+v, want real parent member", accepted)
	}

	invitedList := FamilyService.ListFamilies(invitedUserId)
	if !familyListContains(invitedList, familyId, model.FamilyRoleParent) {
		t.Fatalf("invited family list does not contain accepted family %d: %+v", familyId, invitedList)
	}

	secondInvite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleChild,
	})
	mustPanicWith(t, "user already has a family identity", func() {
		InviteService.AcceptInvite(invitedUserId, secondInvite.Token)
	})
}

func cleanupIntegrationFamily(t *testing.T, familyId int64) {
	if familyId == 0 || resx.Db == nil || resx.Db.Main == nil {
		return
	}

	resx.Db.Main.MustExecute("DELETE FROM point_logs WHERE family_id=@p1", familyId)
	resx.Db.Main.MustExecute("DELETE FROM reward_records WHERE family_id=@p1", familyId)
	resx.Db.Main.MustExecute("DELETE FROM rewards WHERE family_id=@p1", familyId)
	resx.Db.Main.MustExecute("DELETE FROM task_records WHERE family_id=@p1", familyId)
	resx.Db.Main.MustExecute("DELETE FROM tasks WHERE family_id=@p1", familyId)
	resx.Db.Main.MustExecute("DELETE FROM family_invites WHERE family_id=@p1", familyId)
	resx.Db.Main.MustExecute("DELETE FROM family_members WHERE family_id=@p1", familyId)
	resx.Db.Main.MustExecute("DELETE FROM families WHERE id=@p1", familyId)
}

func ensureIntegrationSchema(t *testing.T) {
	t.Helper()

	if !integrationColumnExists("family_members", "status") {
		resx.Db.Main.MustExecute(`
			ALTER TABLE family_members
			ADD COLUMN status ENUM('ACTIVE', 'REMOVED') NOT NULL DEFAULT 'ACTIVE'
		`)
	}

	if !integrationIndexExists("family_members", "uk_family_user") {
		resx.Db.Main.MustExecute("CREATE UNIQUE INDEX uk_family_user ON family_members(family_id, user_id)")
	}
	if !integrationIndexExists("family_members", "idx_family_role") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_family_role ON family_members(family_id, role_type)")
	}
	if !integrationIndexExists("family_members", "idx_user") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_user ON family_members(user_id)")
	}

	if !integrationTableExists("family_invites") {
		resx.Db.Main.MustExecute(`
			CREATE TABLE family_invites (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				family_id BIGINT UNSIGNED NOT NULL,
				inviter_member_id BIGINT UNSIGNED NOT NULL,
				target_role ENUM('ADMIN', 'PARENT', 'CHILD') NOT NULL,
				token VARCHAR(128) NOT NULL,
				status ENUM('ACTIVE', 'ACCEPTED', 'EXPIRED', 'CANCELED') NOT NULL DEFAULT 'ACTIVE',
				expires_at DATETIME NOT NULL,
				accepted_by_user_id BIGINT UNSIGNED DEFAULT NULL,
				accepted_at DATETIME DEFAULT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				UNIQUE KEY uk_token (token),
				KEY idx_family_status (family_id, status)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`)
	}

	if !integrationTableExists("tasks") {
		resx.Db.Main.MustExecute(`
			CREATE TABLE tasks (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				family_id BIGINT UNSIGNED NOT NULL,
				title VARCHAR(128) NOT NULL,
				points INT NOT NULL DEFAULT 1,
				cycle_type ENUM('ONCE', 'DAILY', 'WEEKLY') NOT NULL DEFAULT 'ONCE',
				status TINYINT NOT NULL DEFAULT 1,
				created_by BIGINT UNSIGNED NOT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				KEY idx_family (family_id)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`)
	}
	if !integrationIndexExists("tasks", "idx_family_status") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_family_status ON tasks(family_id, status)")
	}

	if !integrationTableExists("task_records") {
		resx.Db.Main.MustExecute(`
			CREATE TABLE task_records (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				family_id BIGINT UNSIGNED NOT NULL,
				task_id BIGINT UNSIGNED NOT NULL,
				member_id BIGINT UNSIGNED NOT NULL,
				status ENUM('CLAIMED', 'PENDING', 'APPROVED', 'REJECTED') NOT NULL DEFAULT 'PENDING',
				submit_remark VARCHAR(255) DEFAULT NULL,
				submit_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				audit_time DATETIME DEFAULT NULL,
				audit_by BIGINT UNSIGNED DEFAULT NULL,
				audit_remark VARCHAR(255) DEFAULT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				KEY idx_family_member (family_id, member_id),
				KEY idx_status (status)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`)
	}
	if !integrationEnumColumnContains("task_records", "status", "CLAIMED") {
		resx.Db.Main.MustExecute(`
			ALTER TABLE task_records
			MODIFY COLUMN status ENUM('CLAIMED', 'PENDING', 'APPROVED', 'REJECTED') NOT NULL DEFAULT 'PENDING'
		`)
	}
	if !integrationColumnExists("task_records", "submit_remark") {
		resx.Db.Main.MustExecute("ALTER TABLE task_records ADD COLUMN submit_remark VARCHAR(255) DEFAULT NULL")
	}
	if !integrationIndexExists("task_records", "idx_task_member_status") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_task_member_status ON task_records(task_id, member_id, status)")
	}
	if !integrationIndexExists("task_records", "idx_family_status") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_family_status ON task_records(family_id, status)")
	}

	if !integrationTableExists("point_logs") {
		resx.Db.Main.MustExecute(`
			CREATE TABLE point_logs (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				family_id BIGINT UNSIGNED NOT NULL,
				member_id BIGINT UNSIGNED NOT NULL,
				points INT NOT NULL,
				source_type ENUM('TASK', 'REWARD', 'ADJUST') NOT NULL,
				source_id BIGINT UNSIGNED NOT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				KEY idx_member (member_id)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`)
	}
	if !integrationIndexExists("point_logs", "idx_family_member") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_family_member ON point_logs(family_id, member_id)")
	}
	if !integrationIndexExists("point_logs", "idx_source") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_source ON point_logs(source_type, source_id)")
	}

	if !integrationTableExists("rewards") {
		resx.Db.Main.MustExecute(`
			CREATE TABLE rewards (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				family_id BIGINT UNSIGNED NOT NULL,
				name VARCHAR(128) NOT NULL,
				points_cost INT NOT NULL DEFAULT 0,
				stock INT NOT NULL DEFAULT -1,
				status TINYINT NOT NULL DEFAULT 1,
				created_by BIGINT UNSIGNED NOT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				KEY idx_family (family_id)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`)
	}
	if !integrationIndexExists("rewards", "idx_family_status") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_family_status ON rewards(family_id, status)")
	}

	if !integrationTableExists("reward_records") {
		resx.Db.Main.MustExecute(`
			CREATE TABLE reward_records (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				family_id BIGINT UNSIGNED NOT NULL,
				reward_id BIGINT UNSIGNED NOT NULL,
				member_id BIGINT UNSIGNED NOT NULL,
				points_cost INT NOT NULL,
				status ENUM('APPLIED', 'DELIVERED', 'RECEIVED', 'REJECTED') NOT NULL DEFAULT 'APPLIED',
				apply_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				operate_time DATETIME DEFAULT NULL,
				operate_by BIGINT UNSIGNED DEFAULT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				KEY idx_family_member (family_id, member_id),
				KEY idx_status (status)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`)
	}
	if !integrationIndexExists("reward_records", "idx_reward_member_status") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_reward_member_status ON reward_records(reward_id, member_id, status)")
	}
	if !integrationIndexExists("reward_records", "idx_family_status") {
		resx.Db.Main.MustExecute("CREATE INDEX idx_family_status ON reward_records(family_id, status)")
	}
}

func integrationTableExists(table string) bool {
	return integrationCount(`
		SELECT COUNT(*)
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=@p1
	`, table) > 0
}

func integrationColumnExists(table string, column string) bool {
	return integrationCount(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=@p1 AND COLUMN_NAME=@p2
	`, table, column) > 0
}

func integrationIndexExists(table string, index string) bool {
	return integrationCount(`
		SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=@p1 AND INDEX_NAME=@p2
	`, table, index) > 0
}

func integrationEnumColumnContains(table string, column string, value string) bool {
	return integrationCount(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA=DATABASE()
			AND TABLE_NAME=@p1
			AND COLUMN_NAME=@p2
			AND COLUMN_TYPE LIKE @p3
	`, table, column, "%'"+value+"'%") > 0
}

func integrationCount(sql string, args ...any) int {
	value, ok := resx.Db.Main.MustScalarInt(sql, args...)
	if !ok || value == nil {
		return 0
	}
	return *value
}

func familyListContains(list []model.FamilyListItem, familyId int64, role model.FamilyRole) bool {
	for _, item := range list {
		if item.FamilyId == familyId && item.RoleType == role {
			return true
		}
	}
	return false
}

func mustPanicWith(t *testing.T, want string, fn func()) {
	t.Helper()

	defer func() {
		v := recover()
		if v == nil {
			t.Fatalf("expected panic %q", want)
		}
		if got := fmt.Sprint(v); got != want {
			t.Fatalf("panic = %q, want %q", got, want)
		}
	}()

	fn()
}
