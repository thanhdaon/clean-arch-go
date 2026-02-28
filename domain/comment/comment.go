package comment

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type ReferenceType string

const (
	ReferenceTypeUser ReferenceType = "user"
	ReferenceTypeTask ReferenceType = "task"
)

type Reference struct {
	Type ReferenceType
	ID   string
}

var mentionRe = regexp.MustCompile(`@\[(\w+):([^\]]+)\]`)

type Comment interface {
	UUID() string
	TaskID() string
	AuthorID() string
	Content() string
	References() []Reference
	CreatedAt() time.Time
	UpdatedAt() time.Time
	IsDeleted() bool
	Update(editorID, content string) error
	Delete(deleterID string) error
}

type commentRecord struct {
	uuid       string
	taskID     string
	authorID   string
	content    string
	references []Reference
	createdAt  time.Time
	updatedAt  time.Time
	deletedAt  time.Time
}

func (c *commentRecord) UUID() string            { return c.uuid }
func (c *commentRecord) TaskID() string          { return c.taskID }
func (c *commentRecord) AuthorID() string        { return c.authorID }
func (c *commentRecord) Content() string         { return c.content }
func (c *commentRecord) References() []Reference { return c.references }
func (c *commentRecord) CreatedAt() time.Time    { return c.createdAt }
func (c *commentRecord) UpdatedAt() time.Time    { return c.updatedAt }
func (c *commentRecord) IsDeleted() bool         { return !c.deletedAt.IsZero() }

func (c *commentRecord) Update(editorID, content string) error {
	if editorID != c.authorID {
		return errors.New("only comment author can edit this comment")
	}
	if content == "" {
		return errors.New("empty content")
	}
	c.content = content
	c.references = parseReferences(content)
	c.updatedAt = time.Now()
	return nil
}

func (c *commentRecord) Delete(deleterID string) error {
	if deleterID != c.authorID {
		return errors.New("only comment author can delete this comment")
	}
	c.deletedAt = time.Now()
	return nil
}

func parseReferences(content string) []Reference {
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	refs := make([]Reference, 0, len(matches))
	for _, m := range matches {
		refType := ReferenceType(m[1])
		if refType != ReferenceTypeUser && refType != ReferenceTypeTask {
			continue
		}
		refs = append(refs, Reference{Type: refType, ID: m[2]})
	}
	return refs
}

func New(taskID, authorID, content string) (Comment, error) {
	if taskID == "" {
		return nil, errors.New("empty task id")
	}
	if authorID == "" {
		return nil, errors.New("empty author id")
	}
	if content == "" {
		return nil, errors.New("empty content")
	}
	return &commentRecord{
		uuid:       uuid.New().String(),
		taskID:     taskID,
		authorID:   authorID,
		content:    content,
		references: parseReferences(content),
		createdAt:  time.Now(),
	}, nil
}

func From(id, taskID, authorID, content string, createdAt, updatedAt time.Time, deletedAt *time.Time) (Comment, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	if taskID == "" {
		return nil, errors.New("empty task id")
	}
	if authorID == "" {
		return nil, errors.New("empty author id")
	}
	c := &commentRecord{
		uuid:       id,
		taskID:     taskID,
		authorID:   authorID,
		content:    content,
		references: parseReferences(content),
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
	if deletedAt != nil {
		c.deletedAt = *deletedAt
	}
	return c, nil
}
