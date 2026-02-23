package adapters

import (
	"clean-arch-go/app/command"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/user"

	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type MysqlUser struct {
	ID           string         `db:"id"`
	Role         string         `db:"role"`
	Name         sql.NullString `db:"name"`
	Email        sql.NullString `db:"email"`
	PasswordHash sql.NullString `db:"password_hash"`
	DeletedAt    sql.NullTime   `db:"deleted_at"`
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
		ID:           u.UUID(),
		Role:         u.Role().String(),
		Name:         sql.NullString{String: u.Name(), Valid: u.Name() != ""},
		Email:        sql.NullString{String: u.Email(), Valid: u.Email() != ""},
		PasswordHash: sql.NullString{String: u.PasswordHash(), Valid: u.PasswordHash() != ""},
	}

	query := `
		INSERT INTO users
			(id, role, name, email, password_hash)
		VALUES
			(:id, :role, :name, :email, :password_hash)
	`

	if _, err := r.db.NamedExecContext(ctx, query, added); err != nil {
		return errors.E(op, errkind.Internal, err)
	}

	return nil
}

func (r MysqlUserRepository) UpdateByID(ctx context.Context, uuid string, updateFn command.UserUpdater) error {
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
		ID:           updatedUser.UUID(),
		Role:         updatedUser.Role().String(),
		Name:         sql.NullString{String: updatedUser.Name(), Valid: updatedUser.Name() != ""},
		Email:        sql.NullString{String: updatedUser.Email(), Valid: updatedUser.Email() != ""},
		PasswordHash: sql.NullString{String: updatedUser.PasswordHash(), Valid: updatedUser.PasswordHash() != ""},
	}

	query := `
		UPDATE users
		SET
			role = :role,
			name = :name,
			email = :email,
			password_hash = :password_hash
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
	query := "SELECT * FROM `users` WHERE `id` = ? AND `deleted_at` IS NULL"

	if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errkind.NotExist, err)
		}

		return nil, errors.E(op, err)
	}

	name := data.Name.String
	email := data.Email.String
	passwordHash := data.PasswordHash.String

	domainUser, err := user.From(data.ID, data.Role, name, email, passwordHash)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return domainUser, nil
}

func (r MysqlUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	op := errors.Op("MysqlUserRepository.FindByEmail")

	data := MysqlUser{}
	query := "SELECT * FROM `users` WHERE `email` = ? AND `deleted_at` IS NULL"

	if err := r.db.GetContext(ctx, &data, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errkind.NotExist, err)
		}

		return nil, errors.E(op, err)
	}

	name := data.Name.String
	emailStr := data.Email.String
	passwordHash := data.PasswordHash.String

	domainUser, err := user.From(data.ID, data.Role, name, emailStr, passwordHash)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return domainUser, nil
}

func (r MysqlUserRepository) FindAll(ctx context.Context) ([]user.User, error) {
	op := errors.Op("MysqlUserRepository.FindAll")

	var data []MysqlUser
	query := "SELECT * FROM `users` WHERE `deleted_at` IS NULL"

	if err := r.db.SelectContext(ctx, &data, query); err != nil {
		return nil, errors.E(op, err)
	}

	domainUsers := make([]user.User, len(data))
	for i, u := range data {
		name := u.Name.String
		emailStr := u.Email.String
		passwordHash := u.PasswordHash.String

		domainUser, err := user.From(u.ID, u.Role, name, emailStr, passwordHash)
		if err != nil {
			return nil, errors.E(op, err)
		}

		domainUsers[i] = domainUser
	}

	return domainUsers, nil
}

func (r MysqlUserRepository) DeleteByID(ctx context.Context, uuid string) error {
	op := errors.Op("MysqlUserRepository.DeleteByID")

	query := "UPDATE `users` SET `deleted_at` = NOW() WHERE `id` = ?"

	result, err := r.db.ExecContext(ctx, query, uuid)
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
