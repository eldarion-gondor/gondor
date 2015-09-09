#!/bin/bash

set -ex

echo "$1" > VERSION
git commit -am "bumped version to v$1"
git tag "v$1"
git push --tags

# update homebrew/gondor.rb
sed -i '' "s/[0-9a-f]\{64\}/$(git archive --format=tar v$1 | gzip -c | shasum -a 256 | grep -o '[0-9a-f]\{64\}')/" homebrew/gondor.rb
sed -i '' "s/v[0-9]\{1,2\}\.[0-9]\{1,2\}\.[0-9]\{1,2\}/v$1/" homebrew/gondor.rb
git commit -am "updated homebrew for v$1"

# set everything back to dev
echo "dev" > VERSION
git commit -am "set version to dev"

git push
