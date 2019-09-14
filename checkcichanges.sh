#!/bin/bash

# abort on errors
set -e

changes() {
  git diff --stat --cached -- api/ frontend/
}

travis_ignore="[skip ci]"

if ! changes | grep -E "api/|frontend/" ; then
  echo "no ci changes found"
  sed -i.bak -e "1s/^/$travis_ignore /" ".git/COMMIT_EDITMSG"
else
  echo "ci changes found"
fi
