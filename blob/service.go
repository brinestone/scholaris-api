// CRUD endpoints for uploading and serving static files
package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"encore.dev"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/helpers"
	"github.com/brinestone/scholaris/util"
)

const (
	FTUser   = "u"
	FTShared = "sh"
)

// Serve File
//
//encore:api raw public method=GET path=/blob/:key
func ServeFile(w http.ResponseWriter, req *http.Request) {
	var key = encore.CurrentRequest().PathParams.Get("key")
	uid, authed := auth.UserID()
	var userId string
	var userIdParsed *uint64
	if !authed {
		userId = "*"
	} else {
		userId = string(uid)
		tmp, err := strconv.ParseUint(userId, 10, 64)
		if err != nil {
			rlog.Error("user id parse error", "err", err)
			errs.HTTPError(w, &util.ErrUnknown)
			return
		}
		userIdParsed = &tmp
	}
	var fileType dto.PermissionType
	if strings.HasPrefix(key, FTShared) {
		fileType = dto.PTSharedFile
	} else {
		fileType = dto.PTUserFile
	}

	p, err := permissions.CheckPermissionInternal(req.Context(), dto.InternalRelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, userId),
		Relation: dto.PermCanView,
		Target:   dto.IdentifierString(fileType, key),
	})
	if err != nil {
		rlog.Error(util.MsgCallError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}
	if !p.Allowed {
		errs.HTTPError(w, &util.ErrNotFound)
		return
	}

	reader := UploadsBucket.Download(req.Context(), key)
	if errors.Is(reader.Err(), objects.ErrObjectNotFound) {
		errs.HTTPError(w, &util.ErrNotFound)
		return
	} else if reader.Err() != nil {
		rlog.Error("bucket error", "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}
	defer reader.Close()

	mime, err := findUploadMimeType(req.Context(), key)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	tx, err := db.Begin(req.Context())
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}
	defer func() {
		if err := registerDownload(req.Context(), tx, key, userIdParsed); err != nil {
			rlog.Error("error while registering download", "err", err)
		}
	}()

	w.Header().Set("Content-Type", mime)
	io.Copy(w, reader)
}

// Uploads a file
//
//encore:api raw auth method=POST path=/blob/upload
func UploadFile(w http.ResponseWriter, req *http.Request) {

	requestData, err := dto.ParseUploadRequest(req)
	if err != nil {
		rlog.Error("upload error", "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	} else if err = requestData.Validate(); err != nil {
		rlog.Warn("validation error", "err", err)
		errs.HTTPError(w, errs.WrapCode(err, errs.InvalidArgument, "Bad request"))
		return
	}

	uid, _ := auth.UserID()
	userId, _ := strconv.ParseUint(string(uid), 10, 64)

	var owner uint64
	var ownerType, fileType dto.PermissionType
	var key string = helpers.NewUlid()

	if requestData.OwnerInfoSet() {
		owner = requestData.Owner
		ownerType = dto.PermissionType(requestData.OwnerType)
		fileType = dto.PTSharedFile
		key = fmt.Sprintf("%s_%s", FTShared, key)
	} else {
		owner = userId
		ownerType = dto.PTUser
		fileType = dto.PTUserFile
		key = fmt.Sprintf("%s_%s", FTUser, key)
	}

	if requestData.OwnerInfoSet() {
		p, err := permissions.CheckPermissionInternal(req.Context(), dto.InternalRelationCheckRequest{
			Actor:    dto.IdentifierString(dto.PTUser, uid),
			Relation: dto.PermCanUploadFile,
			Target:   dto.IdentifierString(ownerType, owner),
		})
		if err != nil {
			rlog.Error(util.MsgCallError, "err", err)
			errs.HTTPError(w, &util.ErrUnknown)
			return
		}

		if !p.Allowed {
			rlog.Warn("invalid permissions for upload", "owner", owner, "ownerType", ownerType, "user", userId)
			errs.HTTPError(w, &util.ErrForbidden)
			return
		}
	}

	writer := UploadsBucket.Upload(req.Context(), key)
	if err := req.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		writer.Abort(err)
		rlog.Error("upload error", "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	file, fileInfo, err := req.FormFile("file")
	if err != nil {
		writer.Abort(err)
		rlog.Error("upload error", "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		writer.Abort(err)
		rlog.Error("upload error", "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	tx, err := db.Begin(req.Context())
	if err != nil {
		writer.Abort(err)
		rlog.Error(util.MsgDbAccessError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	err = registerUpload(req.Context(), tx, userId, owner, uint64(fileInfo.Size), string(ownerType), fileInfo.Filename, fileInfo.Header.Get("Content-Type"), key)
	if err != nil {
		tx.Rollback()
		writer.Abort(err)
		rlog.Error(util.MsgDbAccessError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	err = permissions.SetPermissions(req.Context(), dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Actor:    dto.IdentifierString(ownerType, owner),
				Relation: dto.PermOwner,
				Target:   dto.IdentifierString(fileType, key),
			},
		},
	})

	if err != nil {
		tx.Rollback()
		writer.Abort(err)
		rlog.Error(util.MsgCallError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}
}

func registerUpload(ctx context.Context, tx *sqldb.Tx, user, owner, size uint64, ownerType, fileName, mimeType, key string) (err error) {
	query := `
		INSERT INTO
			uploads(name,mime_type,size,uploaded_by,owner,owner_type,key)
		VALUES
			($1,$2,$3,$4,$5,$6,$7);
	`

	_, err = tx.Exec(ctx, query, fileName, mimeType, size, user, owner, ownerType, key)
	return
}

func findUploadMimeType(ctx context.Context, key string) (ans string, err error) {
	query := `
		SELECT mime_type FROM uploads WHERE key=$1;
	`
	err = db.QueryRow(ctx, query, key).Scan(&ans)
	return
}

func registerDownload(ctx context.Context, tx *sqldb.Tx, key string, user *uint64) (err error) {
	query := `
		INSERT INTO downloads (downloaded_by, key) VALUES ($1, $2);
	`
	_, err = tx.Exec(ctx, query, user, key)
	return
}
