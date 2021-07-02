package registry

//const viewTimeFormat = "02.01.2006 15:04"

type registry struct {
	Name          string `form:"name" binding:"required,min=3,max=12,registry-name"`
	Description   string `form:"description" valid:"max=250"`
	Admins        string `form:"admins" binding:"registry-admins"`
	SignKeyIssuer string `form:"sign-key-issuer"`
	SignKeyPwd    string `form:"sign-key-pwd"`
	//CreatedAt     time.Time
	//UpdatedAt     time.Time
}

//func (r registry) FormattedCreatedAt() string {
//	return r.CreatedAt.Format(viewTimeFormat)
//}
