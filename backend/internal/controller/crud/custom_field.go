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

var validFieldTypes = map[string]bool{
	"text":        true,
	"number":      true,
	"select":      true,
	"multiselect": true,
	"date":        true,
}

func (c *controller) CreateCustomField(ctx context.Context, req entity.CreateCustomFieldRequest) (*entity.CreateCustomFieldResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	if _, err := c.repository.GetWorkspace(ctx, req.WorkspaceID, uid); err != nil {
		return nil, err
	}
	if req.Field.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if !validFieldTypes[req.Field.FieldType] {
		return nil, fmt.Errorf("invalid field type: %s", req.Field.FieldType)
	}

	now := time.Now()
	m := model.CustomField{
		ID:          c.idgen.NextID(),
		CreatedAt:   now,
		UpdatedAt:   now,
		WorkspaceID: req.WorkspaceID,
		UserID:      uid,
		Name:        req.Field.Name,
		FieldType:   req.Field.FieldType,
		SortOrder:   req.Field.SortOrder,
	}
	if req.Field.Options != nil {
		b, _ := json.Marshal(req.Field.Options)
		m.Options = datatypes.JSON(b)
	}

	created, err := c.repository.CreateCustomField(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("create custom field: %w", err)
	}
	return &entity.CreateCustomFieldResponse{
		Field: fromModelCustomFieldToEntity(created),
	}, nil
}

func (c *controller) GetCustomFields(ctx context.Context, req entity.ListCustomFieldsRequest) (*entity.ListCustomFieldsResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	if _, err := c.repository.GetWorkspace(ctx, req.WorkspaceID, uid); err != nil {
		return nil, err
	}
	ms, err := c.repository.ListCustomFields(ctx, req.WorkspaceID)
	if err != nil {
		return nil, err
	}
	fields := make([]entity.CustomFieldDefinition, len(ms))
	for i, m := range ms {
		fields[i] = fromModelCustomFieldToEntity(m)
	}
	return &entity.ListCustomFieldsResponse{Fields: fields}, nil
}

func (c *controller) UpdateCustomField(ctx context.Context, req entity.UpdateCustomFieldRequest) (*entity.UpdateCustomFieldResponse, error) {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	if _, err := c.repository.GetWorkspace(ctx, req.WorkspaceID, uid); err != nil {
		return nil, err
	}
	m, err := c.repository.GetCustomField(ctx, req.Field.ID, req.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if req.Field.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if !validFieldTypes[req.Field.FieldType] {
		return nil, fmt.Errorf("invalid field type: %s", req.Field.FieldType)
	}

	m.Name = req.Field.Name
	m.FieldType = req.Field.FieldType
	m.SortOrder = req.Field.SortOrder
	m.UpdatedAt = time.Now()
	if req.Field.Options != nil {
		b, _ := json.Marshal(req.Field.Options)
		m.Options = datatypes.JSON(b)
	}

	updated, err := c.repository.UpdateCustomField(ctx, m)
	if err != nil {
		return nil, err
	}
	return &entity.UpdateCustomFieldResponse{
		Field: fromModelCustomFieldToEntity(updated),
	}, nil
}

func (c *controller) DeleteCustomField(ctx context.Context, req entity.DeleteCustomFieldRequest) error {
	uid := monoflake.IDFromBase62(req.UserID).Int64()
	if _, err := c.repository.GetWorkspace(ctx, req.WorkspaceID, uid); err != nil {
		return err
	}
	f, err := c.repository.GetCustomField(ctx, req.FieldID, req.WorkspaceID)
	if err != nil {
		return err
	}
	if err := c.repository.DeleteCustomField(ctx, req.FieldID, req.WorkspaceID); err != nil {
		return err
	}
	return c.repository.ClearCustomFieldFromTasks(ctx, req.WorkspaceID, f.Name)
}

func fromModelCustomFieldToEntity(m model.CustomField) entity.CustomFieldDefinition {
	res := entity.CustomFieldDefinition{
		ID:          m.ID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Name:        m.Name,
		FieldType:   m.FieldType,
		Options:     make([]string, 0),
		SortOrder:   m.SortOrder,
	}
	if len(m.Options) > 0 {
		_ = json.Unmarshal(m.Options, &res.Options)
	}
	return res
}
