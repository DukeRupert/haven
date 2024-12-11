// internal/model/entity/session.go
package entity

import "time"

// HTTPSession represents a session stored in the database
type HTTPSession struct {
	ID         int64      `db:"id" json:"id"`
	Key        string     `db:"key" json:"key"`
	Data       []byte     `db:"data" json:"data"`
	CreatedOn  time.Time  `db:"created_on" json:"created_on"`
	ModifiedOn *time.Time `db:"modified_on" json:"modified_on,omitempty"`
	ExpiresOn  *time.Time `db:"expires_on" json:"expires_on,omitempty"`
}
