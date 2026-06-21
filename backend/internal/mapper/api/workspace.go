package api

import (
	"encoding/json"

	entity "github.com/agentrq/agentrq/backend/internal/data/entity/crud"
	view "github.com/agentrq/agentrq/backend/internal/data/view/api"
	"github.com/gofiber/fiber/v2"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateWorkspaceRequestEntity(c *fiber.Ctx) *entity.CreateWorkspaceRequest {
	var payload view.CreateWorkspaceRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	if payload.Workspace.Name == "" {
		return nil
	}
	return &entity.CreateWorkspaceRequest{
		Workspace: entity.Workspace{
			Name:                 payload.Workspace.Name,
			Description:          payload.Workspace.Description,
			Icon:                 payload.Workspace.Icon,
			NotificationSettings: fromViewNotificationSettingsToEntity(payload.Workspace.NotificationSettings),
			AllowAllCommands:     payload.Workspace.AllowAllCommands,
			SelfLearningLoopNote: payload.Workspace.SelfLearningLoopNote,
			TemplateID:           monoflake.IDFromBase62(payload.Workspace.TemplateID).Int64(),
		},
	}
}

func FromCreateWorkspaceResponseEntityToHTTPResponse(rs *entity.CreateWorkspaceResponse, mcpURL string) []byte {
	payload, _ := json.Marshal(view.CreateWorkspaceResponse{
		Workspace: fromEntityWorkspaceToView(rs.Workspace, mcpURL),
	})
	return payload
}

func FromCreateWorkspaceResponseEntityToMCPResponse(rs *entity.CreateWorkspaceResponse, mcpURL string) []byte {
	w := fromEntityWorkspaceToView(rs.Workspace, mcpURL)
	w.Icon = ""
	payload, _ := json.Marshal(view.CreateWorkspaceResponse{
		Workspace: w,
	})
	return payload
}

func FromGetWorkspaceResponseEntityToHTTPResponse(rs *entity.GetWorkspaceResponse, mcpURL string) []byte {
	payload, _ := json.Marshal(view.GetWorkspaceResponse{
		Workspace: fromEntityWorkspaceToView(rs.Workspace, mcpURL),
	})
	return payload
}

func FromGetWorkspaceResponseEntityToMCPResponse(rs *entity.GetWorkspaceResponse, mcpURL string) []byte {
	w := fromEntityWorkspaceToView(rs.Workspace, mcpURL)
	w.Icon = ""
	payload, _ := json.Marshal(view.GetWorkspaceResponse{
		Workspace: w,
	})
	return payload
}

func FromListWorkspacesResponseEntityToHTTPResponse(rs *entity.ListWorkspacesResponse, mcpURLFn func(int64) string) []byte {
	workspaces := make([]view.Workspace, len(rs.Workspaces))
	for i, p := range rs.Workspaces {
		workspaces[i] = fromEntityWorkspaceToView(p, mcpURLFn(p.ID))
	}
	payload, _ := json.Marshal(view.ListWorkspacesResponse{Workspaces: workspaces})
	return payload
}

func FromListWorkspacesResponseEntityToMCPResponse(rs *entity.ListWorkspacesResponse, mcpURLFn func(int64) string) []byte {
	workspaces := make([]view.Workspace, len(rs.Workspaces))
	for i, p := range rs.Workspaces {
		w := fromEntityWorkspaceToView(p, mcpURLFn(p.ID))
		w.Icon = ""
		workspaces[i] = w
	}
	payload, _ := json.Marshal(view.ListWorkspacesResponse{Workspaces: workspaces})
	return payload
}

func FromHTTPRequestToGetWorkspaceRequestEntity(c *fiber.Ctx) *entity.GetWorkspaceRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	return &entity.GetWorkspaceRequest{ID: id}
}

func FromHTTPRequestToDeleteWorkspaceRequestEntity(c *fiber.Ctx) *entity.DeleteWorkspaceRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	return &entity.DeleteWorkspaceRequest{ID: id}
}

