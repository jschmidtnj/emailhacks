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
sed -i -e 's/\/images\//\/emailhacks\/images\//g' docs/Polyfills/manifest.json

rm -rf docs/Polyfills/manifest.json-e

rm -rf .vuepress/dist/
mv emailhacks/ .vuepress/dist
rm -rf docs/Polyfills/web/Images/emailhacks
cp -r docs/Polyfills/web/Images/ .vuepress/dist/images
rm -rf docs/Polyfills/web/Images/
mv docs/Polyfills/manifest.json .vuepress/dist/manifest.json
rm -rf docs/
