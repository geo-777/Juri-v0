FROM python:3.12-slim
#installs time for calculating runtime
RUN apt-get update && apt-get install -y time 
#creating unprivileged user
RUN useradd -m -u 1001 -s /bin/bash runner

WORKDIR /workspace

COPY run.sh /scripts/run.sh

RUN chmod +x /scripts/run.sh

USER runner


ENTRYPOINT ["/scripts/run.sh"]