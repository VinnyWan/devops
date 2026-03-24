package service

type UpdateUserRequest struct {
	ID           uint    `json:"id" binding:"required"`
	Username     *string `json:"username"`
	Name         *string `json:"name"`
	Email        *string `json:"email"`
	DepartmentID *uint   `json:"departmentId"`
	Status       *string `json:"status"`
}
