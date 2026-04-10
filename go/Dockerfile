# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
ARG WORKOS_CLIENT_ID
ARG WORKOS_REDIRECT_URL
ENV VITE_WORKOS_CLIENT_ID=$WORKOS_CLIENT_ID
ENV VITE_WORKOS_REDIRECT_URI=$WORKOS_REDIRECT_URL
COPY web/ampledata-fe/package.json web/ampledata-fe/package-lock.json ./
RUN npm ci
COPY web/ampledata-fe/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/server ./cmd/server

# Stage 3: Final minimal image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /bin/server /bin/server
COPY --from=frontend-builder /frontend/dist /app/static
EXPOSE 8080
CMD ["/bin/server"]
