package usecase_user

import (
	"app/config"
	"app/entity"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

const SECRET_KEY = "9an0afx$thw)k9#y*_d9-ch^r&a6ndi#x#dwu^52zbqw=hso(9"

type SignedDetails struct {
	ID    int
	Name  string
	Email string
	jwt.StandardClaims
}

type UseCaseUser struct {
	repo IRepositoryUser
}

func NewService(repository IRepositoryUser) *UseCaseUser {
	return &UseCaseUser{repo: repository}
}

func (u *UseCaseUser) LoginUser(email string, password string) (*entity.EntityUser, error) {
	user, err := u.repo.GetByMail(email)

	if err != nil {
		return nil, err
	}

	err = user.ValidatePassword(password)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UseCaseUser) Create(user *entity.EntityUser) error {

	err := user.GetValidated()

	if err != nil {
		return err
	}

	return u.repo.CreateUser(user)
}

func (u *UseCaseUser) Update(user *entity.EntityUser) error {
	return u.repo.UpdateUser(user)
}

func (u *UseCaseUser) Delete(user *entity.EntityUser) error {
	return u.repo.DeleteUser(user)
}

func (u *UseCaseUser) GetUserByToken(token string) (*entity.EntityUser, error) {
	claims, err := ValidateToken(token)

	if err != nil {
		return nil, err
	}

	user, err := u.repo.GetByID(claims.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UseCaseUser) UpdatePassword(id int, oldPassword, newPassword, confirmPassword string) error {

	user, err := u.repo.GetByID(id)

	if err != nil {
		return err
	}

	err = user.ValidatePassword(oldPassword)

	if err != nil {
		return err
	}

	if newPassword != confirmPassword {
		return errors.New("passwords do not match")
	}

	user.UpdatePassword(newPassword)

	err = user.GetValidated()

	if err != nil {
		return err
	}

	err = u.repo.UpdateUser(user)

	return err
}

func (u *UseCaseUser) GetUsersFromIDs(ids []int) (users []entity.EntityUser, err error) {
	return u.repo.GetUsersFromIDs(ids)
}
func (u *UseCaseUser) GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, err error) {
	return u.repo.GetUsers(filters)
}

func (u *UseCaseUser) GetUser(id int) (user *entity.EntityUser, err error) {
	return u.repo.GetUser(id)
}

func JWTTokenGenerator(u entity.EntityUser) (signedToken string, signedRefreshToken string, err error) {

	claims := SignedDetails{
		ID:    int(u.ID),
		Name:  u.Name,
		Email: u.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	refreshClaims := SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7 * 365).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func ValidateToken(signedToken string) (claims *SignedDetails, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {

		return nil, err
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {

		return nil, err
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {

		return nil, err
	}

	return claims, nil
}

func (u *UseCaseUser) CreateAdminUser() error {
	user, err := entity.NewUser(entity.EntityUser{
		Name:     "Admin",
		Email:    config.EnvironmentVariables.DEFAULT_ADMIN_MAIL,
		Password: config.EnvironmentVariables.DEFAULT_ADMIN_PASSWORD,
		IsAdmin:  true,
		Active:   true,
	})

	if err != nil {
		return err
	}

	err = user.GetValidated()

	if err != nil {
		return err
	}

	// check if user already exists
	_, err = u.repo.GetByMail(user.Email)

	if err == nil {
		return err
	}

	return u.repo.CreateUser(user)
}
