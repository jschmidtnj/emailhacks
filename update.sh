#!/bin/bash

# abort on errors
set -e

yarn upgrade
cd docs
yarn upgrade
cd ../frontend
yarn upgrade
cd ../api
go get -u -m
cd ..
