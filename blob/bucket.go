package blob

import "encore.dev/storage/objects"

var UploadsBucket = objects.NewBucket("uploads", objects.BucketConfig{
	Versioned: false,
})
