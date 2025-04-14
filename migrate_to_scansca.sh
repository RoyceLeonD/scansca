#!/bin/bash

# Script to migrate the project structure from Antska to Scansca
#
# This script performs the following consolidations:
#
# 1. Project name change: from "Antska" to "Scansca"
# 2. Directory structure consolidation:
#    - cmd/antska-server/ → cmd/server/
#    - internal/aml/ → internal/sml/ (Scansca Management Layer)
# 3. Import path updates: github.com/royceleond/antska → github.com/royceleond/scansca
# 4. Environment variable prefix changes: ANTSKA_ → SCANSCA_
# 5. Configuration file updates

# Create directory structure
mkdir -p /home/royceld/Programming/Personal/scansca/{cmd/server,internal/sml/{registry,scheduler,state},internal/{connectors,mci,mcp},pkg,documentation,config,docker}

# Copy main files
cp -r /home/royceld/Programming/Personal/antska/README.md /home/royceld/Programming/Personal/scansca/
cp -r /home/royceld/Programming/Personal/antska/go.mod /home/royceld/Programming/Personal/scansca/
cp -r /home/royceld/Programming/Personal/antska/go.sum /home/royceld/Programming/Personal/scansca/
cp -r /home/royceld/Programming/Personal/antska/main.go /home/royceld/Programming/Personal/scansca/
cp -r /home/royceld/Programming/Personal/antska/Makefile /home/royceld/Programming/Personal/scansca/

# Copy cmd directory
cp -r /home/royceld/Programming/Personal/antska/cmd/server/* /home/royceld/Programming/Personal/scansca/cmd/server/

# Copy internal directories
cp -r /home/royceld/Programming/Personal/antska/internal/connectors /home/royceld/Programming/Personal/scansca/internal/
cp -r /home/royceld/Programming/Personal/antska/internal/mci /home/royceld/Programming/Personal/scansca/internal/
cp -r /home/royceld/Programming/Personal/antska/internal/mcp /home/royceld/Programming/Personal/scansca/internal/
cp -r /home/royceld/Programming/Personal/antska/internal/sml/* /home/royceld/Programming/Personal/scansca/internal/sml/

# Copy documentation
cp -r /home/royceld/Programming/Personal/antska/documentation/* /home/royceld/Programming/Personal/scansca/documentation/

# Copy config
cp -r /home/royceld/Programming/Personal/antska/config/scansca.yaml /home/royceld/Programming/Personal/scansca/config/

# Copy docker files
cp -r /home/royceld/Programming/Personal/antska/docker/* /home/royceld/Programming/Personal/scansca/docker/

# Copy pkg directory if it exists
if [ -d "/home/royceld/Programming/Personal/antska/pkg" ]; then
  cp -r /home/royceld/Programming/Personal/antska/pkg/* /home/royceld/Programming/Personal/scansca/pkg/
fi

echo "Migration to Scansca completed!"