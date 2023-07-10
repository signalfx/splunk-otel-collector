FROM quay.io/splunko11ytest/hadoop:2.10.1

# copy config files
COPY config/* /usr/local/hadoop/etc/hadoop/
COPY config/ssh_config /root/.ssh/config

COPY run_master.sh /usr/local/bin/run_master.sh

