FROM gcc:14
#installs time for calculating runtime
RUN apt-get update && apt-get install -y time 

WORKDIR /workspace

COPY run.sh /scripts/run.sh

RUN chmod +x /scripts/run.sh

ENTRYPOINT ["/scripts/run.sh"]