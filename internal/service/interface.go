package service

import (
	"go-short/internal/repository"

	"gorm.io/gorm"
)

type LinkService struct {
	db                  *gorm.DB
	linkRepository      repository.LinkRepository
	userRepository      repository.UserRepository
	accessLogRepository repository.AccessLogRepository
}

func NewLinkService(db *gorm.DB, linkRepository repository.LinkRepository, userRepository repository.UserRepository, accessLogRepository repository.AccessLogRepository) *LinkService {
	return &LinkService{
		db:                  db,
		linkRepository:      linkRepository,
		userRepository:      userRepository,
		accessLogRepository: accessLogRepository,
	}
}

type AdminService struct {
	db                  *gorm.DB
	linkRepository      repository.LinkRepository
	userRepository      repository.UserRepository
	accessLogRepository repository.AccessLogRepository
}

func NewAdminService(db *gorm.DB, linkRepository repository.LinkRepository, userRepository repository.UserRepository, accessLogRepository repository.AccessLogRepository) *AdminService {
	return &AdminService{
		db:                  db,
		linkRepository:      linkRepository,
		userRepository:      userRepository,
		accessLogRepository: accessLogRepository,
	}
}

type UserService struct {
	db             *gorm.DB
	userRepository repository.UserRepository
}

func NewUserService(db *gorm.DB, userRepository repository.UserRepository) *UserService {
	return &UserService{
		db:             db,
		userRepository: userRepository,
	}
}
