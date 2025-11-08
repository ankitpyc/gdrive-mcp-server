package driveapi

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/api/drive/v3"
)

// SuggestFolderForContent suggests a folder based on the content name.
// This is a simple heuristic-based suggestion.
func SuggestFolderForContent(ctx context.Context, srv *drive.Service, contentName string) (string, error) {
	contentNameLower := strings.ToLower(contentName)

	// Simple keyword-based suggestions
	if strings.Contains(contentNameLower, "report") {
		return findOrCreateFolder(ctx, srv, "Reports")
	}
	if strings.Contains(contentNameLower, "image") || strings.Contains(contentNameLower, "photo") {
		return findOrCreateFolder(ctx, srv, "Images")
	}
	if strings.Contains(contentNameLower, "document") || strings.Contains(contentNameLower, "doc") {
		return findOrCreateFolder(ctx, srv, "Documents")
	}
	if strings.Contains(contentNameLower, "code") || strings.Contains(contentNameLower, "src") {
		return findOrCreateFolder(ctx, srv, "Code")
	}

	// Default suggestion if no keywords match
	return findOrCreateFolder(ctx, srv, "Miscellaneous")
}

// findOrCreateFolder checks if a folder exists and returns its ID, otherwise creates it.
func findOrCreateFolder(ctx context.Context, srv *drive.Service, folderName string) (string, error) {
	folderID, err := FindFolderIDByName(ctx, srv, folderName, "") // Search in root
	if err == nil {
		return folderID, nil // Folder found
	}

	// Folder not found, create it
	folderMetadata := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{"root"},
	}
	folder, err := srv.Files.Create(folderMetadata).Fields("id").Do()
	if err != nil {
		log.Printf("Unable to create suggested folder '%s': %v", folderName, err)
		return "", fmt.Errorf("unable to create suggested folder '%s': %w", folderName, err)
	}
	return folder.Id, nil
}
