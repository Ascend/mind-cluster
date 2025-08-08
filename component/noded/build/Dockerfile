FROM ubuntu:22.04

RUN usermod root -s /usr/sbin/nologin

ENV LD_LIBRARY_PATH /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

COPY ./noded /usr/local/bin
COPY ./NodeDConfiguration.json /usr/local/
COPY ./fdConfig.yaml /usr/local/fdConfig.yaml

RUN chmod 550 /usr/local/bin/noded &&\
    chmod 550 /usr/local/bin &&\
    chmod 440 /usr/local/NodeDConfiguration.json &&\
    chmod 440 /usr/local/fdConfig.yaml &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc
