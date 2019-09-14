#!/bin/bash

# abort on errors
set -e

npm run precommit --prefix frontend
gofmt -w api
npm run precommit --prefix docs
git add -A
