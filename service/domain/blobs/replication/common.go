package replication

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

var ErrBlobNotFound = errors.New("blob not found")

type BlobSizeRepository interface {
	// Size returns the size of the blob. If the blob is not found it returns
	// ErrBlobNotFound.
	Size(id refs.Blob) (blobs.Size, error)
}

type WantedBlobsProvider interface {
	GetWantedBlobs() ([]refs.Blob, error)
}

type BlobsThatShouldBePushedProvider interface {
	GetBlobsThatShouldBePushed() ([]refs.Blob, error)
}

type HasBlobHandler interface {
	OnHasReceived(ctx context.Context, peer transport.Peer, blob refs.Blob, size blobs.Size)
}
