FROM node:10

LABEL "com.github.actions.name"="Frontend Build"
LABEL "com.github.actions.description"="Sync frontend to S3 bucket"
LABEL "com.github.actions.icon"="refresh-cw"
LABEL "com.github.actions.color"="green"

# Install app dependencies
COPY package*.json /

RUN npm ci --only=production

# Bundle app source
COPY action.js /action.js

ENTRYPOINT [ "node", "action.js" ]
