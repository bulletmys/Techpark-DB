package usecase

import (
	"fmt"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/user"
)

type UserUC struct {
	UserRepo user.Repository
}

//Возможно стоит сделать проверку на существование и создание юзера в рамках одной транзакции
func (uc UserUC) Create(user models.User) ([]models.User, error) {
	dbUser, err := uc.UserRepo.FindUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to check users existing: %v", err)
	}
	if dbUser != nil {
		return dbUser, nil
	}
	return nil, uc.UserRepo.Create(user)
}

//Можно в качестве аргумента принимать структуру и в нее же записывать результат
func (uc UserUC) Find(user models.User) (*models.User, error) {
	dbUser, err := uc.UserRepo.FindUserByNickname(user.Nickname)
	if err != nil {
		return nil, fmt.Errorf("failed to find user in db: %v", err)
	}
	if dbUser == nil {
		return nil, models.UserNotFound
	}
	return dbUser, nil
}

//Возможно стоит сделать в рамках одной транзакции
func (uc UserUC) Update(user models.User) (*models.User, error) {
	if user.Email != "" {
		if userByEmail, _ := uc.UserRepo.FindUserByEmail(user.Email); userByEmail != nil {
			return nil, models.SameUserExists
		}
	}

	dbUser, err := uc.UserRepo.FindUserByNickname(user.Nickname)
	if err != nil {
		return nil, err
	}
	if dbUser == nil {
		return nil, models.UserNotFound
	}

	MergeModels(dbUser, user)

	if err = uc.UserRepo.UpdateUser(*dbUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}
	return dbUser, nil
}


func MergeModels(dbUser *models.User, updatedUser models.User) {
	if updatedUser.FullName != "" {
		dbUser.FullName = updatedUser.FullName
	}
	if updatedUser.Email != "" {
		dbUser.Email = updatedUser.Email
	}
	if updatedUser.About != "" {
		dbUser.About = updatedUser.About
	}
}
