FROM debian:12

RUN apt-get update && \
    apt-get install -y build-essential gdb default-jre tmux rsyslog

# to see syslogs, run the following in the container
# apt install rsyslog
# service rsyslog start

WORKDIR /instr
