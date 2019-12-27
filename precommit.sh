#!/bin/bash

# abort on errors
set -e

npm run precommit --prefix frontend
gofmt -w api
npm run precommit --prefix docs
npm run precommit --prefix init
npm run precommit --prefix amp
git add -A
