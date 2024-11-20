package users

import "encore.dev/storage/objects"

var profilePhotos = objects.NewBucket("profile-photos", objects.BucketConfig{})

type ProfilePhotoUploader interface {
	objects.Uploader
}

var profilePhotoUploads = objects.BucketRef[ProfilePhotoUploader](profilePhotos)
