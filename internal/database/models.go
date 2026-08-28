package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ImageName        string    `json:"image_name"`
	ImageURL         string    `json:"image_url"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	UserID           uuid.UUID `json:"user_id"`
	ImageDescription *string   `json:"image_description"`
	DisplayOrder     int       `json:"display_order"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
}

type Querier interface {
	ListImages() ([]Image, error)
	GetImage(id uuid.UUID) (*Image, error)
	CreateImage(image *Image) (*Image, error)
	UpdateImage(image *Image) (*Image, error)
	DeleteImage(id uuid.UUID) error
	ReorderImages(id uuid.UUID, displayOrder int, updatedAt time.Time) error
	GetUserByEmail(email string) (*User, error)
	CreateUser(user *User) (*User, error)
}

type Queries struct {
	db *sql.DB
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) ListImages() ([]Image, error) {
	rows, err := q.db.Query(`
		SELECT id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order
		FROM images
		ORDER BY display_order ASC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		err := rows.Scan(
			&img.ID, &img.CreatedAt, &img.UpdatedAt,
			&img.ImageName, &img.ImageURL, &img.ThumbnailURL,
			&img.UserID, &img.ImageDescription, &img.DisplayOrder,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

func (q *Queries) GetImage(id uuid.UUID) (*Image, error) {
	var img Image
	err := q.db.QueryRow(`
		SELECT id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order
		FROM images WHERE id = $1
	`, id).Scan(
		&img.ID, &img.CreatedAt, &img.UpdatedAt,
		&img.ImageName, &img.ImageURL, &img.ThumbnailURL,
		&img.UserID, &img.ImageDescription, &img.DisplayOrder,
	)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (q *Queries) CreateImage(img *Image) (*Image, error) {
	err := q.db.QueryRow(`
		INSERT INTO images (id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order
	`, img.ID, img.CreatedAt, img.UpdatedAt, img.ImageName, img.ImageURL, img.ThumbnailURL, img.UserID, img.ImageDescription, img.DisplayOrder).
		Scan(
			&img.ID, &img.CreatedAt, &img.UpdatedAt,
			&img.ImageName, &img.ImageURL, &img.ThumbnailURL,
			&img.UserID, &img.ImageDescription, &img.DisplayOrder,
		)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (q *Queries) UpdateImage(img *Image) (*Image, error) {
	err := q.db.QueryRow(`
		UPDATE images
		SET image_name = $2, image_url = $3, thumbnail_url = $4, image_description = $5, updated_at = $6, display_order = $7
		WHERE id = $1
		RETURNING id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order
	`, img.ID, img.ImageName, img.ImageURL, img.ThumbnailURL, img.ImageDescription, img.UpdatedAt, img.DisplayOrder).
		Scan(
			&img.ID, &img.CreatedAt, &img.UpdatedAt,
			&img.ImageName, &img.ImageURL, &img.ThumbnailURL,
			&img.UserID, &img.ImageDescription, &img.DisplayOrder,
		)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (q *Queries) DeleteImage(id uuid.UUID) error {
	_, err := q.db.Exec(`DELETE FROM images WHERE id = $1`, id)
	return err
}

func (q *Queries) ReorderImages(id uuid.UUID, displayOrder int, updatedAt time.Time) error {
	_, err := q.db.Exec(`UPDATE images SET display_order = $2, updated_at = $3 WHERE id = $1`, id, displayOrder, updatedAt)
	return err
}

func (q *Queries) GetUserByEmail(email string) (*User, error) {
	var user User
	err := q.db.QueryRow(`
		SELECT id, created_at, updated_at, email, password_hash FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (q *Queries) CreateUser(user *User) (*User, error) {
	err := q.db.QueryRow(`
		INSERT INTO users (id, created_at, updated_at, email, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at, email, password_hash
	`, user.ID, user.CreatedAt, user.UpdatedAt, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}
