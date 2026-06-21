package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	entity "github.com/agentrq/agentrq/backend/internal/data/entity/crud"
	"github.com/agentrq/agentrq/backend/internal/data/model"
	"github.com/mustafaturan/monoflake"
	"gorm.io/datatypes"
)

func (c *controller) CreateWorkspaceTemplate(ctx context.Context, req entity.CreateWorkspaceTemplateRequest) (*entity.CreateWorkspaceTemplateResponse, error) {
	userID := monoflake.IDFromBase62(req.UserID).Int64()
	now := time.Now()
	m := model.WorkspaceTemplate{
		ID:                   c.idgen.NextID(),
		CreatedAt:            now,
		UpdatedAt:            now,
		UserID:               userID,
		Name:                 req.Template.Name,
		Description:          req.Template.Description,
		AllowAllCommands:     req.Template.AllowAllCommands,
		SelfLearningLoopNote: req.Template.SelfLearningLoopNote,
	}
	if req.Template.ColumnConfig != nil {
		b, _ := json.Marshal(req.Template.ColumnConfig)
		m.ColumnConfig = datatypes.JSON(b)
	}
	if req.Template.FilterConfig != nil {
		b, _ := json.Marshal(req.Template.FilterConfig)
		m.FilterConfig = datatypes.JSON(b)
	}
	if req.Template.AutoAllowedTools != nil {
		b, _ := json.Marshal(req.Template.AutoAllowedTools)
		m.AutoAllowedTools = datatypes.JSON(b)
	}
	if req.Template.NotificationSettings != nil {
		b, _ := json.Marshal(req.Template.NotificationSettings)
		m.NotificationSettings = datatypes.JSON(b)
	}

	created, err := c.repository.CreateWorkspaceTemplate(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("create workspace template: %w", err)
	}
	return &entity.CreateWorkspaceTemplateResponse{
		Template: fromModelTemplateToEntity(created),
	}, nil
}

func (c *controller) GetWorkspaceTemplate(ctx context.Context, req entity.GetWorkspaceTemplateRequest) (*entity.GetWorkspaceTemplateResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	m, err := c.repository.GetWorkspaceTemplate(ctx, req.ID, uid)
	if err != nil {
		return nil, err
	}
	return &entity.GetWorkspaceTemplateResponse{
		Template: fromModelTemplateToEntity(m),
	}, nil
}

func (c *controller) ListWorkspaceTemplates(ctx context.Context, req entity.ListWorkspaceTemplatesRequest) (*entity.ListWorkspaceTemplatesResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	ms, err := c.repository.ListWorkspaceTemplates(ctx, uid)
	if err != nil {
		return nil, err
	}
	templates := make([]entity.WorkspaceTemplate, len(ms))
	for i, m := range ms {
		templates[i] = fromModelTemplateToEntity(m)
	}
	return &entity.ListWorkspaceTemplatesResponse{Templates: templates}, nil
}

func (c *controller) UpdateWorkspaceTemplate(ctx context.Context, req entity.UpdateWorkspaceTemplateRequest) (*entity.UpdateWorkspaceTemplateResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	m, err := c.repository.GetWorkspaceTemplate(ctx, req.Template.ID, uid)
	if err != nil {
		return nil, err
	}

	m.Name = req.Template.Name
	m.Description = req.Template.Description
	m.AllowAllCommands = req.Template.AllowAllCommands
	m.SelfLearningLoopNote = req.Template.SelfLearningLoopNote
	m.UpdatedAt = time.Now()

	if req.Template.ColumnConfig != nil {
		b, _ := json.Marshal(req.Template.ColumnConfig)
		m.ColumnConfig = datatypes.JSON(b)
	}
	if req.Template.FilterConfig != nil {
		b, _ := json.Marshal(req.Template.FilterConfig)
		m.FilterConfig = datatypes.JSON(b)
	}
	if req.Template.AutoAllowedTools != nil {
		b, _ := json.Marshal(req.Template.AutoAllowedTools)
		m.AutoAllowedTools = datatypes.JSON(b)
	}
	if req.Template.NotificationSettings != nil {
		b, _ := json.Marshal(req.Template.NotificationSettings)
		m.NotificationSettings = datatypes.JSON(b)
	}

	updated, err := c.repository.UpdateWorkspaceTemplate(ctx, m)
	if err != nil {
		return nil, err
	}
	return &entity.UpdateWorkspaceTemplateResponse{
		Template: fromModelTemplateToEntity(updated),
	}, nil
}

