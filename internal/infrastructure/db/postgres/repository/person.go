package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"github.com/jmoiron/sqlx"
)

// peopleDB é o modelo de banco de dados (Data Model).
//
// Por que ter um modelo separado da entidade de domínio?
//   - Nomes de colunas podem ser diferentes dos campos da entidade
//   - Permite adicionar campos específicos do banco (audit, soft delete)
//   - Desacopla mudanças de schema de mudanças de domínio
//
// As tags `db` são usadas pelo SQLx para fazer o mapeamento automático.
type peopleDB struct {
	ID           string         `db:"id"`
	Name         string         `db:"name"`
	Document     string         `db:"document"`
	Phone        sql.NullString `db:"phone"`
	Email        string         `db:"email"`
	Street       sql.NullString `db:"street"`
	Number       sql.NullString `db:"number"`
	Complement   sql.NullString `db:"complement"`
	Neighborhood sql.NullString `db:"neighborhood"`
	City         sql.NullString `db:"city"`
	State        sql.NullString `db:"state"`
	ZipCode      sql.NullString `db:"zip_code"`
	Active       bool           `db:"active"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

func (c *peopleDB) toEntity() (*person.Person, error) {
	id, err := vo.ParseID(c.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear ID: %w", err)
	}

	cpf, err := vo.NewCPF(c.Document)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear CPF: %w", err)
	}

	email, err := vo.NewEmail(c.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear email: %w", err)
	}

	return &person.Person{
		ID:    id,
		Name:  c.Name,
		CPF:   cpf,
		Phone: vo.ParsePhone(c.Phone.String),
		Email: email,
		Address: vo.Address{
			Street:       c.Street.String,
			Number:       c.Number.String,
			Complement:   c.Complement.String,
			Neighborhood: c.Neighborhood.String,
			City:         c.City.String,
			State:        c.State.String,
			ZipCode:      c.ZipCode.String,
		},
		Active:    c.Active,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}, nil
}

func fromEntity(c *person.Person) peopleDB {
	return peopleDB{
		ID:           c.ID.String(),
		Name:         c.Name,
		Document:     c.CPF.String(),
		Phone:        toNullString(c.Phone.String()),
		Email:        c.Email.String(),
		Street:       toNullString(c.Address.Street),
		Number:       toNullString(c.Address.Number),
		Complement:   toNullString(c.Address.Complement),
		Neighborhood: toNullString(c.Address.Neighborhood),
		City:         toNullString(c.Address.City),
		State:        toNullString(c.Address.State),
		ZipCode:      toNullString(c.Address.ZipCode),
		Active:       c.Active,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

// toNullString converte string para sql.NullString.
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

type PersonRepository struct {
	DB *sqlx.DB
}

func (r *PersonRepository) Create(ctx context.Context, c *person.Person) error {
	query := `
		INSERT INTO people (
			id, name, document, phone, email, 
			street, number, complement, neighborhood, city, state, zip_code,
			active, created_at, updated_at
		) VALUES (
			:id, :name, :document, :phone, :email,
			:street, :number, :complement, :neighborhood, :city, :state, :zip_code,
			:active, :created_at, :updated_at
		)
	`

	dbModel := fromEntity(c)
	_, err := r.DB.NamedExecContext(ctx, query, dbModel)
	return err
}

func (r *PersonRepository) FindByID(ctx context.Context, id vo.ID) (*person.Person, error) {
	query := `
		SELECT id, name, document, phone, email,
			   street, number, complement, neighborhood, city, state, zip_code,
			   active, created_at, updated_at
		FROM people
		WHERE id = $1
	`

	var dbModel peopleDB
	err := r.DB.GetContext(ctx, &dbModel, query, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, person.ErrPersonNotFound
		}
		return nil, err
	}

	return dbModel.toEntity()
}

func (r *PersonRepository) FindByCPF(ctx context.Context, cpf vo.CPF) (*person.Person, error) {
	query := `
		SELECT id, name, document, phone, email,
			   street, number, complement, neighborhood, city, state, zip_code,
			   active, created_at, updated_at
		FROM people
		WHERE document = $1
	`

	var dbModel peopleDB
	err := r.DB.GetContext(ctx, &dbModel, query, cpf.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, person.ErrPersonNotFound
		}
		return nil, err
	}

	return dbModel.toEntity()
}

func (r *PersonRepository) List(ctx context.Context, filter person.ListFilter) (*person.ListResult, error) {
	filter.Normalize()

	// Construir query dinâmica com filtros
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
	if filter.City != "" {
		conditions = append(conditions, "city ILIKE :city")
		args["city"] = "%" + filter.City + "%"
	}
	if filter.State != "" {
		conditions = append(conditions, "state = :state")
		args["state"] = strings.ToUpper(filter.State)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Query para contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM people %s", whereClause)
	var total int

	// Usar NamedQuery para substituir os named params
	countQuery, countArgs, err := sqlx.Named(countQuery, args)
	if err != nil {
		return nil, err
	}
	countQuery = r.DB.Rebind(countQuery)

	err = r.DB.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, err
	}

	// Query para buscar dados paginados
	args["limit"] = filter.Limit
	args["offset"] = filter.Offset()

	dataQuery := fmt.Sprintf(`
		SELECT id, name, document, phone, email,
			   street, number, complement, neighborhood, city, state, zip_code,
			   active, created_at, updated_at
		FROM people
		%s
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`, whereClause)

	dataQuery, dataArgs, err := sqlx.Named(dataQuery, args)
	if err != nil {
		return nil, err
	}
	dataQuery = r.DB.Rebind(dataQuery)

	var dbModels []peopleDB
	err = r.DB.SelectContext(ctx, &dbModels, dataQuery, dataArgs...)
	if err != nil {
		return nil, err
	}

	// Converter para entidades de domínio
	people := make([]*person.Person, 0, len(dbModels))
	for i := range dbModels {
		entity, err := dbModels[i].toEntity()
		if err != nil {
			return nil, err
		}
		people = append(people, entity)
	}

	return &person.ListResult{
		Persons: people,
		Total:   total,
		Page:    filter.Page,
		Limit:   filter.Limit,
	}, nil
}

func (r *PersonRepository) Update(ctx context.Context, c *person.Person) error {
	query := `
		UPDATE people SET
			name = :name,
			phone = :phone,
			email = :email,
			street = :street,
			number = :number,
			complement = :complement,
			neighborhood = :neighborhood,
			city = :city,
			state = :state,
			zip_code = :zip_code,
			active = :active,
			updated_at = :updated_at
		WHERE id = :id
	`

	dbModel := fromEntity(c)
	result, err := r.DB.NamedExecContext(ctx, query, dbModel)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return person.ErrPersonNotFound
	}

	return nil
}

func (r *PersonRepository) Delete(ctx context.Context, id vo.ID) error {
	query := `
		UPDATE people SET
			active = false,
			updated_at = $1
		WHERE id = $2 AND active = true
	`

	result, err := r.DB.ExecContext(ctx, query, time.Now(), id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return person.ErrPersonNotFound
	}

	return nil
}
