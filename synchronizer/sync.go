package synchronizer

import (
	"context"
	"fmt"
	"log"

	"github.com/aryadiwwt/synctodb-anggarandetail/fetcher"
	"github.com/aryadiwwt/synctodb-anggarandetail/storer"
)

// PostSynchronizer mengorkestrasi proses sinkronisasi data post.
type AnggaranDetailSynchronizer struct {
	fetcher fetcher.Fetcher
	storer  storer.Storer
	log     *log.Logger
}

func NewAnggaranDetailSynchronizer(f fetcher.Fetcher, s storer.Storer, l *log.Logger) *AnggaranDetailSynchronizer {
	return &AnggaranDetailSynchronizer{
		fetcher: f,
		storer:  s,
		log:     l,
	}
}

// SynchronizePosts menjalankan seluruh alur kerja sinkronisasi.
func (ps *AnggaranDetailSynchronizer) Synchronize(ctx context.Context) error {
	ps.log.Println("Starting post synchronization...")

	details, err := ps.fetcher.FetchAnggaranDetails(ctx)
	if err != nil {
		return fmt.Errorf("synchronization failed during fetch phase: %w", err)
	}
	ps.log.Printf("Successfully fetched %d details.", len(details))

	if len(details) == 0 {
		ps.log.Println("No new details to synchronize.")
		return nil
	}

	if err := ps.storer.StoreAnggaranDetails(ctx, details); err != nil {
		return fmt.Errorf("synchronization failed during store phase: %w", err)
	}
	ps.log.Println("Successfully stored details to the database.")

	ps.log.Println("detail synchronization finished successfully.")
	return nil
}
