package main

import (
	"context"
	"encoding/json"
	"log"

	"google-drive-mcp-server/pkg/driveapi"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	ctx := context.Background()

	// Initialize Google Drive Service
	srv, err := driveapi.GetDriveService(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Google Drive service: %v", err)
	}

	about, err := srv.About.Get().Fields("user(emailAddress)").Do()
	if err != nil {
		log.Printf("About error: %v", err)
	} else {
		log.Printf("Drive user email: %s", about.User.EmailAddress)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"Google Drive MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true), // Enable tool capabilities
	)

	// Register "fetch the list of root level folders" tool
	listRootFoldersTool := mcp.NewTool("list_root_folders",
		mcp.WithDescription("Fetches the list of root level folders in Google Drive."),
	)
	s.AddTool(listRootFoldersTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		folders, err := driveapi.ListRootFolders(ctx, srv)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result := make([]map[string]string, len(folders))
		for i, folder := range folders {
			result[i] = map[string]string{"id": folder.Id, "name": folder.Name}
		}
		jsonResult, err := json.Marshal(map[string]interface{}{"folders": result})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Register "create a file in the path" tool
	createFileTool := mcp.NewTool("create_file_in_path",
		mcp.WithDescription("Creates a file with the given content in the specified Google Drive path."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("The full path including filename (e.g., 'MyFolder/file.txt')"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The content of the file"),
		),
	)
	s.AddTool(createFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		file, err := driveapi.CreateFileInPath(ctx, srv, filePath, content)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		jsonResult, err := json.Marshal(map[string]interface{}{"file_id": file.Id, "file_name": file.Name})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Register "create a docx file in the path" tool
	createDocxFileTool := mcp.NewTool("create_docx_file_in_path",
		mcp.WithDescription("Creates a .docx file with the given content in the specified Google Drive path."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("The full path including filename (e.g., 'MyFolder/document.docx')"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The content of the file"),
		),
	)
	s.AddTool(createDocxFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		file, err := driveapi.CreateDocxFileInPath(ctx, srv, filePath, content)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		jsonResult, err := json.Marshal(map[string]interface{}{"file_id": file.Id, "file_name": file.Name})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Register "suggest based on the name of the folder which folder this kind of content goes" tool
	suggestFolderTool := mcp.NewTool("suggest_folder_for_content",
		mcp.WithDescription("Suggests a folder based on the content name."),
		mcp.WithString("content_name",
			mcp.Required(),
			mcp.Description("The name of the content to suggest a folder for"),
		),
	)
	s.AddTool(suggestFolderTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentName, err := request.RequireString("content_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		suggestedFolderID, err := driveapi.SuggestFolderForContent(ctx, srv, contentName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		jsonResult, err := json.Marshal(map[string]interface{}{"suggested_folder_id": suggestedFolderID})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Register "list files and folders in a folder" tool
	listFilesAndFoldersTool := mcp.NewTool("list_files_and_folders",
		mcp.WithDescription("Lists files and folders within a specific folder."),
		mcp.WithString("folder_id",
			mcp.Description("The ID of the folder to list files and folders from. Defaults to root."),
		),
	)
	s.AddTool(listFilesAndFoldersTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		folderID := request.GetString("folder_id", "")
		files, err := driveapi.ListFilesAndFoldersInFolder(ctx, srv, folderID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result := make([]map[string]string, len(files))
		for i, file := range files {
			result[i] = map[string]string{"id": file.Id, "name": file.Name, "mime_type": file.MimeType}
		}
		jsonResult, err := json.Marshal(map[string]interface{}{"files": result})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Explicitly add mcp/list_tools for testing
	listToolsMCPTool := mcp.NewTool("mcp/list_tools",
		mcp.WithDescription("Lists all available tools on the MCP server."),
	)
	s.AddTool(listToolsMCPTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tools := s.ListTools()
		toolNames := make([]string, 0, len(tools))
		for name := range tools {
			toolNames = append(toolNames, name)
		}
		jsonResult, err := json.Marshal(map[string]interface{}{"tools": toolNames})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Create a new SSE server
	sseServer := server.NewSSEServer(s,
		server.WithBaseURL("http://localhost:8080"), // Adjust if running on a different host/port
	)

	// Start the SSE server
	log.Printf("Starting MCP SSE server on :8080...")
	if err := sseServer.Start(":8080"); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
