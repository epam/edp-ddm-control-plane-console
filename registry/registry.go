package registry

type registry struct {
	Name                string `form:"name" binding:"required,min=3,max=12,registry-name"`
	Description         string `form:"description" valid:"max=250"`
	Admins              string `form:"admins" binding:"registry-admins"`
	SignKeyIssuer       string `form:"sign-key-issuer"`
	SignKeyPwd          string `form:"sign-key-pwd"`
	RegistryGitTemplate string `form:"registry-git-template" binding:"required"`
	RegistryGitBranch   string `form:"registry-git-branch" binding:"required"`
}
