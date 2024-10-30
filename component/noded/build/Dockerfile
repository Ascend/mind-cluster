FROM ubuntu:18.04

RUN usermod root -s /usr/sbin/nologin

COPY ./noded /usr/local/bin
COPY ./NodeDConfiguration.json /usr/local/
RUN chmod 550 /usr/local/bin/noded &&\
    chmod 550 /usr/local/bin &&\
    chmod 440 /usr/local/NodeDConfiguration.json &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc
