package repository

import (
	"app/entity"

	"gorm.io/gorm"
)

type RepositoryUser struct {
	DB *gorm.DB
}

func NewUserPostgres(DB *gorm.DB) *RepositoryUser {
	return &RepositoryUser{DB: DB}
}

func (u *RepositoryUser) GetByID(id int) (user *entity.EntityUser, err error) {
	u.DB.First(&user, id)

	return user, err
}

func (u *RepositoryUser) GetByMail(email string) (user *entity.EntityUser, err error) {
	err = u.DB.Where("email = ?", email).First(&user).Error

	return user, err
}

func (u *RepositoryUser) CreateUser(user *entity.EntityUser) error {

	return u.DB.Create(&user).Error
}

func (u *RepositoryUser) UpdateUser(user *entity.EntityUser) error {

	_, err := u.GetByMail(user.Email)

	if err != nil {
		return err
	}

	return u.DB.Save(&user).Error
}

func (u *RepositoryUser) DeleteUser(user *entity.EntityUser) error {

	_, err := u.GetByMail(user.Email)

	if err != nil {
		return err
	}

	return u.DB.Delete(&user).Error
}

func (u *RepositoryUser) GetUsersFromIDs(ids []int) (users []entity.EntityUser, err error) {
	users = make([]entity.EntityUser, 0)

	err = u.DB.Where("id IN ?", ids).Find(&users).Error

	return users, err
}

func (u *RepositoryUser) GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, err error) {

	users = make([]entity.EntityUser, 0)

	DBFind := u.DB

	if filters.Search != "" {
		DBFind = DBFind.Where("name LIKE ? or email LIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	if filters.Active != "" {
		DBFind = DBFind.Where("active = ?", filters.Active)
	}

	err = DBFind.Find(&users).Error

	return users, err
}

func (u *RepositoryUser) GetUser(id int) (user *entity.EntityUser, err error) {
	u.DB.First(&user, id)

	return user, err
}