func FromHTTPRequestToUpdateWorkspaceRequestEntity(c *fiber.Ctx) *entity.UpdateWorkspaceRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	var payload view.UpdateWorkspaceRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	return &entity.UpdateWorkspaceRequest{
		Workspace: entity.Workspace{
			ID:                   id,
			Name:                 payload.Workspace.Name,
			Description:          payload.Workspace.Description,
			Icon:                 payload.Workspace.Icon,
			NotificationSettings: fromViewNotificationSettingsToEntity(payload.Workspace.NotificationSettings),
			AutoAllowedTools:     payload.Workspace.AutoAllowedTools,
			AllowAllCommands:     payload.Workspace.AllowAllCommands,
			SelfLearningLoopNote: payload.Workspace.SelfLearningLoopNote,
			TemplateID:           monoflake.IDFromBase62(payload.Workspace.TemplateID).Int64(),
		},
	}
}

func FromUpdateWorkspaceResponseEntityToHTTPResponse(rs *entity.Workspace, mcpURL string) []byte {
	payload, _ := json.Marshal(view.GetWorkspaceResponse{
		Workspace: fromEntityWorkspaceToView(*rs, mcpURL),
	})
	return payload
}

func FromUpdateWorkspaceResponseEntityToMCPResponse(rs *entity.Workspace, mcpURL string) []byte {
	w := fromEntityWorkspaceToView(*rs, mcpURL)
	w.Icon = ""
	payload, _ := json.Marshal(view.GetWorkspaceResponse{
		Workspace: w,
	})
	return payload
}

func fromEntityWorkspaceToView(p entity.Workspace, mcpURL string) view.Workspace {
	v := view.Workspace{
		ID:                   monoflake.ID(p.ID).String(),
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
		Name:                 p.Name,
		Description:          p.Description,
		Icon:                 p.Icon,
		ArchivedAt:           p.ArchivedAt,
		NotificationSettings: fromEntityNotificationSettingsToView(p.NotificationSettings),
		AgentConnected:       p.AgentConnected,
		MCPURL:               mcpURL,
		AutoAllowedTools:     p.AutoAllowedTools,
		AllowAllCommands:     p.AllowAllCommands,
		SelfLearningLoopNote: p.SelfLearningLoopNote,
		TemplateID:           monoflake.ID(p.TemplateID).String(),
	}
	if p.Slack != nil {
		v.Slack = &view.SlackConfig{
			Enabled:     p.Slack.Enabled,
			Installed:   p.Slack.Installed,
			ChannelID:   p.Slack.ChannelID,
			ChannelName: p.Slack.ChannelName,
			AutoCreated: p.Slack.AutoCreated,
			ClientID:    p.Slack.ClientID,
			AuthURL:     p.Slack.AuthURL,
		}
	}
	return v
}

func fromEntityNotificationSettingsToView(p *entity.NotificationSettings) *view.NotificationSettings {
	if p == nil {
		return nil
	}
	return &view.NotificationSettings{
		TaskCreated:         p.TaskCreated,
		TaskStatusUpdated:   p.TaskStatusUpdated,
		TaskReceivedMessage: p.TaskReceivedMessage,
		WorkspaceArchived:   p.WorkspaceArchived,
		WorkspaceUnarchived: p.WorkspaceUnarchived,
		Channels:            p.Channels,
	}
}

func fromViewNotificationSettingsToEntity(p *view.NotificationSettings) *entity.NotificationSettings {
	if p == nil {
		return nil
	}
	return &entity.NotificationSettings{
		TaskCreated:         p.TaskCreated,
		TaskStatusUpdated:   p.TaskStatusUpdated,
		TaskReceivedMessage: p.TaskReceivedMessage,
		WorkspaceArchived:   p.WorkspaceArchived,
		WorkspaceUnarchived: p.WorkspaceUnarchived,
		Channels:            p.Channels,
	}
}

func FromHTTPRequestToCreateWorkspaceTemplateRequest(c *fiber.Ctx) *entity.CreateWorkspaceTemplateRequest {
	var payload view.CreateWorkspaceTemplateRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	if payload.Template.Name == "" {
		return nil
	}
	return &entity.CreateWorkspaceTemplateRequest{
		Template: entity.WorkspaceTemplate{
			Name:                 payload.Template.Name,
			Description:          payload.Template.Description,
			ColumnConfig:         payload.Template.ColumnConfig,
			FilterConfig:         fromViewFilterConfigToEntity(payload.Template.FilterConfig),
			AutoAllowedTools:     payload.Template.AutoAllowedTools,
			AllowAllCommands:     payload.Template.AllowAllCommands,
			NotificationSettings: fromViewNotificationSettingsToEntity(payload.Template.NotificationSettings),
			SelfLearningLoopNote: payload.Template.SelfLearningLoopNote,
		},
	}
}

