# Google Drive MCP Server ☁️

[![Go Reference](https://pkg.go.dev/badge/github.com/ankitpyc/gdrive-mcp-server.svg)](https://pkg.go.dev/github.com/ankitpyc/gdrive-mcp-server)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This project implements a Model Context Protocol (MCP) server that provides tools and resources for interacting with Google Drive. It allows for programmatic access to Google Drive functionalities, enabling other MCP-compatible agents or systems to manage files and folders.

## Current Features ✨

The server currently provides the following functionalities:

-   **File and Folder Listing** 📂: List files and folders within a specified Google Drive folder, including the root.
-   **File Creation** 📄: Create new files with specified content in a given Google Drive path.
-   **DOCX File Creation** 📝: Create new `.docx` files with specified content in a given Google Drive path.
-   **Folder Suggestion** 💡: Suggests a Google Drive folder based on the content name.

## Setup Instructions 🛠️

To set up and run the Google Drive MCP Server, follow these steps:

### Prerequisites ✅

-   Go (version 1.18 or higher)
-   Docker or Podman (for running the MCP server)
-   Google Cloud Project with Google Drive API enabled
-   Service Account Key (JSON file) for authentication with Google Drive API

### Configuration ⚙️

1.  **Enable Google Drive API**:
    -   Go to the [Google Cloud Console](https://console.cloud.google.com/).
    -   Create a new project or select an existing one.
    -   Navigate to "APIs & Services" > "Library".
    -   Search for "Google Drive API" and enable it.

2.  **Create a Service Account**:
    -   In the Google Cloud Console, go to "APIs & Services" > "Credentials".
    -   Click "Create Credentials" > "Service Account".
    -   Follow the steps to create a new service account.
    -   Grant the service account appropriate roles (e.g., "Drive File Organizer" or "Drive Editor") to access Google Drive.

3.  **Generate a Service Account Key**:
    -   After creating the service account, click on its email address.
    -   Go to the "Keys" tab and click "Add Key" > "Create new key".
    -   Select "JSON" as the key type and click "Create".
    -   A JSON file will be downloaded. Rename it to `credentials.json` and place it in the `configs/` directory of this project.

4.  **Share Google Drive Folders/Files with Service Account**:
    -   The service account needs explicit access to the Google Drive folders/files it will interact with. Share the relevant folders/files with the service account's email address (found in the `credentials.json` file).

### Running the Server 🚀

1.  **Build the Docker Image**:
    ```bash
    docker build -t gdrive-mcp-server .
    ```
    or
    ```bash
    podman build -t gdrive-mcp-server .
    ```

2.  **Run the MCP Server**:
    ```bash
    docker run --rm -i -e GOOGLE_APPLICATION_CREDENTIALS=/app/configs/credentials.json -v $(pwd)/configs:/app/configs gdrive-mcp-server
    ```
    or
    ```bash
    podman run --rm -i -e GOOGLE_APPLICATION_CREDENTIALS=/app/configs/credentials.json -v $(pwd)/configs:/app/configs gdrive-mcp-server
    ```
    Ensure that `$(pwd)/configs` correctly points to the directory containing your `credentials.json` file.

## Planned Features 💡

-   **File Update**: Update existing files in Google Drive.
-   **File Deletion**: Delete files from Google Drive.
-   **Folder Creation**: Create new folders in Google Drive.
-   **File Search**: Advanced search capabilities for files based on various criteria (name, type, content).
-   **Permissions Management**: Manage file and folder permissions.
-   **Webhooks/Notifications**: Integrate with Google Drive change notifications.
