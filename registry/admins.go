package registry

import (
	"context"
	"fmt"
	"github.com/gosimple/slug"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ddm-admin-console/service/keycloak"
)

type admin struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	TmpPassword string `json:"tmpPassword"`
}


func (a *App) getAdmins(ctx context.Context, registryName string) ([]admin, error) {
	usrs, err := a.keycloakService.GetUsersByRealm(ctx, a.usersRealm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get users by realm")
	}

	var admins []admin
	for _, u := range usrs {
		for _, g := range u.Spec.Groups {
			if g == groupRoleName(registryName) {
				admins = append(admins, admin{
					Username:  u.Spec.Username,
					Email:     u.Spec.Email,
					FirstName: u.Spec.FirstName,
					LastName:  u.Spec.LastName,
				})
			}
		}
	}

	return admins, nil
}

func (a *App) syncAdmins(ctx context.Context, registryName string, admins []admin) error {
	usrs, err := a.keycloakService.GetUsersByRealm(ctx, a.usersRealm)
	if err != nil {
		return errors.Wrap(err, "unable to get users by realm")
	}

	adminsDict := make(map[string]admin)
	for i, v := range admins {
		adminsDict[v.Email] = admins[i]
	}

	for i := range usrs {
		if _, ok := adminsDict[usrs[i].Spec.Email]; ok {
			// add to groups and roles
			// remove from dict
			if err := a.adminAddToGroupsAndRoles(ctx, registryName, &usrs[i]); err != nil {
				return errors.Wrap(err, "unable to add registry to realm user groups and roles")
			}

			delete(adminsDict, usrs[i].Spec.Email)
		} else {
			// remove from groups and roles
			if err := a.adminRemoveFromGroupsAndRoles(ctx, registryName, &usrs[i]); err != nil {
				return errors.Wrap(err, "unable to remote registry from realm user groups and roles")
			}
		}
	}

	for _, adm := range adminsDict {
		if err := a.adminCreate(ctx, registryName, &adm); err != nil {
			return errors.Wrap(err, "unable to create admin")
		}
	}

	return nil
}

func (a *App) adminCreate(ctx context.Context, registryName string, adm *admin) error {
	if err := a.keycloakService.CreateUser(ctx, &keycloak.KeycloakRealmUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", registryName,
				strings.Replace(slug.Make(adm.Username), "_", "-", -1)),
			Namespace: a.usersNamespace,
		},
		Spec: keycloak.KeycloakRealmUserSpec{
			FirstName:           adm.FirstName,
			LastName:            adm.LastName,
			Username:            adm.Username,
			Roles:               []string{groupRoleName(registryName)},
			Groups:              []string{groupRoleName(registryName)},
			Email:               adm.Email,
			Realm:               a.usersRealm,
			Password:            adm.TmpPassword,
			KeepResource:        true,
			Enabled:             true,
			EmailVerified:       true,
			RequiredUserActions: []string{"UPDATE_PASSWORD"},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to create realm user from admin")
	}

	return nil
}

func groupRoleName(registryName string) string {
	return fmt.Sprintf("cp-registry-admin-%s", registryName)
}

func (a *App) adminRemoveFromGroupsAndRoles(ctx context.Context, registryName string, u *keycloak.KeycloakRealmUser) error {
	groupRemoved, roleRemoved := false, false
	var newGroups, newRoles []string

	for _, g := range u.Spec.Groups {
		if g == groupRoleName(registryName) {
			groupRemoved = true
			continue
		}

		newGroups = append(newGroups, g)
	}

	if groupRemoved {
		u.Spec.Groups = newGroups
	}

	for _, r := range u.Spec.Roles {
		if r == groupRoleName(registryName) {
			roleRemoved = true
			continue
		}

		newRoles = append(newRoles, r)
	}

	if roleRemoved {
		u.Spec.Roles = newRoles
	}

	if groupRemoved || roleRemoved {
		if err := a.keycloakService.UpdateUser(ctx, u); err != nil {
			return errors.Wrap(err, "unable to update user")
		}
	}

	return nil
}

func (a *App) adminAddToGroupsAndRoles(ctx context.Context, registryName string, u *keycloak.KeycloakRealmUser) error {
	addToGroups, addToRoles := true, true
	for _, g := range u.Spec.Groups {
		if g == groupRoleName(registryName) {
			addToGroups = false
			break
		}
	}

	if addToGroups {
		u.Spec.Groups = append(u.Spec.Groups, groupRoleName(registryName))
	}

	for _, r := range u.Spec.Roles {
		if r == groupRoleName(registryName) {
			addToRoles = false
			break
		}
	}

	if addToRoles {
		u.Spec.Roles = append(u.Spec.Roles, groupRoleName(registryName))
	}

	if addToRoles || addToGroups {
		if err := a.keycloakService.UpdateUser(ctx, u); err != nil {
			return errors.Wrap(err, "unable to update user")
		}
	}

	return nil
}
