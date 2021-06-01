package model

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID			int64
	Body		string
	CreatedAt	time.Time
	UpdatedAt	sql.NullTime

	AccountID	int64
	Account		Account

	PostID		int64
	Post		Post
}

type CommentCreateRequest struct {
	Body  	string 	`json:"body" validate:"required"`
	PostID  int64 	`json:"post_id" validate:"required"`
}

type CommentListRequest struct {
	Limit  int
	Offset int
	PostID int
}

type CommentGetRequest struct {
	ID int64
}

type CommentUpdateRequest struct {
	ID    int64  `json:"-"`
	Body  string `json:"body" validate:"required"`
}

type CommentDeleteRequest struct {
	ID int64
}

type CommentResponse struct {
	ID			int64				`json:"id"`
	Body		string				`json:"body"`
	CreatedAt	time.Time			`json:"created_at"`
	UpdatedAt	*time.Time			`json:"updated_at"`

	AccountID	int64           	`json:"account_id"`
	PostID		int64           	`json:"post_id"`
}

func NewCommentResponse(payload *Comment) *CommentResponse {
	res := &CommentResponse{
		ID:        	payload.ID,
		Body:      	payload.Body,
		CreatedAt: 	payload.CreatedAt,
		AccountID: 	payload.AccountID,
		PostID: 	payload.PostID,
	}
	if payload.UpdatedAt.Valid {
		res.UpdatedAt = &payload.UpdatedAt.Time
	}
	return res
}

func NewCommentListResponse(payloads []*Comment) []*CommentResponse {
	res := make([]*CommentResponse, len(payloads))
	for i, payload := range payloads {
		res[i] = NewCommentResponse(payload)
	}
	return res
}
