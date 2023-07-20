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

# Setup DELL repository
RUN apt-get update -y && apt-get install -y --no-install-recommends wget gnupg ca-certificates \
     && echo "deb http://linux.dell.com/repo/community/openmanage/${OMSA_VERSION}/${UBUNTU_VERSION} ${UBUNTU_VERSION} main" | tee -a /etc/apt/sources.list.d/linux.dell.com.sources.list \
    && wget https://linux.dell.com/repo/pgp_pubkeys/0x1285491434D8786F.asc \
    && apt-key add 0x1285491434D8786F.asc \
    && rm 0x1285491434D8786F.asc

# To avoid error postinstall script of srvadmin-all in the contaioner enviroment.
COPY systemctl-dummy /usr/bin/systemctl

# Install DELL's tools
# The installation of dcism will failed due to no GUI in the container. Therefore added "exit 0" to avoid stopping job.
# monitor-hw call /usr/libexec/instsvcdrv-helper which depends on "kmod" package. We must install "kmod"
RUN apt-get update -y \
    && ln -s /usr/lib/x86_64-linux-gnu/libssl.so.3 /usr/lib/libssl.so \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends srvadmin-all kmod \
    && rm /usr/bin/systemctl
