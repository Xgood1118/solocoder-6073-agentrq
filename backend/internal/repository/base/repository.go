package base

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	entity "github.com/agentrq/agentrq/backend/internal/data/entity/crud"
	"github.com/agentrq/agentrq/backend/internal/data/model"
	"github.com/agentrq/agentrq/backend/internal/repository/dbconn"
	"gorm.io/gorm"
)

var customFieldKeyRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func sanitizeCustomFieldKey(key string) string {
	if customFieldKeyRe.MatchString(key) {
		return key
	}
	return ""
}

var ErrNotFound = errors.New("not found")

type Repository interface {
	// Workspace
	CreateWorkspace(ctx context.Context, p model.Workspace) (model.Workspace, error)
	GetWorkspace(ctx context.Context, id int64, userID int64) (model.Workspace, error)
	CheckWorkspaceAccess(ctx context.Context, id int64, userID int64) (bool, error)
	ListWorkspaces(ctx context.Context, userID int64, includeArchived bool) ([]model.Workspace, error)
	DeleteWorkspace(ctx context.Context, id int64, userID int64) error
	UpdateWorkspace(ctx context.Context, p model.Workspace) (model.Workspace, error)

	// Task
	CreateTask(ctx context.Context, t model.Task) (model.Task, error)
	GetTask(ctx context.Context, workspaceID, taskID int64, userID int64) (model.Task, error)
	ListTasks(ctx context.Context, req entity.ListTasksRequest, userID int64) ([]model.Task, error)
	UpdateTask(ctx context.Context, t model.Task) (model.Task, error)
	DeleteTask(ctx context.Context, workspaceID, taskID int64, userID int64) error

	// Message
	CreateMessage(ctx context.Context, m model.Message) error
	ListMessages(ctx context.Context, taskID int64) ([]model.Message, error)
	UpdateMessageMetadata(ctx context.Context, taskID int64, messageID int64, metadata []byte) error
	GetWorkspaceAttachmentIDs(ctx context.Context, workspaceID int64) ([]string, error)

	SystemGetWorkspace(ctx context.Context, id int64) (model.Workspace, error)
	SystemGetTask(ctx context.Context, id int64) (model.Task, error)
	SystemGetMessage(ctx context.Context, id int64) (model.Message, error)
	SystemGetUser(ctx context.Context, id int64) (model.User, error)
	SystemListTasksByStatus(ctx context.Context, status string) ([]model.Task, error)
	SystemCheckTaskExists(ctx context.Context, workspaceID, parentID int64, status string) (bool, error)
	GetDetailedWorkspaceStats(ctx context.Context, workspaceID int64, startTime, endTime int64) (entity.GetDetailedWorkspaceStatsResponse, error)
	GetWorkspaceTaskCounts(ctx context.Context, workspaceID int64) (int64, int64, error)
	GetWorkspaceTaskCountsByCategory(ctx context.Context, workspaceID int64, userID int64) (map[string]int64, error)
	GetTelemetryActionCounts(ctx context.Context) (map[uint8]int64, error)
	FindUserByEmail(ctx context.Context, email string) (model.User, error)
	CreateUser(ctx context.Context, u model.User) (model.User, error)
	UpdateUser(ctx context.Context, u model.User) (model.User, error)
	GetNextTask(ctx context.Context, workspaceID int64, userID int64) (model.Task, error)
	GetGlobalTaskStats(ctx context.Context, userID int64) (entity.GlobalTaskStatsResponse, error)

	// Slack integration
	UpsertSlackWorkspaceLink(ctx context.Context, link model.SlackWorkspaceLink) error
	GetSlackWorkspaceLink(ctx context.Context, workspaceID int64) (model.SlackWorkspaceLink, error)
	GetSlackWorkspaceLinkByChannel(ctx context.Context, channelID string) (model.SlackWorkspaceLink, error)
	DeleteSlackWorkspaceLink(ctx context.Context, workspaceID int64) error
	UpsertSlackTaskThread(ctx context.Context, thread model.SlackTaskThread) error
	GetSlackTaskThreadByTask(ctx context.Context, taskID int64) (model.SlackTaskThread, error)
	GetSlackTaskThreadByChannel(ctx context.Context, channelID, threadTS string) (model.SlackTaskThread, error)

	// Push subscriptions
	SavePushSubscription(ctx context.Context, sub model.PushSubscription) error
	DeletePushSubscription(ctx context.Context, userID int64, endpoint string) error
	DeletePushSubscriptionByWorkspace(ctx context.Context, userID int64, workspaceID int64, endpoint string) error
	GetPushSubscriptionForWorkspace(ctx context.Context, userID int64, workspaceID int64, endpoint string) (bool, error)
	ListPushSubscriptionsByUserAndWorkspace(ctx context.Context, userID int64, workspaceID int64) ([]model.PushSubscription, error)

	// WorkspaceTemplate
	CreateWorkspaceTemplate(ctx context.Context, t model.WorkspaceTemplate) (model.WorkspaceTemplate, error)
	GetWorkspaceTemplate(ctx context.Context, id int64, userID int64) (model.WorkspaceTemplate, error)
	ListWorkspaceTemplates(ctx context.Context, userID int64) ([]model.WorkspaceTemplate, error)
	UpdateWorkspaceTemplate(ctx context.Context, t model.WorkspaceTemplate) (model.WorkspaceTemplate, error)
	DeleteWorkspaceTemplate(ctx context.Context, id int64, userID int64) error

	// CustomField
	CreateCustomField(ctx context.Context, f model.CustomField) (model.CustomField, error)
	GetCustomField(ctx context.Context, id int64, workspaceID int64) (model.CustomField, error)
	ListCustomFields(ctx context.Context, workspaceID int64) ([]model.CustomField, error)
	UpdateCustomField(ctx context.Context, f model.CustomField) (model.CustomField, error)
	DeleteCustomField(ctx context.Context, id int64, workspaceID int64) error
	ClearCustomFieldFromTasks(ctx context.Context, workspaceID int64, fieldKey string) error
}

