package driveapi

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/drive/v3"
)

// ListRootFolders fetches the list of root-level folders in Google Drive.
func ListRootFolders(ctx context.Context, srv *drive.Service) ([]*drive.File, error) {
	// Query for folders that are either in the root or shared with the service account
	q := "('root' in parents or sharedWithMe = true) and mimeType = 'application/vnd.google-apps.folder' and trashed = false"
	r, err := srv.Files.List().Q(q).Fields("files(id, name)").Do()
	if err != nil {
		log.Printf("Unable to retrieve root folders: %v", err)
		return nil, fmt.Errorf("unable to retrieve root folders: %w", err)
	}
	return r.Files, nil
}

// FindFolderIDByName finds a folder by its name within a given parent.
// If parentID is empty, it searches in the root.
func FindFolderIDByName(ctx context.Context, srv *drive.Service, folderName, parentID string) (string, error) {
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", folderName)
	if parentID != "" {
		q = fmt.Sprintf("'%s' in parents and %s", parentID, q)
	} else {
		q = fmt.Sprintf("'root' in parents and %s", q)
	}

	r, err := srv.Files.List().Q(q).Fields("files(id)").Do()
	if err != nil {
		log.Printf("Unable to find folder '%s': %v", folderName, err)
		return "", fmt.Errorf("unable to find folder '%s': %w", folderName, err)
	}

	if len(r.Files) == 0 {
		return "", fmt.Errorf("folder '%s' not found", folderName)
	}
	return r.Files[0].Id, nil
}
