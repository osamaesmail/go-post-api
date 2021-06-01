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

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	List(ctx context.Context, limit, offset int, title string) ([]*model.Post, error)
	Get(ctx context.Context, id int64) (*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id int64) error
}

func NewPostRepository(mysqlClient mysql.Client, redisClient redis.Client) PostRepository {
	return &postRepository{mysqlClient, redisClient}
}

type postRepository struct {
	mysqlClient mysql.Client
	redisClient redis.Client
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	res, err := r.mysqlClient.Conn().ExecContext(ctx, `
	INSERT INTO
		post (title, body, account_id, created_at)
	VALUES
		(?, ?, ?, ?)
	`, post.Title, post.Body, post.AccountID, post.CreatedAt)
	if err != nil {
		return err
	}

	post.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	temp, err := r.Get(ctx, post.ID)
	*post = *temp
	return nil
}

func (r *postRepository) List(ctx context.Context, limit, offset int, title string) ([]*model.Post, error) {
	var posts []*model.Post
	rows, err := r.mysqlClient.Conn().QueryContext(ctx, `
	SELECT post.id, post.title, post.body, post.created_at, post.updated_at, post.account_id
	FROM post WHERE post.title LIKE ? LIMIT ? OFFSET ?`, "%"+title+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := new(model.Post)
		err := rows.Scan(&post.ID, &post.Title, &post.Body, &post.CreatedAt, &post.UpdatedAt, &post.AccountID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *postRepository) Get(ctx context.Context, id int64) (*model.Post, error) {
	post := new(model.Post)
	err := r.redisClient.Cache().Get(ctx, fmt.Sprintf("post_%d", id), post)
	if err != nil && err != cache.ErrCacheMiss {
		return nil, err
	} else if err == nil {
		return post, nil
	}

	err = r.mysqlClient.Conn().QueryRowContext(ctx, `
	SELECT post.id, post.title, post.body, post.created_at, post.updated_at, post.account_id
	FROM post WHERE post.id = ?`, id).
		Scan(&post.ID, &post.Title, &post.Body, &post.CreatedAt, &post.UpdatedAt, &post.AccountID)
	if err != nil {
		return nil, err
	}

	return post, r.redisClient.Cache().Set(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("post_%d", id),
		Value: post,
		TTL:   config.Cfg().RedisTTL,
	})
}

func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	_, err := r.mysqlClient.Conn().ExecContext(ctx, `
	UPDATE
		post
	SET
		title = ?, body = ?, updated_at = ?
	WHERE
		id = ?
	`, post.Title, post.Body, post.UpdatedAt.Time, post.ID)
	if err != nil {
		return err
	}

	err = r.redisClient.Cache().Delete(ctx, fmt.Sprintf("post_%d", post.ID))
	if err != nil && err != cache.ErrCacheMiss {
		return err
	}

	temp, err := r.Get(ctx, post.ID)
	*post = *temp
	return err
}

func (r *postRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.mysqlClient.Conn().ExecContext(ctx, `
	DELETE FROM
		post
	WHERE
		id = ?
	`, id)
	if err != nil {
		return err
	}

	err = r.redisClient.Cache().Delete(ctx, fmt.Sprintf("post_%d", id))
	if err != nil && err != cache.ErrCacheMiss {
		return err
	}

	return nil
}
