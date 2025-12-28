package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewPostgresStore creates a new PostgreSQL connection with connection pooling
func NewPostgresStore(host, port, user, password, dbname string) (*PostgresStore, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

// Close closes the database connection
func (p *PostgresStore) Close() error {
	return p.db.Close()
}

// DB returns the underlying *sql.DB for advanced queries when necessary
func (p *PostgresStore) DB() *sql.DB {
	return p.db
}

// VerifyPersonalAccessToken verifies a raw personal access token against stored bcrypt hashes.
// Returns tokenID and userID on success, or sql.ErrNoRows if not found.
func (p *PostgresStore) VerifyPersonalAccessToken(ctx context.Context, rawToken string) (string, string, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, user_id, token_hash FROM personal_access_tokens WHERE expires_at IS NULL OR expires_at > NOW()`)
	if err != nil {
		log.Printf("[store] VerifyPAT query error: %v", err)
		return "", "", err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		var id string
		var uid string
		var thash string
		if err := rows.Scan(&id, &uid, &thash); err != nil {
			log.Printf("[store] VerifyPAT scan error: %v", err)
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(thash), []byte(rawToken)) == nil {
			// update last_used_at (best-effort)
			if _, err := p.db.ExecContext(ctx, `UPDATE personal_access_tokens SET last_used_at = $1 WHERE id = $2`, time.Now().UTC(), id); err != nil {
				log.Printf("[store] failed to update last_used_at for id=%s: %v", id, err)
			}
			log.Printf("[store] VerifyPAT matched id=%s user=%s (scanned=%d)", id, uid, count)
			return id, uid, nil
		}
	}
	log.Printf("[store] VerifyPAT no match (scanned=%d)", count)
	return "", "", sql.ErrNoRows
}

// =====================================================
// USER OPERATIONS
// =====================================================

// CreateUser creates a new user in the database
func (p *PostgresStore) CreateUser(ctx context.Context, username, email, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, password_hash, role, created_at, updated_at
	`

	var user User
	err := p.db.QueryRowContext(ctx, query, username, email, passwordHash).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (p *PostgresStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user User
	err := p.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (p *PostgresStore) GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := p.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UserExists checks if a user exists by username
func (p *PostgresStore) UserExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := p.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// UpdateUserPassword updates a user's password
func (p *PostgresStore) UpdateUserPassword(ctx context.Context, userID, newPasswordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := p.db.ExecContext(ctx, query, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}

	return nil
}

// =====================================================
// FILE OPERATIONS
// =====================================================

// SaveFileMetadata saves file metadata to the database
func (p *PostgresStore) SaveFileMetadata(ctx context.Context, metadata *FileMetadata) error {
	log.Printf("[DEBUG] SaveFileMetadata: FileID=%s, UserID=%s, FileName=%s, Tags=%v",
		metadata.FileID, metadata.UserID, metadata.FileName, metadata.Tags)

	query := `
		INSERT INTO files (
			id, user_id, file_name, description, mime_type, 
			size, encrypted_size, minio_path, encryption_key, 
			created_at, expires_at, download_count, tags
		) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := p.db.ExecContext(ctx, query,
		metadata.FileID,
		metadata.UserID,
		metadata.FileName,
		metadata.Description,
		metadata.MimeType,
		metadata.Size,
		metadata.EncryptedSize,
		metadata.MinIOPath,
		metadata.EncryptionKey,
		metadata.CreatedAt,
		metadata.ExpiresAt,
		metadata.DownloadCount,
		pq.Array(metadata.Tags),
	)

	if err != nil {
		log.Printf("[ERROR] Failed to save file metadata: %v", err)
		return fmt.Errorf("failed to save file metadata: %w", err)
	}

	log.Printf("[DEBUG] Successfully saved file metadata: FileID=%s", metadata.FileID)

	return nil
}

// GetFileMetadata retrieves file metadata by file ID
func (p *PostgresStore) GetFileMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	query := `
		SELECT id, user_id, file_name, description, mime_type,
		       size, encrypted_size, minio_path, encryption_key,
		       created_at, expires_at, download_count, tags
		FROM files
		WHERE id = $1
	`

	var metadata FileMetadata
	var description sql.NullString
	var expiresAt sql.NullTime

	err := p.db.QueryRowContext(ctx, query, fileID).Scan(
		&metadata.FileID,
		&metadata.UserID,
		&metadata.FileName,
		&description,
		&metadata.MimeType,
		&metadata.Size,
		&metadata.EncryptedSize,
		&metadata.MinIOPath,
		&metadata.EncryptionKey,
		&metadata.CreatedAt,
		&expiresAt,
		&metadata.DownloadCount,
		pq.Array(&metadata.Tags),
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		metadata.Description = description.String
	}
	if expiresAt.Valid {
		metadata.ExpiresAt = &expiresAt.Time
	}

	return &metadata, nil
}

// UpdateFileMetadata updates file metadata (for description/tags changes)
func (p *PostgresStore) UpdateFileMetadata(ctx context.Context, fileID, description string, tags []string) error {
	query := `
		UPDATE files
		SET description = $1, tags = $2
		WHERE id = $3
	`

	result, err := p.db.ExecContext(ctx, query, description, pq.Array(tags), fileID)
	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("file not found: %s", fileID)
	}

	return nil
}

// ListUserFiles retrieves all files for a user
func (p *PostgresStore) ListUserFiles(ctx context.Context, userID string) ([]*FileMetadata, error) {
	query := `
		SELECT id, user_id, file_name, description, mime_type,
		       size, encrypted_size, minio_path, encryption_key,
		       created_at, expires_at, download_count, tags
		FROM files
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	var files []*FileMetadata
	for rows.Next() {
		var metadata FileMetadata
		var description sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(
			&metadata.FileID,
			&metadata.UserID,
			&metadata.FileName,
			&description,
			&metadata.MimeType,
			&metadata.Size,
			&metadata.EncryptedSize,
			&metadata.MinIOPath,
			&metadata.EncryptionKey,
			&metadata.CreatedAt,
			&expiresAt,
			&metadata.DownloadCount,
			pq.Array(&metadata.Tags),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			metadata.Description = description.String
		}
		if expiresAt.Valid {
			metadata.ExpiresAt = &expiresAt.Time
		}

		files = append(files, &metadata)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating files: %w", err)
	}

	return files, nil
}

// SearchFiles searches files by filename or tags
func (p *PostgresStore) SearchFiles(ctx context.Context, userID, query string) ([]*FileMetadata, error) {
	sqlQuery := `
		SELECT id, user_id, file_name, description, mime_type,
		       size, encrypted_size, minio_path, encryption_key,
		       created_at, expires_at, download_count, tags
		FROM files
		WHERE user_id = $1 AND (
			file_name ILIKE $2 OR
			description ILIKE $2 OR
			$3 = ANY(tags)
		)
		ORDER BY created_at DESC
	`

	searchPattern := "%" + query + "%"
	rows, err := p.db.QueryContext(ctx, sqlQuery, userID, searchPattern, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}
	defer rows.Close()

	var files []*FileMetadata
	for rows.Next() {
		var metadata FileMetadata
		var description sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(
			&metadata.FileID,
			&metadata.UserID,
			&metadata.FileName,
			&description,
			&metadata.MimeType,
			&metadata.Size,
			&metadata.EncryptedSize,
			&metadata.MinIOPath,
			&metadata.EncryptionKey,
			&metadata.CreatedAt,
			&expiresAt,
			&metadata.DownloadCount,
			pq.Array(&metadata.Tags),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			metadata.Description = description.String
		}
		if expiresAt.Valid {
			metadata.ExpiresAt = &expiresAt.Time
		}

		files = append(files, &metadata)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating files: %w", err)
	}

	return files, nil
}

// DeleteFileMetadata deletes file metadata
func (p *PostgresStore) DeleteFileMetadata(ctx context.Context, fileID string) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("file not found: %s", fileID)
	}

	return nil
}

// IncrementDownloadCount increments the download counter for a file
func (p *PostgresStore) IncrementDownloadCount(ctx context.Context, fileID string) error {
	query := `
		UPDATE files
		SET download_count = download_count + 1
		WHERE id = $1
	`

	_, err := p.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to increment download count: %w", err)
	}

	return nil
}

// GetExpiredFiles retrieves all files that have expired
func (p *PostgresStore) GetExpiredFiles(ctx context.Context) ([]*FileMetadata, error) {
	query := `
		SELECT id, user_id, file_name, description, mime_type,
		       size, encrypted_size, minio_path, encryption_key,
		       created_at, expires_at, download_count, tags
		FROM files
		WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP
		ORDER BY expires_at ASC
	`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired files: %w", err)
	}
	defer rows.Close()

	var files []*FileMetadata
	for rows.Next() {
		var metadata FileMetadata
		var description sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(
			&metadata.FileID,
			&metadata.UserID,
			&metadata.FileName,
			&description,
			&metadata.MimeType,
			&metadata.Size,
			&metadata.EncryptedSize,
			&metadata.MinIOPath,
			&metadata.EncryptionKey,
			&metadata.CreatedAt,
			&expiresAt,
			&metadata.DownloadCount,
			pq.Array(&metadata.Tags),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			metadata.Description = description.String
		}
		if expiresAt.Valid {
			metadata.ExpiresAt = &expiresAt.Time
		}

		files = append(files, &metadata)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating files: %w", err)
	}

	return files, nil
}
