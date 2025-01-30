package adapters

import (
	"clean-arch-go/domain/errors"
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

	if _, err := r.db.NamedExec(query, added); err != nil {
		return errors.E(op, errors.Internal, err)
	}

	return nil
}

func (r MysqlUserRepository) UpdateByID(ctx context.Context, uuid string, updateFn func(context.Context, user.User) (user.User, error)) error {
	op := errors.Op("MysqlUserRepository.UpdateByID")

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.E(op, errors.Internal, err)
	}

	defer func() {
		err = r.finishTransaction(err, tx)
	}()

	existingUser, err := r.FindById(ctx, uuid)
	if err != nil {
		if errors.Is(errors.NotExist, err) {
			return errors.E(op, errors.NotExist, err)
		}

		return errors.E(op, errors.Internal, err)
	}

	updatedUser, err := updateFn(ctx, existingUser)
	if err != nil {
		return errors.E(op, errors.Internal, err)
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
	result, err := r.db.NamedExec(query, updated)
	if err != nil {
		return errors.E(op, errors.Internal, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.E(op, errors.Internal, err)
	}

	if rowsAffected == 0 {
		return errors.E(op, errors.NotExist, fmt.Errorf("user with id %s not found", uuid))
	}

	return nil
}

func (r MysqlUserRepository) FindById(ctx context.Context, uuid string) (user.User, error) {
	op := errors.Op("MysqlUserRepository.FindById")

	data := MysqlUser{}
	query := "SELECT * FROM `users` WHERE `id` = ?"

	if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errors.NotExist, err)
		}

		return nil, errors.E(op, errors.Internal, err)
	}

	domainUser, err := user.From(data.ID, data.Role)
	if err != nil {
		return nil, errors.E(op, errors.Internal, err)
	}

	return domainUser, nil
}

func (r MysqlUserRepository) RemoveAll(ctx context.Context) error {
	op := errors.Op("MysqlUserRepository.RemoveAll")

	query := `TRUNCATE TABLE users`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return errors.E(op, errors.Internal, err)
	}

	return nil
}

func (r MysqlUserRepository) finishTransaction(err error, tx *sqlx.Tx) error {
	op := errors.Op("MysqlUserRepository.finishTransaction")

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.E(op, errors.Internal, rollbackErr)
		}

		return errors.E(op, errors.Internal, err)
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			return errors.E(op, errors.Internal, "failed to commit transaction")
		}

		return nil
	}
}
