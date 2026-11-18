# The Ubuntu version must be supported by Dell iDRAC Tools. See docs/maintenance.md.
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

# Do not delete the following line to prevent installation errors in a container environment.
COPY systemctl-dummy /usr/bin/systemctl

RUN apt-get update -y \
    && apt-get install -y --no-install-recommends pciutils libargtable2-0 kmod dmidecode

RUN mkdir /pkg
# The packages are extracted by "make download-idractools".
COPY iDRACTools/racadm/UBUNTU24/x86_64/*.deb /pkg/
RUN dpkg -i /pkg/srvadmin-hapi_*.deb \
    && dpkg -i /pkg/srvadmin-idracadm7_*.deb \
    && dpkg -i /pkg/srvadmin-idracadm8_*.deb

RUN apt-get clean \
    && rm -rf /pkg \
    && rm -rf /var/lib/apt/lists/*
