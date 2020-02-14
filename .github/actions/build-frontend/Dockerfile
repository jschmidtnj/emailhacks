FROM node:10

LABEL "com.github.actions.name"="Frontend Build"
LABEL "com.github.actions.description"="Sync frontend to S3 bucket"
LABEL "com.github.actions.icon"="refresh-cw"
LABEL "com.github.actions.color"="green"

# Install app dependencies
# A wildcard is used to ensure both package.json AND package-lock.json are copied
# where available (npm@5+)
COPY package*.json ./

RUN npm install
# If you are building your code for production
# RUN npm ci --only=production

# Bundle app source
COPY action.js ./

ENTRYPOINT [ "node", "action.js" ]
