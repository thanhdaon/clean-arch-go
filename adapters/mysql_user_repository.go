package adapters

import (
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/user"

	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type MysqlUser struct {
	ID   string `db:"id"`
	Role string `db:"role"`
}

type MysqlUserRepository struct {
	db *sqlx.DB
}

func NewMysqlUserRepository(db *sqlx.DB) MysqlUserRepository {
	return MysqlUserRepository{db: db}
}

func (r MysqlUserRepository) Add(ctx context.Context, u user.User) error {
	op := errors.Op("MysqlUserRepository.Add")

	added := MysqlUser{
		ID:   u.UUID(),
		Role: u.Role().String(),
	}

	query := `
		INSERT INTO users
			(id, role) 
		VALUES 
			(:id, :role)
	`

	if _, err := r.db.NamedExecContext(ctx, query, added); err != nil {
		return errors.E(op, errkind.Internal, err)
	}

	return nil
}

func (r MysqlUserRepository) UpdateByID(ctx context.Context, uuid string, updateFn func(context.Context, user.User) (user.User, error)) error {
	op := errors.Op("MysqlUserRepository.UpdateByID")

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.E(op, errkind.Internal, err)
	}

	defer func() {
		err = r.finishTransaction(err, tx)
	}()

	existingUser, err := r.FindById(ctx, uuid)
	if err != nil {
		if errors.Is(errkind.NotExist, err) {
			return errors.E(op, errkind.NotExist, err)
		}

		return errors.E(op, errkind.Internal, err)
	}

	updatedUser, err := updateFn(ctx, existingUser)
	if err != nil {
		return errors.E(op, errkind.Internal, err)
	}

	updated := MysqlUser{
		ID:   updatedUser.UUID(),
		Role: updatedUser.Role().String(),
	}

	query := `
		UPDATE users 
		SET 
			role = :role
		WHERE id = :id;
	`
	result, err := r.db.NamedExecContext(ctx, query, updated)
	if err != nil {
		return errors.E(op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.E(op, err)
	}

	if rowsAffected == 0 {
		return errors.E(op, errkind.NotExist, fmt.Errorf("user with id %s not found", uuid))
	}

	return nil
}

func (r MysqlUserRepository) FindById(ctx context.Context, uuid string) (user.User, error) {
	op := errors.Op("MysqlUserRepository.FindById")

	data := MysqlUser{}
	query := "SELECT * FROM `users` WHERE `id` = ?"

	if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errkind.NotExist, err)
		}

		return nil, errors.E(op, err)
	}

	domainUser, err := user.From(data.ID, data.Role)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return domainUser, nil
}

func (r MysqlUserRepository) RemoveAll(ctx context.Context) error {
	op := errors.Op("MysqlUserRepository.RemoveAll")

	query := `TRUNCATE TABLE users`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlUserRepository) finishTransaction(err error, tx *sqlx.Tx) error {
	op := errors.Op("MysqlUserRepository.finishTransaction")

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.E(op, rollbackErr)
		}

		return errors.E(op, err)
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			return errors.E(op, "failed to commit transaction")
		}

		return nil
	}
}
