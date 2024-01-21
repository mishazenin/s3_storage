package upload

import "context"

type Uploader interface {
	UploadObject(ctx context.Context, obj []byte) (string, error)
}
