FROM node:22-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
#RODRIGO DPS TENS DE ARRANJAR O BUILD PARA PRODUÇÃO, MAS PARA TESTAR LOCALMENTE VOU USAR O DEV
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host"]