func (c *controller) DeleteWorkspaceTemplate(ctx context.Context, req entity.DeleteWorkspaceTemplateRequest) error {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	return c.repository.DeleteWorkspaceTemplate(ctx, req.ID, uid)
}

func (c *controller) SaveWorkspaceAsTemplate(ctx context.Context, req entity.SaveWorkspaceAsTemplateRequest) (*entity.SaveWorkspaceAsTemplateResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	w, err := c.repository.GetWorkspace(ctx, req.WorkspaceID, uid)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	m := model.WorkspaceTemplate{
		ID:                   c.idgen.NextID(),
		CreatedAt:            now,
		UpdatedAt:            now,
		UserID:               uid,
		Name:                 req.Name,
		Description:          req.Description,
		AutoAllowedTools:     w.AutoAllowedTools,
		AllowAllCommands:     w.AllowAllCommands,
		NotificationSettings: w.NotificationSettings,
		SelfLearningLoopNote: w.SelfLearningLoopNote,
	}

	created, err := c.repository.CreateWorkspaceTemplate(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("save workspace as template: %w", err)
	}
	return &entity.SaveWorkspaceAsTemplateResponse{
		Template: fromModelTemplateToEntity(created),
	}, nil
}

func (c *controller) ApplyTemplateToWorkspace(ctx context.Context, req entity.ApplyTemplateToWorkspaceRequest) (*entity.ApplyTemplateToWorkspaceResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	tmpl, err := c.repository.GetWorkspaceTemplate(ctx, req.TemplateID, uid)
	if err != nil {
		return nil, err
	}
	w, err := c.repository.GetWorkspace(ctx, req.WorkspaceID, uid)
	if err != nil {
		return nil, err
	}
	if w.ArchivedAt != nil {
		return nil, fmt.Errorf("cannot apply template to archived workspace")
	}

	w.AutoAllowedTools = tmpl.AutoAllowedTools
	w.AllowAllCommands = tmpl.AllowAllCommands
	w.NotificationSettings = tmpl.NotificationSettings
	w.SelfLearningLoopNote = tmpl.SelfLearningLoopNote
	w.TemplateID = tmpl.ID
	w.UpdatedAt = time.Now()

	updated, err := c.repository.UpdateWorkspace(ctx, w)
	if err != nil {
		return nil, err
	}
	return &entity.ApplyTemplateToWorkspaceResponse{
		Workspace: fromModelWorkspaceToEntity(updated),
	}, nil
}

func fromModelTemplateToEntity(m model.WorkspaceTemplate) entity.WorkspaceTemplate {
	res := entity.WorkspaceTemplate{
		ID:                   m.ID,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
		UserID:               m.UserID,
		Name:                 m.Name,
		Description:          m.Description,
		ColumnConfig:         make([]string, 0),
		FilterConfig:         make(map[string]any),
		AutoAllowedTools:     make([]string, 0),
		AllowAllCommands:     m.AllowAllCommands,
		SelfLearningLoopNote: m.SelfLearningLoopNote,
	}
	if len(m.ColumnConfig) > 0 {
		_ = json.Unmarshal(m.ColumnConfig, &res.ColumnConfig)
	}
	if len(m.FilterConfig) > 0 {
		_ = json.Unmarshal(m.FilterConfig, &res.FilterConfig)
	}
	if len(m.AutoAllowedTools) > 0 {
		_ = json.Unmarshal(m.AutoAllowedTools, &res.AutoAllowedTools)
	}
	if len(m.NotificationSettings) > 0 {
		var ns entity.NotificationSettings
		if err := json.Unmarshal(m.NotificationSettings, &ns); err == nil {
			res.NotificationSettings = &ns
		}
	}
	return res
}
