package entity_test

import (
	"app/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestEntityUser_ValidatePassword(t *testing.T) {
// 	u := entity.EntityUser{
// 		Password: "passwordTest",
// 	}

// 	user, err := entity.NewUser(u)

// 	assert.Nil(t, err)
// 	assert.NotEqual(t, user.Password, u.Password)

// 	err = user.ValidatePassword(u.Password)

// 	assert.Nil(t, err)
// }

func TestEntityUser_ValidatedSuccess(t *testing.T) {

	arg := entity.EntityUser{
		Name:     "Name",
		Email:    "email@email.com",
		Password: "Password",
	}

	user, err := entity.NewUser(arg)
	assert.Nil(t, err)

	err = user.Validate()
	assert.Nil(t, err)

}

func TestEntityUser_ValidatedFail(t *testing.T) {

	arg := entity.EntityUser{
		Name:     "",
		Email:    "",
		Password: "",
	}

	user, err := entity.NewUser(arg)
	assert.Nil(t, err)

	err = user.GetValidated()
	assert.NotNil(t, err)

}
