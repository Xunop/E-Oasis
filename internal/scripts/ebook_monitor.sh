#!/bin/bash

# The directory to monitor. 
# [IMPORTANT] An absolute path is required for the systemd service to work correctly.
MONITOR_DIR="/storage/documents/ebooks"

# The API endpoint for uploading books.
API_URL="http://localhost:8080/api/v1/book"

# authentication cookie.
COOKIE="e-oasis-access-token=eyJhbGciOiJIUzI1NiIsImtpZCI6InYxIiwidHlwIjoiSldUIn0.eyJuYW1lIjoieHVuIiwiaXNzIjoiZS1vYXNpcyIsInN1YiI6IjEiLCJhdWQiOlsidXNlci5hYY2Nlc3MtdG9rZW4iXSwiZXhwIjo0OTExOTgxNjk4LCJpYXQiOjE3NTgzODE2OTh9.ZfO3OX_UYaYswX-z6STMGJUy8k2YAMfuWg1dYdHivUE"

# Check for 'jq' dependency.
if ! command -v jq &> /dev/null
then
    echo "Error: Dependency 'jq' is not installed. Please install it to proceed." >&2
    exit 1
fi

# Ensure the monitor directory exists.
if [ ! -d "$MONITOR_DIR" ]; then
  echo "Error: Directory to monitor '$MONITOR_DIR' does not exist." >&2
  exit 1
fi

echo "Service started: Now monitoring directory: $MONITOR_DIR"

# Use inotifywait to recursively monitor for 'close_write' and 'moved_to' events.
# The --format '%w%f' outputs the full path of the affected file.
inotifywait -m -r -e close_write -e moved_to --format '%w%f' "$MONITOR_DIR" | while read -r FULL_PATH
do
    echo "----------------------------------------"
    echo "Event detected for file: '$FULL_PATH'"

    # Validation Process

    # Check if the path is a regular file.
    if [ ! -f "$FULL_PATH" ]; then
        echo "Validation failed: Target is not a regular file. Skipping."
        continue
    fi

    # Get the file size.
    FILE_SIZE=$(stat -c %s "$FULL_PATH")

    # Check if the file size is greater than zero.
    if [ "$FILE_SIZE" -eq 0 ]; then
        echo "Validation failed: File size is 0 bytes. Skipping."
        continue
    fi
    
    # Check if the file size is stable (to handle large file copies).
    # Wait for 1 second and check the size again. This is crucial for network filesystems.
    sleep 1
    LATER_SIZE=$(stat -c %s "$FULL_PATH")
    
    if [ "$FILE_SIZE" -ne "$LATER_SIZE" ]; then
        echo "Validation failed: File size is still changing ($FILE_SIZE -> $LATER_SIZE). Skipping."
        continue
    fi

    echo "All validations passed (file exists, non-zero, stable size). Preparing to call API."
    
    # Core Actions

    # Upload the book file.
    echo "Uploading file '$FULL_PATH'..."
    UPLOAD_RESPONSE=$(curl -s --connect-timeout 5 --max-time 60 -X POST "$API_URL" \
          --header 'Content-Type: multipart/form-data' \
          --header 'Accept: application/json' \
          --header "Cookie: $COOKIE" \
          --form "file=@$FULL_PATH")

    if [ $? -ne 0 ]; then
        echo "API call failed (upload): Curl command execution error. Exit code: $?" >&2
        continue
    fi

    # Parse the Book ID from the JSON response.
    BOOK_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.id')

    if [ -z "$BOOK_ID" ] || [ "$BOOK_ID" == "null" ]; then
        echo "API call failed (upload): Could not parse a valid Book ID from the response." >&2
        echo "Server response: $UPLOAD_RESPONSE"
        continue
    fi
    echo "File upload complete. Got Book ID: $BOOK_ID"

    # Parse the path to get the tag.
    # Get the file's path relative to the monitored directory.
    RELATIVE_PATH=${FULL_PATH#$MONITOR_DIR/}
    # Get the directory part of the relative path.
    TAG=$(dirname "$RELATIVE_PATH")

    # If the tag is not '.' (meaning the file is in a subdirectory), add the tag.
    if [ "$TAG" != "." ]; then
        echo "Detected subdirectory. Adding '$TAG' as a tag..."
        
        TAG_API_URL="${API_URL}/${BOOK_ID}/tags"
        JSON_PAYLOAD="{\"tagName\": \"$TAG\"}"

        TAG_RESPONSE=$(curl -s --connect-timeout 5 --max-time 30 -X POST "$TAG_API_URL" \
            -H "Content-Type: application/json" \
            -H "Cookie: $COOKIE" \
            -d "$JSON_PAYLOAD")
        
        if [ $? -eq 0 ]; then
            echo "Add tag API call complete. Server response: $TAG_RESPONSE"
        else
            echo "API call failed (tag): Curl command execution error. Exit code: $?" >&2
        fi
    else
        echo "File is in the root directory. No tag to add."
    fi
done
