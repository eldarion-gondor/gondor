#!/bin/bash

set -ex

echo "$1" > VERSION
git commit -am "bumped version to v$1"
git tag "v$1"
git push --tags

echo "dev" > VERSION
git commit -am "set version to dev"
git push
