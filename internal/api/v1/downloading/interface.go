package downloading

import (
	"context"
	"io"
)

type Downlaoder interface {
	DownloadObject(ctx context.Context, objName string) (io.Reader, error)
}
