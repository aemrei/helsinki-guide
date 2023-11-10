package repositories

import (
	"context"
	"errors"
	"time"

	i "github.com/AndreyAD1/helsinki-guide/internal"
	s "github.com/AndreyAD1/helsinki-guide/internal/infrastructure/specifications"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NeighbourhoodRepository interface {
	Add(context.Context, i.Neighbourhood) (*i.Neighbourhood, error)
	Remove(context.Context, i.Neighbourhood) error
	Update(context.Context, i.Neighbourhood) (*i.Neighbourhood, error)
	Query(context.Context, s.Specification) ([]i.Neighbourhood, error)
}

type neighbourhoodStorage struct {
	dbPool *pgxpool.Pool
}

func NewNeighbourhoodRepo(dbPool *pgxpool.Pool) NeighbourhoodRepository {
	return &neighbourhoodStorage{dbPool}
}

func (n *neighbourhoodStorage) Add(
	ctx context.Context, 
	neighbourhood i.Neighbourhood,
) (*i.Neighbourhood, error) {
	query := `INSERT INTO neighbourhoods (name, municipality, created_at)
	VALUES ($1, $2, TIMESTAMP WITH TIME ZONE $3) RETURNING id;`
	created_at := time.Now().Format(time.RFC1123Z)

	var id int64
    err := n.dbPool.QueryRow(
		ctx, 
		query, 
		neighbourhood.Name, 
		neighbourhood.Municipality, 
		created_at,
	).Scan(&id)
    if err != nil {
        var pgxError *pgconn.PgError
        if errors.As(err, &pgxError) {
            if pgxError.Code == pgerrcode.UniqueViolation {
                return nil, ErrDuplicate
            }
        }
        return nil, err
    }
    neighbourhood.ID = id

	return &neighbourhood, nil
}

func (n *neighbourhoodStorage) Remove(ctx context.Context, neighbourhood i.Neighbourhood) error {
	return ErrNotImplemented
}

func (n *neighbourhoodStorage) Update(ctx context.Context, neighbourhood i.Neighbourhood) (*i.Neighbourhood, error) {
	return nil, ErrNotImplemented
}

func (n *neighbourhoodStorage) Query(ctx context.Context, neighbourhood s.Specification) ([]i.Neighbourhood, error) {
	return nil, ErrNotImplemented
}
