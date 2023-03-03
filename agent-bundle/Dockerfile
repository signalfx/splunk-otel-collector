# Copyright 2020 Splunk, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG ARCH=amd64
ARG JDK_VERSION=11.0.18_10
ARG MAVEN_VERSION=3.6.0
ARG PYTHON_VERSION=3.8.10
ARG PIP_VERSION=21.0.1
ARG COLLECTD_VERSION=5.8.0-sfx0
ARG COLLECTD_COMMIT=4d3327b14cf4359029613baf4f90c4952702105e
ARG DOCKER_REPO=docker.io

######### Java monitor dependencies and monitor jar compilation
FROM ${DOCKER_REPO}/ubuntu:18.04 as java

ARG ARCH
ARG JDK_VERSION
ARG MAVEN_VERSION

ENV DEBIAN_FRONTEND noninteractive

RUN apt update &&\
    apt install -y wget

ENV OPENJDK_BASE_URL="https://github.com/adoptium/temurin11-binaries/releases/download"

RUN ENCODED_VER=$(echo $JDK_VERSION | sed 's/_/%2B/g') && \
    if [ "$ARCH" = "amd64" ]; then \
    OPENJDK_URL="${OPENJDK_BASE_URL}/jdk-${ENCODED_VER}/OpenJDK11U-jdk_x64_linux_hotspot_${JDK_VERSION}.tar.gz"; \
    elif [ "$ARCH" = "arm64" ]; then \
    OPENJDK_URL="${OPENJDK_BASE_URL}/jdk-${ENCODED_VER}/OpenJDK11U-jdk_aarch64_linux_hotspot_${JDK_VERSION}.tar.gz"; \
    else \
    OPENJDK_URL="${OPENJDK_BASE_URL}/jdk-${ENCODED_VER}/OpenJDK11U-jdk_${ARCH}_linux_hotspot_${JDK_VERSION}.tar.gz"; \
    fi && \
    wget -O /tmp/openjdk.tar.gz "$OPENJDK_URL"

RUN mkdir -p /opt/root && \
    tar -C /opt/root -xzf /tmp/openjdk.tar.gz && \
    mv /opt/root/jdk* /opt/root/jdk && \
    rm -f /tmp/openjdk.tar.gz

ENV JAVA_HOME /opt/root/jdk

RUN mkdir -p /opt/root/jre && \
    rm -f ${JAVA_HOME}/lib/src.zip && \
    cp -rL ${JAVA_HOME}/bin /opt/root/jre/ && \
    cp -rL ${JAVA_HOME}/lib /opt/root/jre/ && \
    cp -rL ${JAVA_HOME}/conf /opt/root/jre/

ENV MAVEN_BASE_URL="https://archive.apache.org/dist/maven/maven-3"

RUN wget -O /tmp/maven.tar.gz ${MAVEN_BASE_URL}/${MAVEN_VERSION}/binaries/apache-maven-${MAVEN_VERSION}-bin.tar.gz && \
    tar -C /tmp -xzf /tmp/maven.tar.gz && \
    rm -f /tmp/maven.tar.gz

ENV PATH=$PATH:/tmp/apache-maven-${MAVEN_VERSION}/bin

COPY java/ /usr/src/agent-java/
RUN cd /usr/src/agent-java/runner &&\
    mvn -V clean install

RUN cd /usr/src/agent-java/jmx &&\
    mvn -V clean package

RUN rm -rf /tmp/apache-maven*


###### Python bundle builder image ######
FROM ${DOCKER_REPO}/ubuntu:18.04 as python

ARG ARCH
ARG PYTHON_VERSION

ENV DEBIAN_FRONTEND noninteractive

