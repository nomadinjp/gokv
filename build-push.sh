#!/bin/bash

source .env

if [ ! -z "$IMAGE_TAG" ]; then
  docker build --platform=linux/amd64 -t $IMAGE_TAG .
  docker push $IMAGE_TAG
fi