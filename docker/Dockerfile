FROM centos:7 as stage1

COPY setup-hw /usr/local/bin/setup-hw

CMD ["/bin/sleep", "infinity"]


# Install vendor tools
FROM stage1

ARG DSU_VERSION=19.01.00

RUN (curl -sf http://linux.dell.com/repo/hardware/DSU_${DSU_VERSION}/bootstrap.cgi | bash) \
    && yum update -y \
    && yum install -y syscfg srvadmin-idracadm7 dmidecode \
    && yum clean all

ENV PATH /opt/dell/srvadmin/bin:$PATH

CMD ["/bin/sh", "-c", "/usr/libexec/instsvcdrv-helper start && /bin/sleep infinity"]
