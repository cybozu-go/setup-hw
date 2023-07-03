FROM ubuntu:22.04 as stage1

COPY setup-hw /usr/local/bin/setup-hw
COPY monitor-hw /usr/local/sbin/monitor-hw
COPY collector /usr/local/sbin/collector
COPY setup-apply-firmware /usr/local/sbin/setup-apply-firmware
COPY setup-isoreboot /usr/local/sbin/setup-isoreboot

CMD ["/usr/local/sbin/monitor-hw"]


# Install vendor tools
FROM stage1

# See https://linux.dell.com/repo/community/openmanage/
ARG OMSA_VERSION=11000
ARG UBUNTU_VERSION=jammy

# 
RUN apt update -y && apt-get install wget pgp apt-utils -y \
    && echo "deb http://linux.dell.com/repo/community/openmanage/${OMSA_VERSION}/${UBUNTU_VERSION} ${UBUNTU_VERSION} main" | tee -a /etc/apt/sources.list.d/linux.dell.com.sources.list \
    && wget https://linux.dell.com/repo/pgp_pubkeys/0x1285491434D8786F.asc \
    && apt-key add 0x1285491434D8786F.asc

# Exit with error due to can't do configure.
RUN apt-get update -y \
    && ln -s /usr/lib/x86_64-linux-gnu/libssl.so.3 /usr/lib/libssl.so \
    && DEBIAN_FRONTEND=noninteractive apt-get install srvadmin-all -y; exit 0

# Enforce the complete of the configure.
RUN cat <<EOF > /var/lib/dpkg/info/srvadmin-hapi.postinst 
#!/bin/bash
/bin/true
EOF

# Enforce the complete of the configure.
RUN cat <<EOF > /var/lib/dpkg/info/srvadmin-idracadm7.postinst
#!/bin/bash
/bin/true
EOF

RUN dpkg --configure -a

ENV PATH /opt/dell/srvadmin/bin:$PATH
