FROM ghcr.io/cybozu/ubuntu:24.04 AS stage1
LABEL org.opencontainers.image.source="https://github.com/cybozu-go/setup-hw"

COPY setup-hw /usr/local/bin/setup-hw
COPY monitor-hw /usr/local/sbin/monitor-hw
COPY collector /usr/local/sbin/collector
COPY setup-apply-firmware /usr/local/sbin/setup-apply-firmware
COPY setup-isoreboot /usr/local/sbin/setup-isoreboot
CMD ["/usr/local/sbin/monitor-hw"]

# Install vendor tools
FROM stage1
LABEL org.opencontainers.image.source="https://github.com/cybozu-go/setup-hw"

# See https://linux.dell.com/repo/community/openmanage/
ARG OMSA_VERSION=11010
ARG UBUNTU_VERSION=jammy

# Setup DELL repository
RUN apt-get update -y && apt-get install -y --no-install-recommends wget gnupg ca-certificates curl\
    && echo "deb http://linux.dell.com/repo/community/openmanage/${OMSA_VERSION}/${UBUNTU_VERSION} ${UBUNTU_VERSION} main" | tee -a /etc/apt/sources.list.d/linux.dell.com.sources.list \
    && wget https://linux.dell.com/repo/pgp_pubkeys/0x1285491434D8786F.asc \
    && apt-key add 0x1285491434D8786F.asc \
    && rm 0x1285491434D8786F.asc

# To avoid error postinstall script of srvadmin-all in the contaioner enviroment.
COPY systemctl-dummy /usr/bin/systemctl

# Install DELL's tools
# monitor-hw call /usr/libexec/instsvcdrv-helper which depends on "kmod" package. We must install "kmod"
RUN apt-get update -y \
    && ln -s /usr/lib/x86_64-linux-gnu/libssl.so.3 /usr/lib/libssl.so \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends srvadmin-all kmod \
    && rm /usr/bin/systemctl
