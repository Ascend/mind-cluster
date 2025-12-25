FROM ubuntu:22.04 as build

RUN useradd -d /home/hwMindX -u 9000 -m -s  /usr/sbin/nologin hwMindX && \
    usermod root -s /usr/sbin/nologin

COPY ./rl-operator  /usr/local/bin/

RUN chown -R hwMindX:hwMindX /usr/local/bin/rl-operator  &&\
    chmod 500 /usr/local/bin/rl-operator &&\
    chmod 750 /home/hwMindX &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

# use hwMindX as default user of container
USER hwMindX

