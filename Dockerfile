# Use Go 1.22 as the base image for the build stage
FROM golang:1.22-alpine as builder

# Set maintainer information
LABEL maintainer="Seakee <seakee23@gmail.com>"

# Set the working directory for the build
WORKDIR /build

# Copy the go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the project files
COPY . .

# Build the project in the /build directory
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/go-api ./main.go

# Use a smaller base image for the runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the working directory for runtime
WORKDIR /

# Set the build-time argument for TZ
ARG TZ=Asia/Shanghai

# Set the environment variable for TZ
ENV TZ=$TZ

# Install tzdata for timezone support
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone

# Copy the compiled binary file into the /bin directory of the runtime image
COPY --from=builder /build/go-api ./bin

# Copy /bin/lang directory from the builder stage to /bin/lang in the runtime image
COPY --from=builder /build/bin/lang ./bin/lang

# Declare mount points for volumes
VOLUME ["/bin/configs", "/bin/logs"]

# Expose port 8080
EXPOSE 8080

# Set the default command to run when the container starts
ENTRYPOINT ["./bin/go-api"]
