#!/bin/bash

# Configuration
# Usage: ./deploy.sh [VERSION_TAG]
VERSION=${1:-"dev-$(date +%s)"}
IMAGE_NAME="free5gc-nwdaf"
REMOTE_HOST="ns"
SSH_CONFIG="../vagrant/ssh_config"
LOCAL_CHART_PATH="charts/free5gc-nwdaf"
REMOTE_CHART_PATH="/tmp/free5gc-nwdaf"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting deployment pipeline for version: $VERSION${NC}"

# 1. Build Docker Image
echo -e "${GREEN}Building Docker image...${NC}"
docker build -t docker.io/library/$IMAGE_NAME:$VERSION .
if [ $? -ne 0 ]; then echo -e "${RED}Build failed${NC}"; exit 1; fi

# 2. Save Image to Tar
echo -e "${GREEN}Saving image to tar...${NC}"
TEMP_TAR="/tmp/$IMAGE_NAME-$VERSION.tar"
docker save docker.io/library/$IMAGE_NAME:$VERSION -o $TEMP_TAR
if [ $? -ne 0 ]; then echo -e "${RED}Save failed${NC}"; exit 1; fi

# 3. Transfer Image and Chart
echo -e "${GREEN}Transferring to VM...${NC}"
# Transfer image
scp -F $SSH_CONFIG $TEMP_TAR $REMOTE_HOST:/tmp/
if [ $? -ne 0 ]; then echo -e "${RED}Image SCP failed${NC}"; exit 1; fi

# Transfer chart (to ensure latest templates/values are used)
echo -e "${GREEN}Syncing Helm Chart...${NC}"
# We assume remote structure exists, but let's just make sure we copy the content
# Taring local chart to send it efficiently
tar czf /tmp/chart-update.tar.gz -C ./charts/free5gc-nwdaf .
scp -F $SSH_CONFIG /tmp/chart-update.tar.gz $REMOTE_HOST:/tmp/
rm /tmp/chart-update.tar.gz

# 4. Execute Remote Commands
echo -e "${GREEN}Deploying to microk8s...${NC}"
ssh -F $SSH_CONFIG $REMOTE_HOST << EOF
  set -e
  
  # Import Image
  echo "Importing image ($VERSION)..."
  microk8s.ctr image import /tmp/$IMAGE_NAME-$VERSION.tar
  rm /tmp/$IMAGE_NAME-$VERSION.tar

  # Update Chart
  echo "Updating chart files..."
  mkdir -p $REMOTE_CHART_PATH
  tar xzf /tmp/chart-update.tar.gz -C $REMOTE_CHART_PATH
  rm /tmp/chart-update.tar.gz

  # Upgrade Helm Release
  echo "Upgrading Helm release..."
  microk8s helm3 upgrade nwdaf $REMOTE_CHART_PATH -n free5gc \\
    --set nwdaf.image.name=docker.io/library/$IMAGE_NAME \\
    --set nwdaf.image.tag=$VERSION \\
    --set nwdaf.image.pullPolicy=Never \\
    --set nwdaf.initImagePullPolicy=Never \\
    --set initcontainers.curl.registry=docker.io/library \\
    --set initcontainers.curl.image=$IMAGE_NAME \\
    --set initcontainers.curl.tag=$VERSION \\
    --set nwdaf.prometheusUrl=http://prometheus-prometheus.free5gc:9090 \\
    --set nwdaf.nefUrl=http://10.152.183.162:80 \\
    --set nwdaf.ollamaUrl=http://192.168.0.149:11434

  # Restart pods to ensure new version is picked up immediately
  echo "Restarting pods..."
  microk8s kubectl delete pods -n free5gc -l app.kubernetes.io/name=free5gc-nwdaf --ignore-not-found=true
EOF

# Cleanup local tar
rm $TEMP_TAR

echo -e "${GREEN}Deployment Complete!${NC}"
echo -e "${GREEN}   Version: $VERSION${NC}"
echo -e "   Check logs with: ssh -F vagrant/ssh_config ns 'microk8s kubectl logs -n free5gc -l app.kubernetes.io/name=free5gc-nwdaf --tail=50 -f'"
