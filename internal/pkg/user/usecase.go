package user

import "techpark_db/internal/pkg/models"

type UseCase interface {
	Create(user models.User) (*models.User, error)
	Find(user models.User) (*models.User, error)
	Update(user models.User) (*models.User, error)
}
