#!/bin/bash

# Get the directory where the script is located
BASE_DIR="$(dirname "$(realpath "$0")")"

# Define the folder paths relative to the base directory
FOLDER_PATHS=(
    "$BASE_DIR/analyzer/"
    "$BASE_DIR/toolchain"
)

GO_RUNTIME_PATH="$BASE_DIR/go-patch/src"

# Loop through the folder paths
for FOLDER_PATH in "${FOLDER_PATHS[@]}"
do
    echo "Building in: $FOLDER_PATH"

    # Navigate to the folder
    cd "$FOLDER_PATH" || { echo "Folder not found! Skipping..."; continue; }

    # Run go build
    go build

    # Check if go build was successful
    if [ $? -eq 0 ]; then
        echo "Build successful in $FOLDER_PATH"
    else
        echo "Build failed in $FOLDER_PATH"
    fi

    echo # Print a blank line for readability
done

echo "Building go runtime from $GO_RUNTIME_PATH"

cd "$GO_RUNTIME_PATH" || { echo "Folder not found! Skipping..."; return; }

./make.bash

# Check if go build was successful
echo
if [ $? -eq 0 ]; then
    echo "Building runtime successful in $GO_RUNTIME_PATH"
else
    echo "Building runtime failed in $GO_RUNTIME_PATH"
fi

echo # Print a blank line for readability