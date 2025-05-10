package repo

import (
	"chalk/internal/repo/models"
	"chalk/pkg/log"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
)

type FilesRepo interface {
}

func NewFilesRepo(db *pgx.Conn, miniocli *minio.Client) FilesRepo {
	return &filesRepo{
		db:       db,
		miniocli: miniocli,
	}
}

type filesRepo struct {
	db       *pgx.Conn
	bucket   string
	miniocli *minio.Client
}

type UploadFileParams struct {
	Reader         io.Reader
	UploaderUserID int64
	Name           string
	ContentType    string
	Size           int64
}

func (r *filesRepo) UploadFile(ctx context.Context, params UploadFileParams) (models.File, error) {
	const checkUserQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var userExists bool
	err := r.db.QueryRow(ctx, checkUserQuery, params.UploaderUserID).Scan(&userExists)
	if err != nil {
		return models.File{}, fmt.Errorf("check user existence: %w", err)
	}
	if !userExists {
		return models.File{}, ErrUserNotFound
	}

	key, err := uuid.NewV7()
	if err != nil {
		return models.File{}, fmt.Errorf("failed to gen file key: %w", err)
	}

	opts := minio.PutObjectOptions{ContentType: params.ContentType}
	uploadInfo, err := r.miniocli.PutObject(ctx, r.bucket, key.String(), params.Reader, params.Size, opts)
	if err != nil {
		return models.File{}, fmt.Errorf("upload file: %w", err)
	}

	var fileID int64
	uploadedAt := time.Now()
	query := `INSERT INTO files 
	(uploader_user_id, name, content_type, bucket, key, uploaded_at, size)
	VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err = r.db.QueryRow(ctx, query,
		params.UploaderUserID,
		params.Name,
		params.ContentType,
		uploadInfo.Bucket,
		uploadInfo.Key,
		uploadedAt.UTC(),
		uploadInfo.Size,
	).Scan(&fileID)
	if err != nil {
		if err := r.miniocli.RemoveObject(ctx, uploadInfo.Bucket, uploadInfo.Key, minio.RemoveObjectOptions{}); err != nil {
			log.Errorf("remove object: %v", err)
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return models.File{}, ErrUserNotFound
			}
		}
		return models.File{}, fmt.Errorf("insert into files: %w", err)
	}
	return models.File{
		ID:             fileID,
		UploaderUserID: params.UploaderUserID,
		Name:           params.Name,
		ContentType:    params.ContentType,
		Bucket:         uploadInfo.Bucket,
		Key:            uploadInfo.Key,
		Size:           uploadInfo.Size,
		UploadedAt:     uploadedAt,
	}, nil
}

func (r *filesRepo) GetFileInfo(ctx context.Context, fileID int64) (models.File, error) {
	file := models.File{}
	const query = `SELECT id, uploader_user_id, name, content_type, bucket, key, uploaded_at, size WHERE id = $1`
	err := r.db.QueryRow(ctx, query, fileID).Scan(
		file.ID,
		file.UploaderUserID,
		file.Name,
		file.ContentType,
		file.Bucket,
		file.Key,
		file.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.File{}, ErrFileNotFound
		}
		return models.File{}, fmt.Errorf("select file: %v", err)
	}
	return file, nil
}

func (r *filesRepo) GetFileByID(ctx context.Context, fileID int64) (io.Reader, error) {
	fi, err := r.GetFileInfo(ctx, fileID)
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("get file info: %v", err)
	}
	opts := minio.GetObjectOptions{}
	obj, err := r.miniocli.GetObject(ctx, fi.Bucket, fi.Key, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %v", err)
	}

	return obj, nil
}

// func (r *filesRepo) GetFileByKey(ctx context.Context, fileID int64) (io.Reader, error) {
// 	fi, err := r.GetFileInfo(ctx, fileID)
// 	if err != nil {
// 		if errors.Is(err, ErrFileNotFound) {
// 			return nil, ErrFileNotFound
// 		}
// 		return nil, fmt.Errorf("get file info: %v", err)
// 	}
// 	opts := minio.GetObjectOptions{}
// 	obj, err := r.miniocli.GetObject(ctx, fi.Bucket, fi.Key, opts)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get object: %v", err)
// 	}

// 	return obj, nil
// }
