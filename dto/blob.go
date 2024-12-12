package dto

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type UploadRequest struct {
	// Permissions []string
	Owner     uint64
	OwnerType string
}

func (u UploadRequest) OwnerInfoSet() bool {
	if u.Owner != 0 && len(u.OwnerType) > 0 {
		pt, valid := ParsePermissionType(u.OwnerType)
		if !valid || (pt != PTUser && pt != PTInstitution && pt != PTTenant) {
			return false
		}
		return true
	}
	return false
}

func ParseUploadRequest(r *http.Request) (u UploadRequest, err error) {
	if r.Form.Has("owner") {
		u.Owner, err = strconv.ParseUint(r.FormValue("owner"), 10, 64)
		if err != nil {
			return
		}
	}

	if r.Form.Has("ownerType") {
		u.OwnerType = r.FormValue("ownerType")
	}
	return
}

func (u UploadRequest) Validate() (err error) {
	msgs := make([]string, 0)
	// if len(u.Permissions) > 0 {
	// 	permissionsValid := helpers.Every(u.Permissions, func(p string) (ans bool) {
	// 		split := strings.Split(p, ":")
	// 		pt, ans := ParsePermissionType(split[0])
	// 		ans = ans && (pt == PTInstitution || pt == PTUser || pt == PTTenant)
	// 		return
	// 	})
	// 	if !permissionsValid {
	// 		msgs = append(msgs, "Invalid permissions")
	// 	}
	// }

	if u.Owner != 0 && len(u.OwnerType) > 0 {
		pt, ans := ParsePermissionType(u.OwnerType)
		if !ans || (pt != PTUser && pt != PTInstitution && pt != PTTenant) {
			msgs = append(msgs, "Invalid value for x-owner-type header")
		}
	}

	if len(msgs) > 0 {
		err = errors.New(strings.Join(msgs, "\n"))
	}
	return
}