type repository struct {
	db dbconn.DBConn
}

func New(db dbconn.DBConn) Repository {
	return &repository{db: db}
}

func (r *repository) conn(ctx context.Context) *gorm.DB {
	return r.db.Conn(ctx).WithContext(ctx)
}

// ── Workspaces ──────────────────────────────────────────────────────────────────

func (r *repository) CreateWorkspace(ctx context.Context, p model.Workspace) (model.Workspace, error) {
	if err := r.conn(ctx).Create(&p).Error; err != nil {
		return model.Workspace{}, err
	}
	return p, nil
}

func (r *repository) GetWorkspace(ctx context.Context, id int64, userID int64) (model.Workspace, error) {
	var p model.Workspace
	err := r.conn(ctx).Where("id = ? AND user_id = ?", id, userID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Workspace{}, ErrNotFound
	}
	return p, err
}

func (r *repository) CheckWorkspaceAccess(ctx context.Context, id int64, userID int64) (bool, error) {
	var count int64
	err := r.conn(ctx).Model(&model.Workspace{}).Where("id = ? AND user_id = ?", id, userID).Count(&count).Error
	return count > 0, err
}

func (r *repository) ListWorkspaces(ctx context.Context, userID int64, includeArchived bool) ([]model.Workspace, error) {
	var workspaces []model.Workspace
	query := r.conn(ctx).Where("user_id = ?", userID)
	if !includeArchived {
		query = query.Where("archived_at IS NULL")
	}
	err := query.Order("created_at desc").Find(&workspaces).Error
	return workspaces, err
}

func (r *repository) UpdateWorkspace(ctx context.Context, p model.Workspace) (model.Workspace, error) {
	if err := r.conn(ctx).Save(&p).Error; err != nil {
		return model.Workspace{}, err
	}
	return p, nil
}

