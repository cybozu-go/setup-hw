FROM centos:7 as stage1

COPY setup-hw /usr/local/bin/setup-hw
COPY monitor-hw /usr/local/sbin/monitor-hw
COPY collector /usr/local/sbin/collector
COPY setup-apply-firmware /usr/local/sbin/setup-apply-firmware
COPY setup-isoreboot /usr/local/sbin/setup-isoreboot

CMD ["/usr/local/sbin/monitor-hw"]


# Install vendor tools
FROM stage1

ARG DSU_VERSION=19.07.00

RUN curl -sf -o bootstrap.cgi https://linux.dell.com/repo/hardware/DSU_${DSU_VERSION}/bootstrap.cgi \
    && bash bootstrap.cgi \
    && rm bootstrap.cgi \
    && yum update -y \
    && yum install -y syscfg srvadmin-idracadm7 dmidecode \
    && yum clean all

ENV PATH /opt/dell/srvadmin/bin:$PATH
