package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	entity "bitbucket.org/appmax-space/go-boilerplate/internal/domain/entity_example"
	"bitbucket.org/appmax-space/go-boilerplate/internal/domain/entity_example/vo"
	"github.com/jmoiron/sqlx"
)

// entityDB é o modelo de banco de dados (Data Model).
//
// Por que ter um modelo separado da entidade de domínio?
//   - Nomes de colunas podem ser diferentes dos campos da entidade
//   - Permite adicionar campos específicos do banco (audit, soft delete)
//   - Desacopla mudanças de schema de mudanças de domínio
type entityDB struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (e *entityDB) toEntity() (*entity.Entity, error) {
	id, parseErr := vo.ParseID(e.ID)
	if parseErr != nil {
		return nil, fmt.Errorf("parsing ID: %w", parseErr)
	}

	email, emailErr := vo.NewEmail(e.Email)
	if emailErr != nil {
		return nil, fmt.Errorf("parsing email: %w", emailErr)
	}

	return &entity.Entity{
		ID:        id,
		Name:      e.Name,
		Email:     email,
		Active:    e.Active,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}, nil
}

func fromDomainEntity(e *entity.Entity) entityDB {
	return entityDB{
		ID:        e.ID.String(),
		Name:      e.Name,
		Email:     e.Email.String(),
		Active:    e.Active,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// EntityRepository implementa a interface Repository para Entity.
type EntityRepository struct {
	writer *sqlx.DB
	reader *sqlx.DB
}

// NewEntityRepository cria uma nova instância do repositório.
func NewEntityRepository(writer, reader *sqlx.DB) *EntityRepository {
	return &EntityRepository{writer: writer, reader: reader}
}

func (r *EntityRepository) Create(ctx context.Context, e *entity.Entity) error {
	query := `
		INSERT INTO entities (
			id, name, email, active, created_at, updated_at
		) VALUES (
			:id, :name, :email, :active, :created_at, :updated_at
		)
	`

	dbModel := fromDomainEntity(e)
	_, execErr := r.writer.NamedExecContext(ctx, query, dbModel)
	return execErr
}

func (r *EntityRepository) FindByID(ctx context.Context, id vo.ID) (*entity.Entity, error) {
	query := `
		SELECT id, name, email, active, created_at, updated_at
		FROM entities
		WHERE id = $1
	`

	var dbModel entityDB
	selectErr := r.reader.GetContext(ctx, &dbModel, query, id.String())
	if selectErr != nil {
		if errors.Is(selectErr, sql.ErrNoRows) {
			return nil, entity.ErrEntityNotFound
		}
		return nil, selectErr
	}

	return dbModel.toEntity()
}

func (r *EntityRepository) FindByEmail(ctx context.Context, email vo.Email) (*entity.Entity, error) {
	query := `
		SELECT id, name, email, active, created_at, updated_at
		FROM entities
		WHERE email = $1
	`

	var dbModel entityDB
	selectErr := r.reader.GetContext(ctx, &dbModel, query, email.String())
	if selectErr != nil {
		if errors.Is(selectErr, sql.ErrNoRows) {
			return nil, entity.ErrEntityNotFound
		}
		return nil, selectErr
	}

	return dbModel.toEntity()
}

func (r *EntityRepository) List(ctx context.Context, filter entity.ListFilter) (*entity.ListResult, error) {
	filter.Normalize()

	// Build dynamic query with filters
	var conditions []string
	args := make(map[string]interface{})

	if filter.ActiveOnly {
		conditions = append(conditions, "active = true")
	}
	if filter.Name != "" {
		conditions = append(conditions, "name ILIKE :name")
		args["name"] = "%" + filter.Name + "%"
	}
	if filter.Email != "" {
		conditions = append(conditions, "email ILIKE :email")
		args["email"] = "%" + filter.Email + "%"
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Wrap COUNT + SELECT in a read-only transaction for consistent pagination.
	// Without a transaction, rows could be inserted/deleted between the two queries,
	// causing total count to be inconsistent with the returned data.
	tx, txErr := r.reader.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if txErr != nil {
		return nil, fmt.Errorf("beginning read transaction: %w", txErr)
	}
	defer func() { _ = tx.Rollback() }()

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM entities %s", whereClause)
	var total int

	countQuery, countArgs, namedErr := sqlx.Named(countQuery, args)
	if namedErr != nil {
		return nil, namedErr
	}
	countQuery = tx.Rebind(countQuery)

	countErr := tx.GetContext(ctx, &total, countQuery, countArgs...)
	if countErr != nil {
		return nil, countErr
	}

	// Paginated data query
	args["limit"] = filter.Limit
	args["offset"] = filter.Offset()

	dataQuery := fmt.Sprintf(`
		SELECT id, name, email, active, created_at, updated_at
		FROM entities
		%s
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`, whereClause)

	dataQuery, dataArgs, dataNamedErr := sqlx.Named(dataQuery, args)
	if dataNamedErr != nil {
		return nil, dataNamedErr
	}
	dataQuery = tx.Rebind(dataQuery)

	var dbModels []entityDB
	selectErr := tx.SelectContext(ctx, &dbModels, dataQuery, dataArgs...)
	if selectErr != nil {
		return nil, selectErr
	}

	// Commit the read-only transaction (also valid to let defer Rollback handle it,
	// but explicit commit is cleaner for read-only transactions).
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("committing read transaction: %w", commitErr)
	}

	// Convert to domain entities
	entities := make([]*entity.Entity, 0, len(dbModels))
	for i := range dbModels {
		e, convertErr := dbModels[i].toEntity()
		if convertErr != nil {
			return nil, convertErr
		}
		entities = append(entities, e)
	}

	return &entity.ListResult{
		Entities: entities,
		Total:    total,
		Page:     filter.Page,
		Limit:    filter.Limit,
	}, nil
}

func (r *EntityRepository) Update(ctx context.Context, e *entity.Entity) error {
	tx, txErr := r.writer.BeginTxx(ctx, nil)
	if txErr != nil {
		return fmt.Errorf("beginning transaction: %w", txErr)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		UPDATE entities SET
			name = :name,
			email = :email,
			active = :active,
			updated_at = :updated_at
		WHERE id = :id
	`

	dbModel := fromDomainEntity(e)
	result, execErr := tx.NamedExecContext(ctx, query, dbModel)
	if execErr != nil {
		return execErr
	}

	rowsAffected, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		return rowsErr
	}

	if rowsAffected == 0 {
		return entity.ErrEntityNotFound
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}

	return nil
}

func (r *EntityRepository) Delete(ctx context.Context, id vo.ID) error {
	query := `
		UPDATE entities SET
			active = false,
			updated_at = $1
		WHERE id = $2 AND active = true
	`

	result, execErr := r.writer.ExecContext(ctx, query, time.Now(), id.String())
	if execErr != nil {
		return execErr
	}

	rowsAffected, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		return rowsErr
	}

	if rowsAffected == 0 {
		return entity.ErrEntityNotFound
	}

	return nil
}