RUN sed -i -e '/^deb-src/d' /etc/apt/sources.list &&\
    apt-get update &&\
    apt-get install -y \
      curl \
      dpkg \
      net-tools \
      software-properties-common \
      wget \
      autoconf \
      automake \
      autotools-dev \
      bison \
      build-essential \
      debhelper \
      debian-archive-keyring \
      debootstrap \
      devscripts \
      dh-make \
      dpatch \
      fakeroot \
      flex \
      gcc \
      git-core \
      iptables-dev \
      libatasmart-dev \
      libcurl4-openssl-dev \
      libdbi0-dev \
      libdbus-1-dev \
      libdbus-glib-1-dev \
      libdistro-info-perl \
      libesmtp-dev \
      libexpat1-dev \
      libffi-dev \
      libganglia1-dev \
      libgcrypt11-dev \
      libglib2.0-dev \
      libldap2-dev \
      libltdl-dev \
      libmemcached-dev \
      libmicrohttpd-dev \
      libmnl-dev \
      libmodbus-dev \
      libnotify-dev \
      liboping-dev \
      libow-dev \
      libpcap-dev \
      libperl-dev \
      libpq-dev \
      libprotobuf-c0-dev \
      librabbitmq-dev \
      librdkafka-dev \
      librrd-dev \
      libsensors4-dev \
      libsnmp-dev \
      libssl-dev \
      libtool \
      libudev-dev \
      libvarnishapi-dev \
      libvirt-dev \
      libxml2-dev \
      libyajl-dev \
      lsb-release \
      pbuilder \
      pkg-config \
      po-debconf \
      protobuf-c-compiler \
      quilt \
      zlib1g-dev

RUN wget https://dev.mysql.com/get/mysql-apt-config_0.8.12-1_all.deb && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 467B942D3A79BD29 && \
    dpkg -i mysql-apt-config_0.8.12-1_all.deb && \
    rm -f mysql-apt-config_0.8.12-1_all.deb && \
    apt-get update && apt-get install -y libmysqlclient-dev libcurl4-gnutls-dev

RUN wget -O /tmp/Python-${PYTHON_VERSION}.tgz https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz &&\
    cd /tmp &&\
    tar xzf Python-${PYTHON_VERSION}.tgz && \
    cd Python-${PYTHON_VERSION} && \
    ./configure --enable-shared --enable-ipv6 --with-system-ffi --with-system-expat && \
    make -j`nproc` && make install libinstall && \
    ldconfig

RUN rm -rf /tmp/Python*


###### Collectd builder image ######
FROM python as collectd

ARG ARCH
ARG COLLECTD_VERSION
ARG COLLECTD_COMMIT

RUN cd /tmp &&\
    wget https://github.com/signalfx/collectd/archive/${COLLECTD_COMMIT}.tar.gz &&\
    tar -xvf ${COLLECTD_COMMIT}.tar.gz &&\
    mkdir -p /usr/src/ &&\
    mv collectd-${COLLECTD_COMMIT}* /usr/src/collectd &&\
    rm -f ${COLLECTD_COMMIT}.tar.gz

# Hack to get our custom version compiled into collectd
RUN echo "#!/bin/bash" > /usr/src/collectd/version-gen.sh &&\
    echo "printf \${COLLECTD_VERSION//-/.}" >> /usr/src/collectd/version-gen.sh

COPY --from=java /opt/root/jdk/ /opt/root/jdk/

WORKDIR /usr/src/collectd

ARG extra_cflags="-O2"
ENV CFLAGS "-Wno-deprecated-declarations -Wno-format-truncation -fPIC $extra_cflags"
ENV CXXFLAGS $CFLAGS
ENV JAVA_HOME /opt/root/jdk

# In the bundle, the java plugin so will live in /lib/collectd and the JVM
# exists at /jre
ENV JAVA_LDFLAGS "-Wl,-rpath -Wl,\$\$\ORIGIN/../../jre/lib/server"

