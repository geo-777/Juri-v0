FROM eclipse-temurin:17-jdk-jammy

# Create unprivileged user
RUN useradd -m -u 1001 -s /bin/bash runner

# Install /usr/bin/time
RUN apt-get update \
    && apt-get install -y --no-install-recommends time \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

USER runner

# Keep container alive for Docker SDK Exec
CMD ["sleep", "infinity"]