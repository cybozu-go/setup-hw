FROM ghcr.io/cybozu/ubuntu:22.04 AS stage1
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
    && apt-get install -y --no-install-recommends pciutils libargtable2-0 kmod

RUN mkdir /pkg
COPY iDRACTools/racadm/UBUNTU22/x86_64/*.deb /pkg/
RUN dpkg -i /pkg/srvadmin-hapi_11.3.0.0_amd64.deb \
    && dpkg -i /pkg/srvadmin-idracadm7_11.3.0.0_all.deb \
    && dpkg -i /pkg/srvadmin-idracadm8_11.3.0.0_amd64.deb

RUN apt-get clean \
    && rm -rf /pkg \
    && rm -rf /var/lib/apt/lists/*
