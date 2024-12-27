// internal/repository/publication/repository.go
package publication

import (
    "context"
    "fmt"
    "time"
    
    "github.com/DukeRupert/haven/internal/model/entity"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

var ErrNotFound = fmt.Errorf("schedule publication not found")

func (r *Repository) Create(ctx context.Context, facilityID int, publishedThrough time.Time) (entity.SchedulePublication, error) {
    var pub entity.SchedulePublication
    err := r.pool.QueryRow(ctx, `
        INSERT INTO schedule_publications (facility_id, published_through)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at, facility_id, published_through
    `, facilityID, publishedThrough).Scan(
        &pub.ID,
        &pub.CreatedAt,
        &pub.UpdatedAt,
        &pub.FacilityID,
        &pub.PublishedThrough,
    )
    if err != nil {
        return pub, fmt.Errorf("error creating schedule publication: %w", err)
    }
    return pub, nil
}

func (r *Repository) Update(ctx context.Context, facilityID int, publishedThrough time.Time) (entity.SchedulePublication, error) {
   var pub entity.SchedulePublication
   err := r.pool.QueryRow(ctx, `
       INSERT INTO schedule_publications (facility_id, published_through)
       VALUES ($1, $2)
       ON CONFLICT (facility_id) DO UPDATE 
       SET published_through = EXCLUDED.published_through,
           updated_at = CURRENT_TIMESTAMP
       RETURNING id, created_at, updated_at, facility_id, published_through
   `, facilityID, publishedThrough).Scan(
       &pub.ID,
       &pub.CreatedAt,
       &pub.UpdatedAt,
       &pub.FacilityID,
       &pub.PublishedThrough,
   )
   if err != nil {
       return pub, fmt.Errorf("error upserting schedule publication: %w", err)
   }
   return pub, nil
}

func (r *Repository) GetByFacilityID(ctx context.Context, facilityID int) (entity.SchedulePublication, error) {
    var pub entity.SchedulePublication
    err := r.pool.QueryRow(ctx, `
        SELECT id, created_at, updated_at, facility_id, published_through
        FROM schedule_publications
        WHERE facility_id = $1
    `, facilityID).Scan(
        &pub.ID,
        &pub.CreatedAt,
        &pub.UpdatedAt,
        &pub.FacilityID,
        &pub.PublishedThrough,
    )
    if err != nil {
        if err == pgx.ErrNoRows {
            return pub, ErrNotFound
        }
        return pub, fmt.Errorf("error getting schedule publication: %w", err)
    }
    return pub, nil
}