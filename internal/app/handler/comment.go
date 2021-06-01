package handler

import (
	"encoding/json"
	"fmt"
	"github.com/osamaesmail/go-post-api/internal/app/model"
	"github.com/osamaesmail/go-post-api/internal/app/service"
	"github.com/osamaesmail/go-post-api/internal/constant"
	"github.com/osamaesmail/go-post-api/internal/validation"
	"github.com/osamaesmail/go-post-api/internal/web"
	"net/http"
	"strconv"
)

type CommentHandler interface {
	Create() http.HandlerFunc
	List() http.HandlerFunc
	Get() http.HandlerFunc
	Update() http.HandlerFunc
	Delete() http.HandlerFunc
}

func NewCommentHandler(commentService service.CommentService) CommentHandler {
	return &commentHandler{commentService}
}

type commentHandler struct {
	commentService service.CommentService
}

// @Router /comments [post]
// @Tags comments
// @Summary Create comment
// @Description TODO
// @Accept json
// @Produce json
// @Param payload body model.CommentCreateRequest true "body request"
// @Success 201 {object} model.CommentResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
func (h *commentHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.CommentCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, constant.ErrRequestBody)
			return
		}

		err = validation.Struct(req)
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, err)
			return
		}

		res, err := h.commentService.Create(r.Context(), req)
		if err != nil {
			switch err {
			case constant.ErrUnauthorized:
				web.MarshalError(w, http.StatusUnauthorized, err)
				return
			default:
				web.MarshalError(w, http.StatusInternalServerError, err)
				return
			}
		}

		web.MarshalPayload(w, http.StatusCreated, res)
	}
}

// @Router /comments [get]
// @Tags comments
// @Summary List comments
// @Description TODO
// @Produce json
// @Param limit query int false "pagination limit"
// @Param offset query int false "pagination offset"
// @Param post_id query int false "post id"
// @Success 200 {array} model.CommentResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
func (h *commentHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := web.GetPagination(r)
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, err)
			return
		}

		postID, err :=  strconv.Atoi(r.URL.Query().Get("post_id"))
		fmt.Println(postID)
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, err)
			return
		}

		req := model.CommentListRequest{
			Limit:  limit,
			Offset: offset,
			PostID:  postID,
		}

		res, err := h.commentService.List(r.Context(), req)
		if err != nil {
			web.MarshalError(w, http.StatusInternalServerError, err)
			return
		}

		web.MarshalPayload(w, http.StatusOK, res)
	}
}

// @Router /comments/{comment_id} [get]
// @Tags comments
// @Summary Get comment
// @Description TODO
// @Accept json
// @Produce json
// @Param comment_id path int true "comment id" Format(int64)
// @Success 200 {object} model.CommentResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
func (h *commentHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := web.GetUrlPathInt64(r, "comment_id")
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, err)
			return
		}

		req := model.CommentGetRequest{ID: id}
		res, err := h.commentService.Get(r.Context(), req)
		if err != nil {
			switch err {
			case constant.ErrCommentNotFound:
				web.MarshalError(w, http.StatusNotFound, err)
				return
			default:
				web.MarshalError(w, http.StatusInternalServerError, err)
				return
			}
		}

		web.MarshalPayload(w, http.StatusOK, res)
	}
}

// @Router /comments/{comment_id} [put]
// @Tags comments
// @Summary Update comment
// @Description TODO
// @Accept json
// @Produce json
// @Param comment_id path int true "comment id" Format(int64)
// @Param payload body model.CommentUpdateRequest true "body request"
// @Success 200 {object} model.CommentResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
func (h *commentHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := web.GetUrlPathInt64(r, "comment_id")
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, err)
			return
		}

		req := model.CommentUpdateRequest{ID: id}
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, constant.ErrRequestBody)
			return
		}

		err = validation.Struct(req)
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, err)
			return
		}

		res, err := h.commentService.Update(r.Context(), req)
		if err != nil {
			switch err {
			case constant.ErrUnauthorized:
				web.MarshalError(w, http.StatusUnauthorized, err)
				return
			case constant.ErrCommentNotFound:
				web.MarshalError(w, http.StatusNotFound, err)
				return
			default:
				web.MarshalError(w, http.StatusInternalServerError, err)
				return
			}
		}

		web.MarshalPayload(w, http.StatusOK, res)
	}
}

// @Router /comments/{comment_id} [delete]
// @Tags comments
// @Summary Delete comment
// @Description TODO
// @Produce json
// @Param comment_id path int true "comment id" Format(int64)
// @Success 204
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
func (h *commentHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := web.GetUrlPathInt64(r, "comment_id")
		if err != nil {
			web.MarshalError(w, http.StatusBadRequest, err)
			return
		}

		req := model.CommentDeleteRequest{ID: id}
		err = h.commentService.Delete(r.Context(), req)
		if err != nil {
			switch err {
			case constant.ErrUnauthorized:
				web.MarshalError(w, http.StatusUnauthorized, err)
				return
			case constant.ErrCommentNotFound:
				web.MarshalError(w, http.StatusNotFound, err)
				return
			default:
				web.MarshalError(w, http.StatusInternalServerError, err)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