func FromCreateWorkspaceTemplateResponseEntityToHTTPResponse(rs *entity.CreateWorkspaceTemplateResponse) []byte {
	payload, _ := json.Marshal(view.CreateWorkspaceTemplateResponse{
		Template: fromEntityTemplateToView(rs.Template),
	})
	return payload
}

func FromGetWorkspaceTemplateResponseEntityToHTTPResponse(rs *entity.GetWorkspaceTemplateResponse) []byte {
	payload, _ := json.Marshal(view.GetWorkspaceTemplateResponse{
		Template: fromEntityTemplateToView(rs.Template),
	})
	return payload
}

func FromListWorkspaceTemplatesResponseEntityToHTTPResponse(rs *entity.ListWorkspaceTemplatesResponse) []byte {
	templates := make([]view.WorkspaceTemplate, len(rs.Templates))
	for i, t := range rs.Templates {
		templates[i] = fromEntityTemplateToView(t)
	}
	payload, _ := json.Marshal(view.ListWorkspaceTemplatesResponse{Templates: templates})
	return payload
}

func FromHTTPRequestToUpdateWorkspaceTemplateRequest(c *fiber.Ctx) *entity.UpdateWorkspaceTemplateRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	var payload view.UpdateWorkspaceTemplateRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	return &entity.UpdateWorkspaceTemplateRequest{
		Template: entity.WorkspaceTemplate{
			ID:                   id,
			Name:                 payload.Template.Name,
			Description:          payload.Template.Description,
			ColumnConfig:         payload.Template.ColumnConfig,
			FilterConfig:         fromViewFilterConfigToEntity(payload.Template.FilterConfig),
			AutoAllowedTools:     payload.Template.AutoAllowedTools,
			AllowAllCommands:     payload.Template.AllowAllCommands,
			NotificationSettings: fromViewNotificationSettingsToEntity(payload.Template.NotificationSettings),
			SelfLearningLoopNote: payload.Template.SelfLearningLoopNote,
		},
	}
}

func FromUpdateWorkspaceTemplateResponseEntityToHTTPResponse(rs *entity.UpdateWorkspaceTemplateResponse) []byte {
	payload, _ := json.Marshal(view.UpdateWorkspaceTemplateResponse{
		Template: fromEntityTemplateToView(rs.Template),
	})
	return payload
}

func FromHTTPRequestToSaveWorkspaceAsTemplateRequest(c *fiber.Ctx) *entity.SaveWorkspaceAsTemplateRequest {
	var payload view.SaveWorkspaceAsTemplateRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	workspaceID := monoflake.IDFromBase62(payload.WorkspaceID).Int64()
	if workspaceID == 0 || payload.Name == "" {
		return nil
	}
	return &entity.SaveWorkspaceAsTemplateRequest{
		WorkspaceID: workspaceID,
		Name:        payload.Name,
		Description: payload.Description,
	}
}

func FromSaveWorkspaceAsTemplateResponseEntityToHTTPResponse(rs *entity.SaveWorkspaceAsTemplateResponse) []byte {
	payload, _ := json.Marshal(view.SaveWorkspaceAsTemplateResponse{
		Template: fromEntityTemplateToView(rs.Template),
	})
	return payload
}

func FromHTTPRequestToApplyTemplateToWorkspaceRequest(c *fiber.Ctx) *entity.ApplyTemplateToWorkspaceRequest {
	var payload view.ApplyTemplateToWorkspaceRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	templateID := monoflake.IDFromBase62(payload.TemplateID).Int64()
	workspaceID := monoflake.IDFromBase62(payload.WorkspaceID).Int64()
	if templateID == 0 || workspaceID == 0 {
		return nil
	}
	return &entity.ApplyTemplateToWorkspaceRequest{
		TemplateID:  templateID,
		WorkspaceID: workspaceID,
	}
}

