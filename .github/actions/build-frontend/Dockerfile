FROM node:10

LABEL "com.github.actions.name"="Frontend Build"
LABEL "com.github.actions.description"="Sync frontend to S3 bucket"
LABEL "com.github.actions.icon"="refresh-cw"
LABEL "com.github.actions.color"="green"

WORKDIR /github/workspace

# Install app dependencies
COPY package*.json ./

RUN npm ci --only=production

# Bundle app source
COPY action.js ./action.js

COPY build.sh ./build.sh

ENTRYPOINT [ "./build.sh" ]
