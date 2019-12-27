#!/bin/bash

# abort on errors
set -e

yarn install
cd frontend
yarn install
cd ../docs
yarn install
cd ../init
yarn install
cd ../amp
yarn install
cd ..
