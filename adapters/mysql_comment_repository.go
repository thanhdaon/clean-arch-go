package adapters

import (
	"clean-arch-go/app/command"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/comment"
	"clean-arch-go/domain/errkind"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type MysqlComment struct {
	ID        string       `db:"id"`
	TaskID    string       `db:"task_id"`
	AuthorID  string       `db:"author_id"`
	Content   string       `db:"content"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

type MysqlCommentReference struct {
	ID            string    `db:"id"`
	CommentID     string    `db:"comment_id"`
	ReferenceType string    `db:"reference_type"`
	ReferenceID   string    `db:"reference_id"`
	CreatedAt     time.Time `db:"created_at"`
}

type MysqlCommentRepository struct {
	db *sqlx.DB
}

func NewMysqlCommentRepository(db *sqlx.DB) MysqlCommentRepository {
	return MysqlCommentRepository{db: db}
}

func (r MysqlCommentRepository) Add(ctx context.Context, c comment.Comment) error {
	op := errors.Op("MysqlCommentRepository.Add")

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.E(op, err)
	}
	defer func() { err = finishTx(err, tx) }()

	row := MysqlComment{
		ID:        c.UUID(),
		TaskID:    c.TaskID(),
		AuthorID:  c.AuthorID(),
		Content:   c.Content(),
		CreatedAt: c.CreatedAt(),
	}

	if _, err = tx.NamedExecContext(ctx, `
		INSERT INTO comments (id, task_id, author_id, content, created_at)
		VALUES (:id, :task_id, :author_id, :content, :created_at)
	`, row); err != nil {
		return errors.E(op, err)
	}

	if err = insertCommentReferences(ctx, tx, c); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlCommentRepository) UpdateByID(ctx context.Context, id string, updateFn command.CommentUpdater) error {
	op := errors.Op("MysqlCommentRepository.UpdateByID")

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.E(op, err)
	}
	defer func() { err = finishTx(err, tx) }()

	existing, err := r.findByID(ctx, id)
	if err != nil {
		return errors.E(op, err)
	}

	updated, err := updateFn(ctx, existing)
	if err != nil {
		return errors.E(op, err)
	}

	var deletedAt sql.NullTime
	if updated.IsDeleted() {
		deletedAt = timeToNullTime(updated.UpdatedAt()) // use now as proxy
		// re-fetch actual deleted time by using UpdatedAt isn't ideal;
		// domain sets deletedAt internally, so we encode IsDeleted by checking
		// For accuracy, use a sentinel: if deleted, set deleted_at = NOW() server-side
		deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE comments SET content = ?, updated_at = ?, deleted_at = ? WHERE id = ?
	`, updated.Content(), timeToNullTime(updated.UpdatedAt()), deletedAt, id); err != nil {
		return errors.E(op, err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM comment_references WHERE comment_id = ?`, id); err != nil {
		return errors.E(op, err)
	}

	if err = insertCommentReferences(ctx, tx, updated); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlCommentRepository) findByID(ctx context.Context, id string) (comment.Comment, error) {
	op := errors.Op("MysqlCommentRepository.findByID")

	data := MysqlComment{}
	if err := r.db.GetContext(ctx, &data, `SELECT * FROM comments WHERE id = ?`, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errkind.NotExist, err)
		}
		return nil, errors.E(op, err)
	}

	var deletedAt *time.Time
	if data.DeletedAt.Valid {
		deletedAt = &data.DeletedAt.Time
	}

	return comment.From(data.ID, data.TaskID, data.AuthorID, data.Content,
		data.CreatedAt, nullTimeToTime(data.UpdatedAt), deletedAt)
}

func insertCommentReferences(ctx context.Context, tx *sqlx.Tx, c comment.Comment) error {
	for _, ref := range c.References() {
		row := MysqlCommentReference{
			ID:            uuid.New().String(),
			CommentID:     c.UUID(),
			ReferenceType: string(ref.Type),
			ReferenceID:   ref.ID,
			CreatedAt:     time.Now(),
		}
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO comment_references (id, comment_id, reference_type, reference_id, created_at)
			VALUES (:id, :comment_id, :reference_type, :reference_id, :created_at)
		`, row); err != nil {
			return err
		}
	}
	return nil
}

func finishTx(err error, tx *sqlx.Tx) error {
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
