package dashboard

func (a *App) createRoutes() {
	a.router.GET("/", a.main)
	a.router.GET("/auth/callback", a.auth)
	a.router.GET("/admin/logout", a.logout)
}
