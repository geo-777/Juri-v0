FROM python:3.12-slim

# Create unprivileged user
RUN useradd -m -u 1000 -s /bin/bash runner

# Install GNU time
RUN apt-get update \
    && apt-get install -y --no-install-recommends time \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY runner /usr/local/bin/runner

RUN chmod +x /usr/local/bin/runner

USER runner

CMD ["/usr/local/bin/runner"]