func (r *repository) DeleteWorkspace(ctx context.Context, id int64, userID int64) error {
	return r.conn(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete all messages for all tasks in this workspace
		if err := tx.Where("task_id IN (?)", tx.Model(&model.Task{}).Select("id").Where("workspace_id = ?", id)).Delete(&model.Message{}).Error; err != nil {
			return err
		}

		// 2. Delete all tasks in this workspace
		if err := tx.Where("workspace_id = ?", id).Delete(&model.Task{}).Error; err != nil {
			return err
		}

		// 3. Delete the workspace itself
		res := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Workspace{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ── Tasks ─────────────────────────────────────────────────────────────────────

func (r *repository) CreateTask(ctx context.Context, t model.Task) (model.Task, error) {
	if err := r.conn(ctx).Create(&t).Error; err != nil {
		return model.Task{}, err
	}
	return t, nil
}

func (r *repository) GetTask(ctx context.Context, workspaceID, taskID int64, userID int64) (model.Task, error) {
	var t model.Task
	err := r.conn(ctx).
		Preload("Messages").
		Where("id = ? AND workspace_id = ? AND user_id = ?", taskID, workspaceID, userID).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Task{}, ErrNotFound
	}
	return t, err
}

func (r *repository) ListTasks(ctx context.Context, req entity.ListTasksRequest, userID int64) ([]model.Task, error) {
	var tasks []model.Task
	q := r.conn(ctx).Where("user_id = ?", userID)

	if req.WorkspaceID != 0 {
		q = q.Where("workspace_id = ?", req.WorkspaceID)
	}
	if req.CreatedBy != "" {
		q = q.Where("created_by = ?", req.CreatedBy)
	}
	if len(req.Status) > 0 {
		q = q.Where("status IN ?", req.Status)
	}

	if req.Filter == "pending_approval" {
		// Find tasks whose most recent message is a permission_request.
		// PostgreSQL: JSONB columns don't support LIKE; cast to text or use @> containment.
		// SQLite: metadata is plain text, LIKE works fine.
		var metadataExpr string
		if r.conn(ctx).Dialector.Name() == "postgres" {
			metadataExpr = "metadata @> '{\"type\":\"permission_request\"}'::jsonb"
		} else {
			metadataExpr = "metadata LIKE '%\"type\":\"permission_request\"%'"
		}
		q = q.Where("id IN (SELECT task_id FROM messages m1 WHERE created_at = (SELECT MAX(created_at) FROM messages m2 WHERE m2.task_id = m1.task_id) AND " + metadataExpr + ")")
	}

	if len(req.CustomFieldFilter) > 0 {
		dialect := r.conn(ctx).Dialector.Name()
		for key, val := range req.CustomFieldFilter {
			cleanKey := sanitizeCustomFieldKey(key)
			if cleanKey == "" {
				continue
			}
			strVal := fmt.Sprintf("%v", val)
			if dialect == "sqlite" {
				q = q.Where(fmt.Sprintf("json_extract(custom_fields, '$.%s') = ?", cleanKey), strVal)
			} else {
				q = q.Where(fmt.Sprintf("custom_fields::jsonb->>'%s' = ?", cleanKey), strVal)
			}
		}
	}

	orderBy := "created_at desc"
	if req.CustomFieldSort != "" {
		cleanKey := sanitizeCustomFieldKey(req.CustomFieldSort)
		if cleanKey != "" {
			dialect := r.conn(ctx).Dialector.Name()
			dir := "ASC"
			if strings.EqualFold(req.CustomFieldSortDir, "desc") {
				dir = "DESC"
			}
			if dialect == "sqlite" {
				orderBy = fmt.Sprintf("json_extract(custom_fields, '$.%s') %s", cleanKey, dir)
			} else {
				orderBy = fmt.Sprintf("(custom_fields::jsonb->>'%s') %s", cleanKey, dir)
			}
		}
	} else if req.Filter == "pending_approval" {
		orderBy = "created_at asc"
	} else if len(req.Status) > 1 {
		// Mixed statuses, likely "active" view (ongoing, blocked, notstarted, cron)
		// We prioritize status: ongoing (0) > blocked (1) > cron (2) > notstarted (3)
		orderBy = "CASE WHEN status = 'ongoing' THEN 0 WHEN status = 'blocked' THEN 1 WHEN status = 'cron' THEN 2 ELSE 3 END, updated_at DESC"
	} else if len(req.Status) == 1 {
		status := req.Status[0]
		if status == "notstarted" {
			dialect := r.conn(ctx).Dialector.Name()
			var sortExpr string
			if dialect == "sqlite" {
				sortExpr = "(CASE WHEN sort_order > 0 THEN sort_order ELSE CAST(strftime('%s', created_at) AS REAL) END)"
			} else {
				sortExpr = "(CASE WHEN sort_order > 0 THEN sort_order ELSE EXTRACT(EPOCH FROM created_at) END)"
			}
			orderBy = fmt.Sprintf("%s ASC, id ASC", sortExpr)
		} else if status != "cron" {
			orderBy = "updated_at desc"
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100 // Enforce default safety limit to prevent OOM
	}
	if limit > 500 {
		limit = 500 // Hard cap to prevent PostgreSQL / backend overload
	}
	q = q.Limit(limit)

	if req.Offset > 0 {
		q = q.Offset(req.Offset)
	}

	err := q.Order(orderBy).Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	if req.PreloadMessages && len(tasks) > 0 {
		var metadataExpr string
		if r.conn(ctx).Dialector.Name() == "postgres" {
			metadataExpr = "metadata @> '{\"type\":\"permission_request\",\"status\":\"pending\"}'::jsonb"
		} else {
			metadataExpr = "metadata LIKE '%\"type\":\"permission_request\"%' AND metadata LIKE '%\"status\":\"pending\"%'"
		}

		taskIDs := make([]int64, len(tasks))
		taskMap := make(map[int64]*model.Task, len(tasks))
		for i := range tasks {
			taskIDs[i] = tasks[i].ID
			taskMap[tasks[i].ID] = &tasks[i]
		}

		const chunkSize = 500
		for i := 0; i < len(taskIDs); i += chunkSize {
			end := i + chunkSize
			if end > len(taskIDs) {
				end = len(taskIDs)
			}
			batch := taskIDs[i:end]

			var batchMessages []model.Message
			err := r.conn(ctx).
				Where("task_id IN ?", batch).
				Where("id = (SELECT MAX(id) FROM messages m2 WHERE m2.task_id = messages.task_id) OR (" + metadataExpr + ")").
				Order("created_at asc").
				Find(&batchMessages).Error
			if err != nil {
				return nil, err
			}

			for _, msg := range batchMessages {
				if t, ok := taskMap[msg.TaskID]; ok {
					t.Messages = append(t.Messages, msg)
				}
			}
		}
	}

	return tasks, nil
}

func (r *repository) GetNextTask(ctx context.Context, workspaceID int64, userID int64) (model.Task, error) {
	var t model.Task
	dialect := r.conn(ctx).Dialector.Name()
	var sortExpr string
	if dialect == "sqlite" {
		sortExpr = "(CASE WHEN sort_order > 0 THEN sort_order ELSE CAST(strftime('%s', created_at) AS REAL) END)"
	} else {
		// Assume Postgres
		sortExpr = "(CASE WHEN sort_order > 0 THEN sort_order ELSE EXTRACT(EPOCH FROM created_at) END)"
	}

	err := r.conn(ctx).
		Where("workspace_id = ? AND user_id = ? AND status = ? AND assignee = ?", workspaceID, userID, "notstarted", "agent").
		Order(fmt.Sprintf("%s ASC, id ASC", sortExpr)).
		First(&t).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Task{}, ErrNotFound
	}
	return t, err
}

func (r *repository) UpdateTask(ctx context.Context, t model.Task) (model.Task, error) {
	if err := r.conn(ctx).Save(&t).Error; err != nil {
		return model.Task{}, err
	}
	return t, nil
}

func (r *repository) DeleteTask(ctx context.Context, workspaceID, taskID int64, userID int64) error {
	return r.conn(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete all messages for this task
		if err := tx.Where("task_id = ?", taskID).Delete(&model.Message{}).Error; err != nil {
			return err
		}

		// 2. Delete the task
		res := tx.Where("id = ? AND workspace_id = ? AND user_id = ?", taskID, workspaceID, userID).
			Delete(&model.Task{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (r *repository) CreateMessage(ctx context.Context, m model.Message) error {
	return r.conn(ctx).Create(&m).Error
}

func (r *repository) ListMessages(ctx context.Context, taskID int64) ([]model.Message, error) {
	var msgs []model.Message
	err := r.conn(ctx).Where("task_id = ?", taskID).Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

func (r *repository) UpdateMessageMetadata(ctx context.Context, taskID int64, messageID int64, metadata []byte) error {
	return r.conn(ctx).Model(&model.Message{}).Where("id = ? AND task_id = ?", messageID, taskID).Update("metadata", metadata).Error
}

func (r *repository) GetWorkspaceAttachmentIDs(ctx context.Context, workspaceID int64) ([]string, error) {
	var attachmentIDs []string

	// 1. Get attachments from tasks
	var taskAttachments []string
	err := r.conn(ctx).Model(&model.Task{}).Where("workspace_id = ?", workspaceID).Pluck("attachments", &taskAttachments).Error
	if err == nil {
		for _, ta := range taskAttachments {
			if len(ta) > 0 {
				var atts []entity.Attachment
				if err := json.Unmarshal([]byte(ta), &atts); err == nil {
					for _, a := range atts {
						if a.ID != "" {
							attachmentIDs = append(attachmentIDs, a.ID)
						}
					}
				}
			}
		}
	}

	// 2. Get attachments from messages
	var msgAttachments []string
	err = r.conn(ctx).Model(&model.Message{}).
		Joins("JOIN tasks ON tasks.id = messages.task_id").
		Where("tasks.workspace_id = ?", workspaceID).
		Pluck("messages.attachments", &msgAttachments).Error
	if err == nil {
		for _, ma := range msgAttachments {
			if len(ma) > 0 {
				var atts []entity.Attachment
				if err := json.Unmarshal([]byte(ma), &atts); err == nil {
					for _, a := range atts {
						if a.ID != "" {
							attachmentIDs = append(attachmentIDs, a.ID)
						}
					}
				}
			}
		}
	}

	return attachmentIDs, nil
}

func (r *repository) SystemGetWorkspace(ctx context.Context, id int64) (model.Workspace, error) {
	var p model.Workspace
	err := r.conn(ctx).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Workspace{}, ErrNotFound
	}
	return p, err
}

func (r *repository) SystemGetTask(ctx context.Context, id int64) (model.Task, error) {
	var t model.Task
	err := r.conn(ctx).First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Task{}, ErrNotFound
	}
	return t, err
}

func (r *repository) SystemGetMessage(ctx context.Context, id int64) (model.Message, error) {
	var m model.Message
	err := r.conn(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Message{}, ErrNotFound
	}
	return m, err
}

func (r *repository) SystemGetUser(ctx context.Context, id int64) (model.User, error) {
	var u model.User
	err := r.conn(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

func (r *repository) SystemListTasksByStatus(ctx context.Context, status string) ([]model.Task, error) {
	var tasks []model.Task
	err := r.conn(ctx).Where("status = ?", status).Find(&tasks).Error
	return tasks, err
}

func (r *repository) SystemCheckTaskExists(ctx context.Context, workspaceID, parentID int64, status string) (bool, error) {
	var count int64
	err := r.conn(ctx).Model(&model.Task{}).
		Where("workspace_id = ? AND parent_id = ? AND status = ?", workspaceID, parentID, status).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) GetDetailedWorkspaceStats(ctx context.Context, workspaceID int64, startTime, endTime int64) (entity.GetDetailedWorkspaceStatsResponse, error) {
	var res entity.GetDetailedWorkspaceStatsResponse

	// Dialect specific date formatting
	dialect := r.conn(ctx).Dialector.Name()
	var dateExpr string
	if dialect == "sqlite" {
		dateExpr = "strftime('%Y-%m-%d', datetime(occurred_at, 'unixepoch', 'localtime'))"
	} else {
		// Assume Postgres
		dateExpr = "TO_CHAR(TO_TIMESTAMP(occurred_at) AT TIME ZONE 'UTC', 'YYYY-MM-DD')"
	}

	// 1. Get Summary Stats
	type countResult struct {
		Action uint8
		Count  int64
	}
	var summaryResults []countResult
	err := r.conn(ctx).Model(&model.Telemetry{}).
		Select("action, count(*) as count").
		Where("workspace_id = ? AND occurred_at >= ? AND occurred_at <= ?", workspaceID, startTime, endTime).
		Group("action").
		Scan(&summaryResults).Error
	if err != nil {
		return res, err
	}

	for _, row := range summaryResults {
		switch row.Action {
		case model.ActionIDTaskComplete:
			res.Summary.TasksCompleted = row.Count
		case model.ActionIDTaskFromScheduled:
			res.Summary.TasksScheduled = row.Count
		case model.ActionIDMessageCreate:
			res.Summary.Messages = row.Count
		case model.ActionIDTaskApproveManual, model.ActionIDMCPPermissionManual:
			res.Summary.ManualApprovals += row.Count
		case model.ActionIDMCPPermissionAuto:
			res.Summary.AutoApprovals += row.Count
		case model.ActionIDTaskRejectManual, model.ActionIDMCPPermissionDeny:
			res.Summary.Denies += row.Count
		}
	}

	// 2. Get Timeseries for Tasks Completed
	err = r.conn(ctx).Model(&model.Telemetry{}).
		Select(dateExpr+" as date, count(*) as count").
		Where("workspace_id = ? AND occurred_at >= ? AND occurred_at <= ? AND action = ?", workspaceID, startTime, endTime, model.ActionIDTaskComplete).
		Group("date").
		Order("date ASC").
		Scan(&res.Timeseries.TasksCompleted).Error
	if err != nil {
		return res, err
	}

	// 3. Get Timeseries for Messages
	err = r.conn(ctx).Model(&model.Telemetry{}).
		Select(dateExpr+" as date, count(*) as count").
		Where("workspace_id = ? AND occurred_at >= ? AND occurred_at <= ? AND action = ?", workspaceID, startTime, endTime, model.ActionIDMessageCreate).
		Group("date").
		Order("date ASC").
		Scan(&res.Timeseries.Messages).Error

	return res, err
}

func (r *repository) GetWorkspaceTaskCounts(ctx context.Context, workspaceID int64) (int64, int64, error) {
	var total, active int64
	err := r.conn(ctx).Model(&model.Task{}).
		Where("workspace_id = ?", workspaceID).
		Count(&total).Error
	if err != nil {
		return 0, 0, err
	}

	err = r.conn(ctx).Model(&model.Task{}).
		Where("workspace_id = ? AND status NOT IN ?", workspaceID, []string{"completed", "archived"}).
		Count(&active).Error
	return active, total, err
}

func (r *repository) GetTelemetryActionCounts(ctx context.Context) (map[uint8]int64, error) {
	type countResult struct {
		Action uint8
		Count  int64
	}
	var results []countResult
	err := r.conn(ctx).Model(&model.Telemetry{}).
		Select("action, count(*) as count").
		Group("action").
		Scan(&results).Error

	m := make(map[uint8]int64)
	for _, rr := range results {
		m[rr.Action] = rr.Count
	}
	return m, err
}

// ── Users ─────────────────────────────────────────────────────────────────────

func (r *repository) FindUserByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.conn(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

func (r *repository) CreateUser(ctx context.Context, u model.User) (model.User, error) {
	if err := r.conn(ctx).Create(&u).Error; err != nil {
		return model.User{}, err
	}
	return u, nil
}

func (r *repository) UpdateUser(ctx context.Context, u model.User) (model.User, error) {
	if err := r.conn(ctx).Save(&u).Error; err != nil {
		return model.User{}, err
	}
	return u, nil
}

// ── Slack Integration ─────────────────────────────────────────────────────────

func (r *repository) UpsertSlackWorkspaceLink(ctx context.Context, link model.SlackWorkspaceLink) error {
	return r.conn(ctx).Save(&link).Error
}

func (r *repository) GetSlackWorkspaceLink(ctx context.Context, workspaceID int64) (model.SlackWorkspaceLink, error) {
	var l model.SlackWorkspaceLink
	err := r.conn(ctx).First(&l, "workspace_id = ?", workspaceID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SlackWorkspaceLink{}, ErrNotFound
	}
	return l, err
}

func (r *repository) GetSlackWorkspaceLinkByChannel(ctx context.Context, channelID string) (model.SlackWorkspaceLink, error) {
	var l model.SlackWorkspaceLink
	err := r.conn(ctx).First(&l, "slack_channel_id = ?", channelID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SlackWorkspaceLink{}, ErrNotFound
	}
	return l, err
}

func (r *repository) DeleteSlackWorkspaceLink(ctx context.Context, workspaceID int64) error {
	return r.conn(ctx).Delete(&model.SlackWorkspaceLink{}, "workspace_id = ?", workspaceID).Error
}

func (r *repository) UpsertSlackTaskThread(ctx context.Context, thread model.SlackTaskThread) error {
	return r.conn(ctx).Save(&thread).Error
}

func (r *repository) GetSlackTaskThreadByTask(ctx context.Context, taskID int64) (model.SlackTaskThread, error) {
	var t model.SlackTaskThread
	err := r.conn(ctx).First(&t, "task_id = ?", taskID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SlackTaskThread{}, ErrNotFound
	}
	return t, err
}

func (r *repository) GetSlackTaskThreadByChannel(ctx context.Context, channelID, threadTS string) (model.SlackTaskThread, error) {
	var t model.SlackTaskThread
	err := r.conn(ctx).Where("slack_channel_id = ? AND thread_ts = ?", channelID, threadTS).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SlackTaskThread{}, ErrNotFound
	}
	return t, err
}

func (r *repository) GetGlobalTaskStats(ctx context.Context, userID int64) (entity.GlobalTaskStatsResponse, error) {
	var res entity.GlobalTaskStatsResponse
	var pending, scheduled int64

	err := r.conn(ctx).Model(&model.Task{}).
		Where("user_id = ? AND status IN ?", userID, []string{"notstarted", "ongoing", "blocked"}).
		Count(&pending).Error
	if err != nil {
		return res, err
	}

	err = r.conn(ctx).Model(&model.Task{}).
		Where("user_id = ? AND status = ?", userID, "cron").
		Count(&scheduled).Error
	if err != nil {
		return res, err
	}

	res.PendingTasks = pending
	res.ScheduledTasks = scheduled
	return res, nil
}

func (r *repository) GetWorkspaceTaskCountsByCategory(ctx context.Context, workspaceID int64, userID int64) (map[string]int64, error) {
	counts := map[string]int64{
		"ongoing":    0,
		"notstarted": 0,
		"scheduled":  0,
		"completed":  0,
		"pending":    0,
	}

	type statusCount struct {
		Status string
		Count  int64
	}
	var results []statusCount
	if err := r.conn(ctx).Model(&model.Task{}).
		Select("status, count(*) as count").
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, res := range results {
		switch res.Status {
		case "ongoing", "blocked":
			counts["ongoing"] += res.Count
		case "notstarted":
			counts["notstarted"] += res.Count
		case "cron":
			counts["scheduled"] += res.Count
		case "completed", "rejected":
			counts["completed"] += res.Count
		}
	}

	// 5. Pending (Action Required)
	var pending int64
	var metadataExpr string
	if r.conn(ctx).Dialector.Name() == "postgres" {
		metadataExpr = "metadata @> '{\"type\":\"permission_request\"}'::jsonb"
	} else {
		metadataExpr = "metadata LIKE '%\"type\":\"permission_request\"%'"
	}
	err := r.conn(ctx).Model(&model.Task{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Where("id IN (SELECT task_id FROM messages m1 WHERE created_at = (SELECT MAX(created_at) FROM messages m2 WHERE m2.task_id = m1.task_id) AND " + metadataExpr + ")").
		Count(&pending).Error
	if err != nil {
		return nil, err
	}
	counts["pending"] = pending

	return counts, nil
}

// ── Push Subscriptions ──────────────────────────────────────────────────────────

func (r *repository) SavePushSubscription(ctx context.Context, sub model.PushSubscription) error {
	return r.conn(ctx).
		Where(model.PushSubscription{Endpoint: sub.Endpoint, WorkspaceID: sub.WorkspaceID}).
		Assign(sub).
		FirstOrCreate(&sub).Error
}

func (r *repository) DeletePushSubscription(ctx context.Context, userID int64, endpoint string) error {
	return r.conn(ctx).
		Where("user_id = ? AND endpoint = ?", userID, endpoint).
		Delete(&model.PushSubscription{}).Error
}

func (r *repository) DeletePushSubscriptionByWorkspace(ctx context.Context, userID int64, workspaceID int64, endpoint string) error {
	return r.conn(ctx).
		Where("user_id = ? AND workspace_id = ? AND endpoint = ?", userID, workspaceID, endpoint).
		Delete(&model.PushSubscription{}).Error
}

func (r *repository) GetPushSubscriptionForWorkspace(ctx context.Context, userID int64, workspaceID int64, endpoint string) (bool, error) {
	var count int64
	err := r.conn(ctx).Model(&model.PushSubscription{}).
		Where("user_id = ? AND workspace_id = ? AND endpoint = ?", userID, workspaceID, endpoint).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) ListPushSubscriptionsByUserAndWorkspace(ctx context.Context, userID int64, workspaceID int64) ([]model.PushSubscription, error) {
	var subs []model.PushSubscription
	err := r.conn(ctx).Where("user_id = ? AND workspace_id = ?", userID, workspaceID).Find(&subs).Error
	return subs, err
}

// ── WorkspaceTemplates ────────────────────────────────────────────────────────

func (r *repository) CreateWorkspaceTemplate(ctx context.Context, t model.WorkspaceTemplate) (model.WorkspaceTemplate, error) {
	if err := r.conn(ctx).Create(&t).Error; err != nil {
		return model.WorkspaceTemplate{}, err
	}
	return t, nil
}

func (r *repository) GetWorkspaceTemplate(ctx context.Context, id int64, userID int64) (model.WorkspaceTemplate, error) {
	var t model.WorkspaceTemplate
	err := r.conn(ctx).Where("id = ? AND user_id = ?", id, userID).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.WorkspaceTemplate{}, ErrNotFound
	}
	return t, err
}

func (r *repository) ListWorkspaceTemplates(ctx context.Context, userID int64) ([]model.WorkspaceTemplate, error) {
	var templates []model.WorkspaceTemplate
	err := r.conn(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&templates).Error
	return templates, err
}

func (r *repository) UpdateWorkspaceTemplate(ctx context.Context, t model.WorkspaceTemplate) (model.WorkspaceTemplate, error) {
	if err := r.conn(ctx).Save(&t).Error; err != nil {
		return model.WorkspaceTemplate{}, err
	}
	return t, nil
}

func (r *repository) DeleteWorkspaceTemplate(ctx context.Context, id int64, userID int64) error {
	res := r.conn(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.WorkspaceTemplate{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ── CustomFields ──────────────────────────────────────────────────────────────

func (r *repository) CreateCustomField(ctx context.Context, f model.CustomField) (model.CustomField, error) {
	if err := r.conn(ctx).Create(&f).Error; err != nil {
		return model.CustomField{}, err
	}
	return f, nil
}

func (r *repository) GetCustomField(ctx context.Context, id int64, workspaceID int64) (model.CustomField, error) {
	var f model.CustomField
	err := r.conn(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.CustomField{}, ErrNotFound
	}
	return f, err
}

func (r *repository) ListCustomFields(ctx context.Context, workspaceID int64) ([]model.CustomField, error) {
	var fields []model.CustomField
	err := r.conn(ctx).Where("workspace_id = ?", workspaceID).Order("sort_order asc, id asc").Find(&fields).Error
	return fields, err
}

func (r *repository) UpdateCustomField(ctx context.Context, f model.CustomField) (model.CustomField, error) {
	if err := r.conn(ctx).Save(&f).Error; err != nil {
		return model.CustomField{}, err
	}
	return f, nil
}

func (r *repository) DeleteCustomField(ctx context.Context, id int64, workspaceID int64) error {
	res := r.conn(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&model.CustomField{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) ClearCustomFieldFromTasks(ctx context.Context, workspaceID int64, fieldKey string) error {
	var tasks []model.Task
	if err := r.conn(ctx).
		Where("workspace_id = ? AND custom_fields IS NOT NULL", workspaceID).
		Find(&tasks).Error; err != nil {
		return err
	}
	for _, t := range tasks {
		if len(t.CustomFields) == 0 {
			continue
		}
		var fields map[string]any
		if err := json.Unmarshal(t.CustomFields, &fields); err != nil {
			continue
		}
		if _, exists := fields[fieldKey]; !exists {
			continue
		}
		delete(fields, fieldKey)
		updated, err := json.Marshal(fields)
		if err != nil {
			continue
		}
		if err := r.conn(ctx).Model(&model.Task{}).Where("id = ?", t.ID).Update("custom_fields", updated).Error; err != nil {
			return err
		}
	}
	return nil
}
