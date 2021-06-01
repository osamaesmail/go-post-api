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
	"golang.org/x/crypto/bcrypt"
)

type AccountService interface {
	Create(ctx context.Context, req model.AccountCreateRequest) (*model.AccountResponse, error)
	List(ctx context.Context, req model.AccountListRequest) ([]*model.AccountResponse, error)
	Get(ctx context.Context, req model.AccountGetRequest) (*model.AccountResponse, error)
	Update(ctx context.Context, req model.AccountUpdateRequest) (*model.AccountResponse, error)
	UpdatePassword(ctx context.Context, req model.AccountPasswordUpdateRequest) (*model.AccountResponse, error)
	Delete(ctx context.Context, req model.AccountDeleteRequest) error
}

func NewAccountService(accountRepository repository.AccountRepository) AccountService {
	return &accountService{accountRepository}
}

type accountService struct {
	accountRepository repository.AccountRepository
}

func (s *accountService) Create(ctx context.Context, req model.AccountCreateRequest) (*model.AccountResponse, error) {
	_, err := s.accountRepository.GetByEmail(ctx, req.Email)
	if err != nil && err != sql.ErrNoRows {
		logger.Log().Err(err).Msg("failed to get account by email")
		return nil, constant.ErrServer
	} else if err == nil {
		return nil, constant.ErrEmailRegistered
	}

	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log().Err(err).Msg("failed to generate from password")
		return nil, constant.ErrServer
	}

	account := &model.Account{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(password),
		CreatedAt: time.Now(),
	}

	err = s.accountRepository.Create(ctx, account)
	if err != nil {
		logger.Log().Err(err).Msg("failed to create account")
		return nil, constant.ErrServer
	}

	return model.NewAccountResponse(account), nil
}

func (s *accountService) List(ctx context.Context, req model.AccountListRequest) ([]*model.AccountResponse, error) {
	accounts, err := s.accountRepository.List(ctx, req.Limit, req.Offset, req.Name)
	if err != nil {
		logger.Log().Err(err).Msg("failed to list accounts")
		return nil, constant.ErrServer
	}

	return model.NewAccountListResponse(accounts), nil
}

func (s *accountService) Get(ctx context.Context, req model.AccountGetRequest) (*model.AccountResponse, error) {
	account, err := s.accountRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, s.switchErrAccountNotFoundOrErrServer(err)
	}

	return model.NewAccountResponse(account), nil
}

func (s *accountService) Update(ctx context.Context, req model.AccountUpdateRequest) (*model.AccountResponse, error) {
	if !middleware.IsMe(ctx, req.ID) {
		return nil, constant.ErrUnauthorized
	}

	account, err := s.accountRepository.GetByEmail(ctx, req.Email)
	if err != nil && err != sql.ErrNoRows {
		logger.Log().Err(err).Msg("failed to get account by email")
		return nil, constant.ErrServer
	} else if err == nil && account.ID != req.ID {
		return nil, constant.ErrEmailRegistered
	}

	account, err = s.accountRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, s.switchErrAccountNotFoundOrErrServer(err)
	}

	account.Name = req.Name
	account.Email = req.Email
	account.UpdatedAt.Time = time.Now()

	err = s.accountRepository.Update(ctx, account)
	if err != nil {
		return nil, s.switchErrAccountNotFoundOrErrServer(err)
	}

	return model.NewAccountResponse(account), nil
}

func (s *accountService) UpdatePassword(ctx context.Context, req model.AccountPasswordUpdateRequest) (*model.AccountResponse, error) {
	if !middleware.IsMe(ctx, req.ID) {
		return nil, constant.ErrUnauthorized
	}

	account, err := s.accountRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, s.switchErrAccountNotFoundOrErrServer(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(req.OldPassword))
	if err != nil {
		return nil, constant.ErrWrongPassword
	}

	password, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Log().Err(err).Msg("failed to generate from password")
		return nil, constant.ErrServer
	}

	account.Password = string(password)
	account.UpdatedAt.Time = time.Now()

	err = s.accountRepository.Update(ctx, account)
	if err != nil {
		return nil, s.switchErrAccountNotFoundOrErrServer(err)
	}

	return model.NewAccountResponse(account), nil
}

func (s *accountService) Delete(ctx context.Context, req model.AccountDeleteRequest) error {
	if !middleware.IsMe(ctx, req.ID) {
		return constant.ErrUnauthorized
	}

	err := s.accountRepository.Delete(ctx, req.ID)
	if err != nil {
		return s.switchErrAccountNotFoundOrErrServer(err)
	}

	return nil
}

func (s *accountService) switchErrAccountNotFoundOrErrServer(err error) error {
	switch err {
	case sql.ErrNoRows:
		return constant.ErrAccountNotFound
	default:
		logger.Log().Err(err).Msg("failed to execute operation account repository")
		return constant.ErrServer
	}
}
