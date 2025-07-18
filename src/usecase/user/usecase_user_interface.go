package usecase_user

import "app/entity"

//go:generate mockgen -destination=../../mocks/mock_usecase_repository_user.go -package=mocks app/usecase/user IRepositoryUser
type IRepositoryUser interface {
	GetByID(id int) (user *entity.EntityUser, err error)
	GetByMail(email string) (user *entity.EntityUser, err error)
	CreateUser(user *entity.EntityUser) error
	UpdateUser(user *entity.EntityUser) error
	DeleteUser(user *entity.EntityUser) error
	GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, err error)
	GetUsersFromIDs(ids []int) (users []entity.EntityUser, err error)
	GetUser(id int) (user *entity.EntityUser, err error)
}

//go:generate mockgen -destination=../../mocks/mock_usecase_user.go -package=mocks app/usecase/user IUsecaseUser
type IUsecaseUser interface {
	LoginUser(email string, password string) (*entity.EntityUser, error)
	Create(user *entity.EntityUser) error
	Update(user *entity.EntityUser) error
	Delete(user *entity.EntityUser) error
	UpdatePassword(id int, oldPassword, newPassword, confirmPassword string) error
	GetUserByToken(token string) (*entity.EntityUser, error)
	GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, err error)
	GetUsersFromIDs(ids []int) (users []entity.EntityUser, err error)
	GetUser(id int) (user *entity.EntityUser, err error)
}
