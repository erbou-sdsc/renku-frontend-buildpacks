#!/usr/bin/env bash
set -xeou pipefail

# The image name excluding the tag
IMAGE=$1
# The location where the builder is located
BUILDER_IMAGE=$2
# The path of the image definition
IMAGE_PATH=$3
# The front end
FRONTEND=$4
# If set to anything the image will be pushed to the image repository
PUBLISH=$5

TAG=$(git describe --exact-match --tags)
if [[ -z $TAG ]]; then
	TAG=$(git rev-parse HEAD | cut -c -7)
fi
echo "Building sample image $IMAGE:$TAG using builder $BUILDER_IMAGE"

if [[ -z $PUBLISH ]]; then
	echo "Will not publish the push the image to the repository"
	pack build $IMAGE:$TAG --clear-cache --path ${IMAGE_PATH} --env BP_RENKU_FRONTENDS=${FRONTEND} --builder ${BUILDER_IMAGE}:${TAG}
else
	echo "Found publish flag, will push the image to the repo"
	pack build $IMAGE:$TAG --clear-cache --path ${IMAGE_PATH} --env BP_RENKU_FRONTENDS=${FRONTEND} --builder ${BUILDER_IMAGE}:${TAG} --publish
fi
