# --- Etapa 1: compilar el binario ---
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/cortaurl .

# --- Etapa 2: imagen final mínima ---
FROM alpine:latest
COPY --from=build /bin/cortaurl /bin/cortaurl
EXPOSE 8080
ENTRYPOINT ["/bin/cortaurl"]