func FromApplyTemplateToWorkspaceResponseEntityToHTTPResponse(rs *entity.ApplyTemplateToWorkspaceResponse, mcpURL string) []byte {
	payload, _ := json.Marshal(view.ApplyTemplateToWorkspaceResponse{
		Workspace: fromEntityWorkspaceToView(rs.Workspace, mcpURL),
	})
	return payload
}

func fromEntityTemplateToView(p entity.WorkspaceTemplate) view.WorkspaceTemplate {
	return view.WorkspaceTemplate{
		ID:                   monoflake.ID(p.ID).String(),
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
		Name:                 p.Name,
		Description:          p.Description,
		ColumnConfig:         p.ColumnConfig,
		FilterConfig:         p.FilterConfig,
		AutoAllowedTools:     p.AutoAllowedTools,
		AllowAllCommands:     p.AllowAllCommands,
		NotificationSettings: fromEntityNotificationSettingsToView(p.NotificationSettings),
		SelfLearningLoopNote: p.SelfLearningLoopNote,
	}
}

func FromHTTPRequestToCreateCustomFieldRequest(c *fiber.Ctx) *entity.CreateCustomFieldRequest {
	var payload view.CreateCustomFieldRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	workspaceID := monoflake.IDFromBase62(c.Params("workspaceId")).Int64()
	if workspaceID == 0 || payload.Field.Name == "" {
		return nil
	}
	return &entity.CreateCustomFieldRequest{
		WorkspaceID: workspaceID,
		Field: entity.CustomFieldDefinition{
			Name:      payload.Field.Name,
			FieldType: payload.Field.FieldType,
			Options:   payload.Field.Options,
			SortOrder: payload.Field.SortOrder,
		},
	}
}

func FromCreateCustomFieldResponseEntityToHTTPResponse(rs *entity.CreateCustomFieldResponse) []byte {
	payload, _ := json.Marshal(view.CreateCustomFieldResponse{
		Field: fromEntityCustomFieldToView(rs.Field),
	})
	return payload
}

func FromListCustomFieldsResponseEntityToHTTPResponse(rs *entity.ListCustomFieldsResponse) []byte {
	fields := make([]view.CustomFieldDefinition, len(rs.Fields))
	for i, f := range rs.Fields {
		fields[i] = fromEntityCustomFieldToView(f)
	}
	payload, _ := json.Marshal(view.ListCustomFieldsResponse{Fields: fields})
	return payload
}

func FromHTTPRequestToUpdateCustomFieldRequest(c *fiber.Ctx) *entity.UpdateCustomFieldRequest {
	fieldID := monoflake.IDFromBase62(c.Params("fieldId")).Int64()
	if fieldID == 0 {
		return nil
	}
	var payload view.UpdateCustomFieldRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	workspaceID := monoflake.IDFromBase62(c.Params("workspaceId")).Int64()
	if workspaceID == 0 || payload.Field.Name == "" {
		return nil
	}
	return &entity.UpdateCustomFieldRequest{
		WorkspaceID: workspaceID,
		Field: entity.CustomFieldDefinition{
			ID:          fieldID,
			Name:        payload.Field.Name,
			FieldType:   payload.Field.FieldType,
			Options:     payload.Field.Options,
			SortOrder:   payload.Field.SortOrder,
		},
	}
}

func FromUpdateCustomFieldResponseEntityToHTTPResponse(rs *entity.UpdateCustomFieldResponse) []byte {
	payload, _ := json.Marshal(view.UpdateCustomFieldResponse{
		Field: fromEntityCustomFieldToView(rs.Field),
	})
	return payload
}

func fromEntityCustomFieldToView(p entity.CustomFieldDefinition) view.CustomFieldDefinition {
	return view.CustomFieldDefinition{
		ID:          monoflake.ID(p.ID).String(),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		WorkspaceID: monoflake.ID(p.WorkspaceID).String(),
		Name:        p.Name,
		FieldType:   p.FieldType,
		Options:     p.Options,
		SortOrder:   p.SortOrder,
	}
}

func fromViewFilterConfigToEntity(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}
