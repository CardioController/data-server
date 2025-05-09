FROM node:slim as fe-builder
COPY fe/ /app
WORKDIR /app
ENV PUBLIC_PB_ADDR="/"
COPY fe/package*.json ./
COPY fe/*config* ./
RUN npm ci
COPY fe/. .
RUN npm run build

FROM golang:alpine
WORKDIR /src
COPY main.go go.mod go.sum ./
COPY migrations/ ./migrations
COPY helper/ ./helper
RUN go build -o /bin/pocketbase

FROM alpine:latest
WORKDIR /pb
COPY --from=1 /bin/pocketbase ./
COPY --from=fe-builder /app/build ./pb_public

VOLUME [ "/pb/pb_data" ]

CMD [ "/pb/pocketbase", "serve", "--http=0.0.0.0:8080" ]