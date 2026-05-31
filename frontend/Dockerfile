FROM node:22-alpine

RUN npm install -g http-server

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

EXPOSE 8080

CMD ["http-server", "dist", "-s", "-P", "http://main-server:8080"]
