#!/bin/bash

# Docker Build and Push Script for Little Einstein API
# Usage: ./build-and-push.sh

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="nfarley/api-image"
TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"

echo -e "${BLUE}🚀 Building and pushing Little Einstein API container${NC}"
echo -e "${BLUE}Image: ${FULL_IMAGE_NAME}${NC}"
echo ""

# Step 1: Docker login (with sudo)
echo -e "${YELLOW}📝 Step 1: Docker login${NC}"
sudo docker login
echo ""

# Step 2: Build the Docker image
echo -e "${YELLOW}🔨 Step 2: Building Docker image${NC}"
sudo docker build -t "${FULL_IMAGE_NAME}" .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Docker build successful${NC}"
else
    echo -e "${RED}❌ Docker build failed${NC}"
    exit 1
fi
echo ""

# Step 3: Push the Docker image
echo -e "${YELLOW}📤 Step 3: Pushing Docker image to Docker Hub${NC}"
sudo docker push "${FULL_IMAGE_NAME}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Docker push successful${NC}"
else
    echo -e "${RED}❌ Docker push failed${NC}"
    exit 1
fi
echo ""

# Step 4: Clean up local images (optional)
echo -e "${YELLOW}🧹 Step 4: Clean up local images${NC}"
read -p "Do you want to remove all local images for this project to save space? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Remove the just-built image
    sudo docker rmi "${FULL_IMAGE_NAME}" 2>/dev/null || true

    # Remove all other nfarley/api-image images (including untagged ones)
    echo -e "${BLUE}Removing all nfarley/api-image images...${NC}"
    sudo docker images --format "table {{.Repository}}:{{.Tag}}\t{{.ID}}" | grep "nfarley/api-image" | grep -v "REPOSITORY" | while read line; do
        IMAGE_ID=$(echo "$line" | awk '{print $2}')
        if [ ! -z "$IMAGE_ID" ]; then
            sudo docker rmi "$IMAGE_ID" 2>/dev/null || true
        fi
    done

    # Clean up any dangling images
    sudo docker image prune -f >/dev/null 2>&1 || true

    echo -e "${GREEN}✅ All local project images removed${NC}"
else
    echo -e "${BLUE}ℹ️  Local images kept${NC}"
fi
echo ""

# Step 5: Show completion
echo -e "${GREEN}🎉 Build and push completed successfully!${NC}"
echo -e "${BLUE}Image pushed: ${FULL_IMAGE_NAME}${NC}"
echo -e "${BLUE}You can now use this image in Azure App Service${NC}"
echo ""

# Optional: Show docker images
echo -e "${YELLOW}📋 Current local Docker images:${NC}"
sudo docker images | grep "nfarley/api-image" || echo "No local images found for nfarley/api-image"