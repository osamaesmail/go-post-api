package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/osamaesmail/go-post-api/internal/app/model"
	"github.com/osamaesmail/go-post-api/internal/app/repository"
	"github.com/osamaesmail/go-post-api/internal/constant"
	"github.com/osamaesmail/go-post-api/internal/logger"
	"github.com/osamaesmail/go-post-api/internal/security/middleware"
)

type CommentService interface {
	Create(ctx context.Context, req model.CommentCreateRequest) (*model.CommentResponse, error)
	List(ctx context.Context, req model.CommentListRequest) ([]*model.CommentResponse, error)
	Get(ctx context.Context, req model.CommentGetRequest) (*model.CommentResponse, error)
	Update(ctx context.Context, req model.CommentUpdateRequest) (*model.CommentResponse, error)
	Delete(ctx context.Context, req model.CommentDeleteRequest) error
}

func NewCommentService(commentRepository repository.CommentRepository) CommentService {
	return &commentService{commentRepository}
}

type commentService struct {
	commentRepository repository.CommentRepository
}

func (s *commentService) Create(ctx context.Context, req model.CommentCreateRequest) (*model.CommentResponse, error) {
	claimsID, valid := middleware.GetClaimsID(ctx)
	if !valid {
		return nil, constant.ErrUnauthorized
	}


	comment := &model.Comment{
		Body:      	req.Body,
		CreatedAt: 	time.Now(),
		AccountID: 	claimsID,
		PostID: 	req.PostID,
	}

	err := s.commentRepository.Create(ctx, comment)
	if err != nil {
		logger.Log().Err(err).Msg("failed to create comment")
		return nil, constant.ErrServer
	}

	return model.NewCommentResponse(comment), nil
}

func (s *commentService) List(ctx context.Context, req model.CommentListRequest) ([]*model.CommentResponse, error) {
	comments, err := s.commentRepository.List(ctx, req.Limit, req.Offset, req.PostID)
	if err != nil {
		logger.Log().Err(err).Msg("failed to list comments")
		return nil, constant.ErrServer
	}

	return model.NewCommentListResponse(comments), nil
}

func (s *commentService) Get(ctx context.Context, req model.CommentGetRequest) (*model.CommentResponse, error) {
	comment, err := s.commentRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, s.switchErrCommentNotFoundOrErrServer(err)
	}

	return model.NewCommentResponse(comment), nil
}

func (s *commentService) Update(ctx context.Context, req model.CommentUpdateRequest) (*model.CommentResponse, error) {
	comment, err := s.commentRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, s.switchErrCommentNotFoundOrErrServer(err)
	}

	if !middleware.IsMe(ctx, comment.AccountID) {
		return nil, constant.ErrUnauthorized
	}

	comment.Body = req.Body
	comment.UpdatedAt.Time = time.Now()

	err = s.commentRepository.Update(ctx, comment)
	if err != nil {
		return nil, s.switchErrCommentNotFoundOrErrServer(err)
	}

	return model.NewCommentResponse(comment), nil
}

func (s *commentService) Delete(ctx context.Context, req model.CommentDeleteRequest) error {
	comment, err := s.commentRepository.Get(ctx, req.ID)
	if err != nil {
		return s.switchErrCommentNotFoundOrErrServer(err)
	}

	if !middleware.IsMe(ctx, comment.AccountID) {
		return constant.ErrUnauthorized
	}

	err = s.commentRepository.Delete(ctx, req.ID)
	if err != nil {
		return s.switchErrCommentNotFoundOrErrServer(err)
	}

	return nil
}

func (s *commentService) switchErrCommentNotFoundOrErrServer(err error) error {
	switch err {
	case sql.ErrNoRows:
		return constant.ErrCommentNotFound
	default:
		logger.Log().Err(err).Msg("failed to execute operation post repository")
		return constant.ErrServer
	}
}
