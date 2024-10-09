FROM ubuntu:18.04

RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&\
    usermod root -s /usr/sbin/nologin

COPY ./noded /usr/local/bin
COPY ./NodeDConfiguration.json /usr/local/
RUN chmod 550 /usr/local/bin/noded &&\
    chmod 550 /usr/local/bin &&\
    chmod 440 /usr/local/NodeDConfiguration.json &&\
    chmod 750 /home/hwMindX &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc
