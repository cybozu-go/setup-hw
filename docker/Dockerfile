FROM centos:7 as stage1

# The latest DSU requires systemd.
# The following snippet enables systemd as described in:
# https://github.com/docker-library/docs/tree/master/centos#systemd-integration
ENV container docker
RUN (cd /lib/systemd/system/sysinit.target.wants/; \
     for i in *; do [ $i == systemd-tmpfiles-setup.service ] || rm -f $i; done); \
    rm -f /lib/systemd/system/multi-user.target.wants/*;\
    rm -f /etc/systemd/system/*.wants/*;\
    rm -f /lib/systemd/system/local-fs.target.wants/*; \
    rm -f /lib/systemd/system/sockets.target.wants/*udev*; \
    rm -f /lib/systemd/system/sockets.target.wants/*initctl*; \
    rm -f /lib/systemd/system/basic.target.wants/*;\
    rm -f /lib/systemd/system/anaconda.target.wants/*;

VOLUME [ "/sys/fs/cgroup" ]

CMD ["/usr/sbin/init"]

# COPY setup-hw /usr/local/omsa/bin/setup-hw

# Install vendor tools
FROM stage1

ARG DSU_VERSION=19.01.00

RUN (curl -s http://linux.dell.com/repo/hardware/DSU_${DSU_VERSION}/bootstrap.cgi | bash) \
    && yum update -y \
    && yum install -y syscfg srvadmin-idracadm7 dmidecode \
    && yum clean all \
    && systemctl enable instsvcdrv.service

ENV PATH /opt/dell/srvadmin/bin:$PATH
