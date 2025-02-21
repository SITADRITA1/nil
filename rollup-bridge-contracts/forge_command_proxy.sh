#!/bin/bash

# Set the root directory of the workspace
WORKSPACE_ROOT=$(cd "$(dirname "$0")/../../.." && pwd)

# Print the WORKSPACE_ROOT for debugging
echo "WORKSPACE_ROOT: $WORKSPACE_ROOT"

# Extract remappings from foundry.toml
REMAPPINGS=$(grep 'remappings' foundry.toml | sed 's/remappings = \[//; s/\]//; s/,//g; s/"//g' | tr '\n' ' ')

# Adjust remappings to use absolute paths
REMAPPINGS=$(echo $REMAPPINGS | sed "s|node_modules|$WORKSPACE_ROOT/nil/node_modules|g")

# Format remappings as individual arguments
REMAPPINGS=$(echo $REMAPPINGS | awk '{for(i=1;i<=NF;i++) print "--remappings", $i}' | tr '\n' ' ')

# Run forge with the remappings as arguments
eval forge "$@" $REMAPPINGS