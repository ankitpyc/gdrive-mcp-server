package driveapi

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
)

// CreateFileInPath creates a file with the given content in the specified Google Drive path.
// The path should be a slash-separated string (e.g., "MyFolder/SubFolder/file.txt").
func CreateFileInPath(ctx context.Context, srv *drive.Service, filePath, content string) (*drive.File, error) {
	fileName := filepath.Base(filePath)
	folderPath := filepath.Dir(filePath)

	parentID, err := getOrCreateFolderPath(ctx, srv, folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create folder path: %w", err)
	}

	fileMetadata := &drive.File{
		Name:    fileName,
		Parents: []string{parentID},
	}
	res, err := srv.Files.Create(fileMetadata).SupportsAllDrives(true).Media(bytes.NewReader([]byte(content))).Do()
	if err != nil {
		log.Printf("Unable to create file '%s': %v", fileName, err)
		return nil, fmt.Errorf("unable to create file '%s': %w", fileName, err)
	}
	return res, nil
}

// getOrCreateFolderPath recursively finds or creates the folder path.
func getOrCreateFolderPath(ctx context.Context, srv *drive.Service, folderPath string) (string, error) {
	if folderPath == "." || folderPath == "/" {
		return "root", nil // Root folder
	}

	pathParts := strings.Split(folderPath, "/")
	currentParentID := "root"

	for _, part := range pathParts {
		if part == "" {
			continue
		}
		folderID, err := FindFolderIDByName(ctx, srv, part, currentParentID)
		if err != nil {
			// Folder not found, create it
			folderMetadata := &drive.File{
				Name:     part,
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{currentParentID},
			}
			folder, err := srv.Files.Create(folderMetadata).Fields("id").Do()
			if err != nil {
				return "", fmt.Errorf("unable to create folder '%s': %w", part, err)
			}
			currentParentID = folder.Id
		} else {
			currentParentID = folderID
		}
	}
	return currentParentID, nil
}

// ListFilesAndFoldersInFolder lists files and folders within a specific folder.
func ListFilesAndFoldersInFolder(ctx context.Context, srv *drive.Service, folderID string) ([]*drive.File, error) {
	if folderID == "" {
		folderID = "root"
	}
	q := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	r, err := srv.Files.List().Q(q).Fields("files(id, name, mimeType)").Do()
	if err != nil {
		log.Printf("Unable to retrieve files and folders from folder '%s': %v", folderID, err)
		return nil, fmt.Errorf("unable to retrieve files and folders from folder '%s': %w", folderID, err)
	}
	return r.Files, nil
}

// CreateDocxFileInPath creates a .docx file with the given content in the specified Google Drive path.
// The path should be a slash-separated string (e.g., "MyFolder/SubFolder/document.docx").
func CreateDocxFileInPath(ctx context.Context, srv *drive.Service, filePath, content string) (*drive.File, error) {
	fileName := filepath.Base(filePath)
	folderPath := filepath.Dir(filePath)

	if !strings.HasSuffix(strings.ToLower(fileName), ".docx") {
		return nil, fmt.Errorf("file name must have a .docx extension")
	}

	parentID, err := getOrCreateFolderPath(ctx, srv, folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create folder path: %w", err)
	}

	fileMetadata := &drive.File{
		Name:     fileName,
		Parents:  []string{parentID},
		MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	res, err := srv.Files.Create(fileMetadata).SupportsAllDrives(true).Media(bytes.NewReader([]byte(content))).Do()
	if err != nil {
		log.Printf("Unable to create file '%s': %v", fileName, err)
		return nil, fmt.Errorf("unable to create file '%s': %w", fileName, err)
	}
	return res, nil
}
