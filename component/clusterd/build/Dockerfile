FROM ubuntu:22.04

RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&\
    usermod root -s /usr/sbin/nologin

COPY ./clusterd /usr/local/bin
COPY ./relationFaultCustomization.json /home/hwMindX/relationFaultCustomization.json
COPY ./faultDuration.json /home/hwMindX/faultDuration.json
COPY ./publicFaultConfiguration.json /home/hwMindX/publicFaultConfiguration.json
COPY ./fdConfig.yaml /home/hwMindX/fdConfig.yaml

RUN chown -R hwMindX:hwMindX /home/hwMindX &&\
    chmod 555 /usr/local/bin/clusterd &&\
    chmod 750 /home/hwMindX &&\
    chmod 440 /home/hwMindX/relationFaultCustomization.json &&\
    chmod 440 /home/hwMindX/faultDuration.json &&\
    chmod 440 /home/hwMindX/publicFaultConfiguration.json &&\
    chmod 440 /home/hwMindX/fdConfig.yaml &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

# hwMindX is used as the default user of the container
USER hwMindX