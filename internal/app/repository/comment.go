package repository

import (
	"context"
	"fmt"

	cache "github.com/go-redis/cache/v8"
	"github.com/osamaesmail/go-post-api/internal/app/model"
	"github.com/osamaesmail/go-post-api/internal/config"
	"github.com/osamaesmail/go-post-api/internal/db/mysql"
	"github.com/osamaesmail/go-post-api/internal/db/redis"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	List(ctx context.Context, limit, offset, post_id int) ([]*model.Comment, error)
	Get(ctx context.Context, id int64) (*model.Comment, error)
	Update(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, id int64) error
}

func NewCommentRepository(mysqlClient mysql.Client, redisClient redis.Client) CommentRepository {
	return &commentRepository{mysqlClient, redisClient}
}

type commentRepository struct {
	mysqlClient mysql.Client
	redisClient redis.Client
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	res, err := r.mysqlClient.Conn().ExecContext(ctx, `
	INSERT INTO
		comment (body, account_id, post_id, created_at)
	VALUES
		(?, ?, ?, ?)
	`, comment.Body, comment.AccountID, comment.PostID, comment.CreatedAt)
	if err != nil {
		return err
	}

	comment.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	temp, err := r.Get(ctx, comment.ID)
	*comment = *temp
	return nil
}

func (r *commentRepository) List(ctx context.Context, limit, offset, post_id int) ([]*model.Comment, error) {
	var comments []*model.Comment
	rows, err := r.mysqlClient.Conn().QueryContext(ctx, `
	SELECT
		comment.id, comment.body, comment.created_at, comment.updated_at, comment.account_id, comment.post_id
	FROM comment
	WHERE post_id = ?
	LIMIT ? OFFSET ?`,
	post_id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		comment := new(model.Comment)
		err := rows.Scan(&comment.ID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt, &comment.AccountID, &comment.PostID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *commentRepository) Get(ctx context.Context, id int64) (*model.Comment, error) {
	comment := new(model.Comment)
	err := r.redisClient.Cache().Get(ctx, fmt.Sprintf("comment_%d", id), comment)
	if err != nil && err != cache.ErrCacheMiss {
		return nil, err
	} else if err == nil {
		return comment, nil
	}

	err = r.mysqlClient.Conn().QueryRowContext(ctx, `
	SELECT comment.id, comment.body, comment.created_at, comment.updated_at, comment.account_id, comment.post_id
	FROM comment
	WHERE comment.id = ?
	`, id,
	).Scan(&comment.ID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt, &comment.AccountID, &comment.PostID)
	if err != nil {
		return nil, err
	}

	return comment, r.redisClient.Cache().Set(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("comment_%d", id),
		Value: comment,
		TTL:   config.Cfg().RedisTTL,
	})
}

func (r *commentRepository) Update(ctx context.Context, comment *model.Comment) error {
	_, err := r.mysqlClient.Conn().ExecContext(ctx, `
	UPDATE
		comment
	SET
		body = ?, updated_at = ?
	WHERE
		id = ?
	`, comment.Body, comment.UpdatedAt.Time, comment.ID)
	if err != nil {
		return err
	}

	err = r.redisClient.Cache().Delete(ctx, fmt.Sprintf("comment_%d", comment.ID))
	if err != nil && err != cache.ErrCacheMiss {
		return err
	}

	temp, err := r.Get(ctx, comment.ID)
	*comment = *temp
	return err
}

func (r *commentRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.mysqlClient.Conn().ExecContext(ctx, `
	DELETE FROM
		comment
	WHERE
		id = ?
	`, id)
	if err != nil {
		return err
	}

	err = r.redisClient.Cache().Delete(ctx, fmt.Sprintf("comment_%d", id))
	if err != nil && err != cache.ErrCacheMiss {
		return err
	}

	return nil
}
