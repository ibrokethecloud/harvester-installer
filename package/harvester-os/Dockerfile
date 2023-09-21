ARG BASE_OS_IMAGE
FROM ${BASE_OS_IMAGE}

RUN curl -sfL https://github.com/rancher/wharfie/releases/latest/download/wharfie-amd64  -o /usr/bin/wharfie && chmod +x /usr/bin/wharfie

COPY files/ /
RUN chmod 0600 /system/oem/*

COPY harvester-release.yaml /etc/

RUN zypper ar http://download.opensuse.org/distribution/leap/15.4/repo/oss leap && \
    zypper ar http://download.opensuse.org/update/leap/15.4/oss update_leap && \
    zypper ar https://download.opensuse.org/update/leap/15.4/sle update_sle && \
    zypper --gpg-auto-import-keys --non-interactive in  gawk gcc python kmod kernel-source make && \
    zypper clean -a
RUN mkinitrd

ARG HARVESTER_PRETTY_NAME
RUN sed -i "s/^PRETTY_NAME.*/PRETTY_NAME=\"$HARVESTER_PRETTY_NAME\"/g" /etc/os-release && \
    echo GRUB_ENTRY_NAME="\"$HARVESTER_PRETTY_NAME\"" >> /etc/os-release

