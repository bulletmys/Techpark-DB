package user

import "techpark_db/internal/pkg/models"

type Repository interface {
	Create(user models.User) error
	FindUser(user models.User) ([]models.User, error)
	UpdateUser(user models.User) error
	FindUserByEmail(email string) (*models.User, error)
	FindUserByNickname(nick string) (*models.User, error)
	GetForumUsers(slug, since string, limit int, desc bool) ([]models.User, error)
}
