FROM node:{{VERSION}}
ENV PORT {{PORT}}
EXPOSE {{PORT}}

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
COPY package.json .
RUN npm install
COPY . .

CMD ["npm", "start"]
