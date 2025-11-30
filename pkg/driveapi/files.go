package driveapi

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
)

// SearchDriveItems searches for files and folders based on a query string.
// The query string should follow the Google Drive API search syntax (e.g., "name contains 'Projects'").
func SearchDriveItems(ctx context.Context, srv *drive.Service, query string) ([]*drive.File, error) {
	var allFiles []*drive.File
	pageToken := ""

	for {
		req := srv.Files.List().Q(query).Fields("nextPageToken, files(id, name, mimeType)")
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			log.Printf("Unable to search drive items with query '%s': %v", query, err)
			return nil, fmt.Errorf("unable to search drive items: %w", err)
		}
		allFiles = append(allFiles, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return allFiles, nil
}

// ReadFileContent reads the content of a file, handling different MIME types.
// For .docx and Google Docs files, it attempts to export them as plain text.
func ReadFileContent(ctx context.Context, srv *drive.Service, fileID string, mimeType string) (string, error) {
	var resp *http.Response
	var err error

	switch mimeType {
	// CASE A: Google Native Docs (Must use Export)
	case "application/vnd.google-apps.document":
		// FIX 1: Use .Download() instead of .Do() to get the response body
		resp, err = srv.Files.Export(fileID, "text/plain").Context(ctx).Download()
		if err != nil {
			return "", fmt.Errorf("unable to export google doc '%s': %w", fileID, err)
		}

	// CASE B: Binary Files (.docx, .pdf) (Must use Get -> Download)
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/pdf":
		// FIX 2: .docx files cannot be 'Exported'. They are binary blobs, so we use Get().Download()
		// WARNING: This returns binary data (ZIP for docx, PDF bytes), not plain text.
		// You will need a parser library to convert this string into readable text.
		resp, err = srv.Files.Get(fileID).Context(ctx).Download()
		if err != nil {
			return "", fmt.Errorf("unable to download binary file '%s': %w", fileID, err)
		}

	// CASE C: Plain Text
	default:
		if strings.HasPrefix(mimeType, "text/") {
			resp, err = srv.Files.Get(fileID).Context(ctx).Download()
			if err != nil {
				return "", fmt.Errorf("unable to download text file '%s': %w", fileID, err)
			}
		} else {
			return "", fmt.Errorf("unsupported mime type for reading: %s", mimeType)
		}
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("unable to read file content: %w", err)
	}
	return buf.String(), nil
}

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

// FindFileIDByName finds a file by its name within a specific parent folder.
func FindFileIDByName(ctx context.Context, srv *drive.Service, fileName, parentID string) (string, error) {
	q := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false and mimeType != 'application/vnd.google-apps.folder'", fileName, parentID)
	r, err := srv.Files.List().Q(q).Fields("files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve files: %w", err)
	}
	if len(r.Files) == 0 {
		return "", fmt.Errorf("file '%s' not found in parent '%s'", fileName, parentID)
	}
	return r.Files[0].Id, nil
}

// UpdateDocxFileContent updates the content of an existing .docx file.
// The path should be a slash-separated string (e.g., "MyFolder/SubFolder/document.docx").
func UpdateDocxFileContent(ctx context.Context, srv *drive.Service, filePath, content string) (*drive.File, error) {
	fileName := filepath.Base(filePath)
	folderPath := filepath.Dir(filePath)

	if !strings.HasSuffix(strings.ToLower(fileName), ".docx") {
		return nil, fmt.Errorf("file name must have a .docx extension")
	}

	parentID, err := getOrCreateFolderPath(ctx, srv, folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create folder path: %w", err)
	}

	fileID, err := FindFileIDByName(ctx, srv, fileName, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find docx file '%s': %w", fileName, err)
	}

	fileMetadata := &drive.File{
		Name: fileName,
		// MimeType is set to application/vnd.openxmlformats-officedocument.wordprocessingml.document
		// explicitly during update to ensure it's treated as a DOCX.
		// If not set, it might default to plain text or other mime type on update.
		MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	res, err := srv.Files.Update(fileID, fileMetadata).SupportsAllDrives(true).Media(bytes.NewReader([]byte(content))).Do()
	if err != nil {
		log.Printf("Unable to update file '%s': %v", fileName, err)
		return nil, fmt.Errorf("unable to update file '%s': %w", fileName, err)
	}
	return res, nil
}
