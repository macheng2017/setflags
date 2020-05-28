package models

import (
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"set-flags/pkg/utils"
	"time"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	IdentityNumber string    `json:"identity_number"`
	MixinUserID    string    `json:"mixin_user_id"`
	FullName       string    `json:"full_name"`
	AvatarUrl      string    `json:"avatar_url"`
	AccessToken    string    `json:"access_token"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserSchema struct {
	ID             uuid.UUID `json:"id"`
	IdentityNumber int       `json:"identity_number"`
	FullName       string    `json:"full_name"`
	AvatarUrl      string    `json:"avatar_url"`
}

func FindUser(userId uuid.UUID) *User {
	var users []User
	db.Find(&users)
	for _, u := range users {
		if u.ID == userId {
			return &u
		}
	}
	return nil
}

func FindUserById(userId string) (user *UserSchema) {
	db.Where("id = ?", userId).First(&user)
	return
}

func CreateUser(userInfo *utils.UserInfo, accessToken string) bool {
	db.Create(&User{
		IdentityNumber: userInfo.IdentityNumber,
		FullName:       userInfo.Name,
		AvatarUrl:      "",
		AccessToken:    accessToken,
	})

	return true
}

func UserExist(userId string) bool {
	var count int

	db.Model(&User{}).Where("mixin_user_id = ?", userId).Count(&count)

	return count == 1
}

func UpdateUser(userInfo *utils.UserInfo, accessToken string) {
	db.Model(&User{}).
		Updates(map[string]interface{}{"full_name": userInfo.Name, "access_token": accessToken})
}

// BeforeCreate will set a UUID rather than numeric ID.
func (u *User) BeforeCreate(scope *gorm.Scope) error {
	uuid_, _ := uuid.NewV4()
	scope.SetColumn("ID", uuid_)
	scope.SetColumn("CreatedAt", time.Now())
	return nil
}

func (u *User) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}