# xencpu plugin is not supported on ppc64le
RUN DISABLE_XENCPU=""; if [ "$ARCH" = "ppc64le" ]; then DISABLE_XENCPU="--disable-xencpu"; fi; \
    autoreconf -vif &&\
    ./configure $DISABLE_XENCPU \
        --prefix="/usr" \
        --localstatedir="/var" \
        --sysconfdir="/etc/collectd" \
        --enable-all-plugins \
        --disable-apple_sensors \
        --disable-aquaero \
        --disable-barometer \
        --disable-dpdkstat \
        --disable-dpdkevents \
        --disable-gps \
        --disable-grpc \
        --disable-intel_pmu \
        --disable-intel_rdt \
        --disable-lpar \
        --disable-lua \
        --disable-lvm \
        --disable-mic \
        --disable-mqtt \
        --disable-netapp \
        --disable-nut \
        --disable-ipmi \
        --disable-oracle \
        --disable-pf \
        --disable-redis \
        --disable-routeros \
        --disable-sigrok \
        --disable-tape \
        --disable-tokyotyrant \
        --disable-turbostat \
        --disable-write_mongodb \
        --disable-write_redis \
        --disable-write_riemann \
        --disable-xmms \
        --disable-zone \
        --without-libstatgrab \
        --disable-silent-rules \
        --disable-static \
        --with-java=${JAVA_HOME} \
        LIBPYTHON_LDFLAGS="$(python3.8-config --ldflags) -lpython3.8" \
        LIBPYTHON_CPPFLAGS="$(python3.8-config --includes)" \
        LIBPYTHON_LIBS="$(python3.8-config --libs) -lpython3.8"

# Compile all of collectd first, including plugins
RUN make -j`nproc` &&\
    make install

COPY scripts/collect-libs /opt/collect-libs
RUN /opt/collect-libs /opt/deps /usr/sbin/collectd /usr/lib/collectd/

# Compile patchelf statically from source
RUN cd /tmp &&\
    wget https://nixos.org/releases/patchelf/patchelf-0.11/patchelf-0.11.tar.gz &&\
    tar -xf patchelf*.tar.gz &&\
    cd patchelf-0.11* &&\
    ./configure LDFLAGS="-static" &&\
    make &&\
    make install

RUN rm -rf /tmp/patchelf*

# For some reason libvarnishapi doesn't properly depend on libm, so make it
# right.
RUN patchelf --add-needed libm-2.23.so /opt/deps/libvarnishapi.so.1.0.6


###### Python Plugin Image ######
FROM collectd as python-plugins

ARG PIP_VERSION

RUN python3 -m pip install --upgrade pip==$PIP_VERSION

# rust is required for compiling cryptography on ppc64le
RUN if [ "$ARCH" = "ppc64le" ]; then apt-get install -y cargo rustc; fi

# Mirror the same dir structure that exists in the original source
COPY scripts/get-collectd-plugins.py /opt/scripts/
COPY scripts/requirements.txt /opt/
COPY collectd-plugins.yaml /opt/

RUN python3 -m pip install -r /opt/requirements.txt

RUN mkdir -p /opt/collectd-python &&\
    python3 -u /opt/scripts/get-collectd-plugins.py /opt/collectd-python

COPY python/ /opt/sfxpython/
RUN cd /opt/sfxpython && python3 -m pip install .

RUN python3 -m pip list

# Remove pip to avoid usage in python monitor and in exec'd container
RUN python3 -m pip uninstall pip -y
# Delete all bundled wheels
RUN find /usr/local/lib/python3.8 -name "*.whl" -delete

# Delete all compiled python to save space
RUN find /usr/local/lib/python3.8 -name "*.pyc" -o -name "*.pyo" | xargs rm
# We don't support compiling extension modules so don't need this directory
RUN rm -rf /usr/local/lib/python3.8/config-*-linux-gnu

RUN find /usr/local/lib/python3.8 -type d -name __pycache__ | xargs rm -rf {} \;
RUN find /usr/local/lib/python3.8 -wholename "*test*.key" -delete -or -wholename "*test*.pem" -delete


####### Extra packages that don't make sense to pull down in any other stage ########
FROM ${DOCKER_REPO}/ubuntu:18.04 as extra-packages

