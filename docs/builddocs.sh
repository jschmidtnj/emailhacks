#!/bin/bash

# abort on errors
set -e

# build
yarn api

# remove current directories
rm -rf .vuepress/public/api

# copy to public directories
mv apidocs .vuepress/public/api

# build
yarn generate
yarn pwa
sed -i -e 's/\/images\//\/emailhacks\/images\//g' docs/manifest.json

rm -rf docs/manifest.json-e

rm -rf .vuepress/dist/
mv emailhacks/ .vuepress/dist
rm -rf docs/web/Images/emailhacks
cp -r docs/web/Images/ .vuepress/dist/images
rm -rf docs/web/Images/
mv docs/manifest.json .vuepress/dist/manifest.json
rm -rf docs/
