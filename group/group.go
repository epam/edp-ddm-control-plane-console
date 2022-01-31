package group

type group struct {
	Name        string `form:"name" binding:"required,min=3,max=12,registry-name"`
	Description string `form:"description" valid:"max=250"`
}
