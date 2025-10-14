docker-run := "docker run --rm -v maven_cache:/tmp/maven -v $(pwd):/app -w /app -t -e MAVEN_CONFIG=/tmp/maven/.m2 -e MAVEN_OPTS='-Duser.home=/tmp/maven' maven:3.8"

[private]
default:
	@just --list

# Build the protocal mapper
build:
  @{{docker-run}} mvn clean package

# Check for Maven dependency updates
mvn-dep-check:
  @{{docker-run}} mvn versions:display-dependency-updates 

# Start Keycloak using Docker Compose
dc-up: build
  @docker compose up -d

# Stop and remove Keycloak
dc-down:
  @docker compose down

# Restart Keycloak
dc-restart:
  @docker compose restart

# Tail the Docker Compose logs
follow-logs:
  @docker compose logs -f

# Build the protocol mapper, restart Keycloak and follow logs
build-and-restart: build dc-restart follow-logs

