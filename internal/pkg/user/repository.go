package user

import "techpark_db/internal/pkg/models"

type Repository interface {
	Create(user models.User) error
	FindUser(user models.User) (*models.User, error)
	UpdateUser(user models.User) error
	FindUserByEmail(email string) (*models.User, error)
}
