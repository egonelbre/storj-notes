package notes

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"storj.io/uplink"
)

// Service implements accessing notes.
type Service struct {
	project *uplink.Project
	bucket  string
}

// Open opens a new service.
func Open(ctx context.Context, access *uplink.Access, bucket string) (*Service, error) {
	// Open the project.
	project, err := uplink.OpenProject(ctx, access)
	if err != nil {
		return nil, fmt.Errorf("failed to open project: %w", err)
	}

	// Ensure that our notes bucket exists.
	_, err = project.EnsureBucket(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure bucket %q: %w", bucket, err)
	}

	return &Service{
		project: project,
		bucket:  bucket,
	}, nil
}

// Get loads the specific note from Storj network.
func (service *Service) Get(ctx context.Context, identifier string) (Note, error) {
	// Start a new download from the network.
	download, err := service.project.DownloadObject(ctx, service.bucket, identifier, nil)
	if err != nil {
		return Note{}, fmt.Errorf("failed to start download %q: %w", identifier, err)
	}
	defer download.Close()

	// Currently we are downloading only small notes, hence we can read it into memory.
	// When working with larger files (1GB) we should stream them.
	data, err := ioutil.ReadAll(download)
	if err != nil {
		return Note{}, fmt.Errorf("failed to download %q: %w", identifier, err)
	}

	note := ParseNote(download.Info(), data)
	return note, nil
}

// Set sets the specific note.
func (service *Service) Set(ctx context.Context, identifier, value string) error {
	upload, err := service.project.UploadObject(ctx, service.bucket, identifier, nil)
	if err != nil {
		return fmt.Errorf("failed to start upload %q: %w", identifier, err)
	}

	// Currently we are uploading only small notes, hence we can keep them in memory.
	// When working with larger files (1GB) we should stream them.
	_, err = upload.Write([]byte(value))
	if err != nil {
		// Ensure that the satellite knows we failed to upload.
		aborterr := upload.Abort()
		return fmt.Errorf("failed to upload %q: %w, %v", identifier, err, aborterr)
	}

	// Set some additional information about the note.
	err = upload.SetCustomMetadata(ctx, uplink.CustomMetadata{
		uploadTime: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		// Ensure that the satellite knows we failed to upload.
		aborterr := upload.Abort()
		return fmt.Errorf("failed to set metadata %q: %w, %v", identifier, err, aborterr)
	}

	// Commit the data to the database.
	err = upload.Commit()
	if err != nil {
		return fmt.Errorf("failed commit %q: %w", identifier, err)
	}

	return nil
}

// Delete deletes the specified note.
func (service *Service) Delete(ctx context.Context, identifier string) error {
	_, err := service.project.DeleteObject(ctx, service.bucket, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete %q: %w", identifier, err)
	}

	return nil
}

// List lists all the items.
func (service *Service) List(ctx context.Context, prefix string) ([]NoteMeta, error) {
	it := service.project.ListObjects(ctx, service.bucket, &uplink.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
		Custom:    true,
	})
	var notes []NoteMeta
	for it.Next() {
		info := it.Item()
		meta := ParseNoteMeta(info)
		notes = append(notes, meta)
	}
	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed (prefix=%q): %w", prefix, err)
	}
	return notes, nil
}

// Close closes the service and resources.
func (service *Service) Close() error {
	return service.project.Close()
}