ENV DEBIAN_FRONTEND noninteractive

RUN apt update &&\
    apt install -y \
        curl \
        host \
        iproute2 \
        netcat \
        netcat.openbsd

COPY scripts/collect-libs /opt/collect-libs

ENV useful_bins=" \
  /bin/bash \
  /bin/cat \
  /bin/cp \
  /bin/date \
  /bin/echo \
  /bin/grep \
  /bin/kill \
  /bin/ln \
  /bin/ls \
  /bin/mkdir \
  /bin/mount \
  /bin/nc \
  /bin/ps \
  /bin/rm \
  /bin/sh \
  /bin/ss \
  /bin/tar \
  /bin/umount \
  /usr/bin/curl \
  /usr/bin/dirname \
  /usr/bin/find \
  /usr/bin/host \
  /usr/bin/realpath \
  /usr/bin/tail \
  "
RUN mkdir -p /opt/root/lib &&\
    /opt/collect-libs /opt/root/lib ${useful_bins}

RUN mkdir -p /opt/root/bin &&\
    cp $useful_bins /opt/root/bin

# Gather all our bins/libs and set rpath on the properly.  Interpreter has to
# be set at runtime (or in the final docker stage for docker runs).
COPY --from=collectd /usr/local/bin/patchelf /usr/bin/

# Gather Python dependencies
COPY --from=python-plugins /usr/local/lib/python3.8 /opt/root/lib/python3.8
COPY --from=python-plugins /usr/local/lib/libpython3.8.so.1.0 /opt/root/lib
COPY --from=python-plugins /usr/local/bin/python3.8 /opt/root/bin/python

# Gather compiled collectd plugin libraries
COPY --from=collectd /usr/sbin/collectd /opt/root/bin/collectd
COPY --from=collectd /opt/deps/ /opt/root/lib/
COPY --from=collectd /usr/lib/collectd/*.so /opt/root/lib/collectd/

COPY --from=java /opt/root/jre/ /opt/root/jre/

COPY scripts/patch-rpath /usr/bin/
RUN patch-rpath /opt/root


###### Final Agent Image #######
# This build stage is meant as the final target when running the agent in a
# container environment (e.g. directly with Docker or on K8s).  The stages
# below this are special-purpose.
FROM scratch as final-image

COPY --from=collectd /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENV SSL_CERT_FILE /etc/ssl/certs/ca-certificates.crt

COPY --from=collectd /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=collectd /usr/local/bin/patchelf /bin/

# Pull in the Linux dynamic link loader at a fixed path across all
# architectures.  Binaries will later be patched to use this interpreter
# natively.
COPY --from=extra-packages /lib/*-linux-gnu/ld-2.27.so /bin/ld-linux.so

# Java dependencies
COPY --from=extra-packages /opt/root/jre/ /jre
COPY --from=java /usr/src/agent-java/jmx/target/agent-jmx-monitor-1.0-SNAPSHOT.jar /lib/jmx-monitor.jar

COPY --from=extra-packages /opt/root/lib/ /lib/
COPY --from=extra-packages /opt/root/bin/ /bin/

# Some extra non-binary collectd resources
COPY --from=collectd /usr/share/collectd/postgresql_default.conf /postgresql_default.conf
COPY --from=collectd /usr/share/collectd/types.db /types.db
COPY --from=collectd /usr/share/collectd/java/ /collectd-java/
COPY signalfx_types.db /signalfx_types.db

# Pull in Python collectd plugin scripts
COPY --from=python-plugins /opt/collectd-python/ /collectd-python/

COPY scripts/umount-hostfs-non-persistent /bin/umount-hostfs-non-persistent
COPY scripts/patch-interpreter /bin/patch-interpreter
RUN ["/bin/ld-linux.so", "/bin/sh", "/bin/patch-interpreter", "/"]

RUN mkdir -p /run/collectd /var/run

WORKDIR /

CMD ["/bin/bash"]
