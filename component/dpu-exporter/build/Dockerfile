FROM ubuntu:22.04

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser &&\
    usermod root -s /usr/sbin/nologin

COPY ./dpu-exporter /usr/local/bin/
COPY ./config.json  /etc/dpu-exporter/config.json

RUN chown root:root /usr/local/bin/dpu-exporter  &&\
    chmod 750 -R /home/HwHiAiUser &&\
    chmod 550 /usr/local/bin/ &&\
    chmod 500 /usr/local/bin/dpu-exporter &&\
    mkdir -p /etc/dpu-exporter &&\
    chmod 440 /etc/dpu-exporter/config.json &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc

ENTRYPOINT ["/bin/bash", "-c", "cat /usr/local/agreement.txt; exec /bin/bash"]
