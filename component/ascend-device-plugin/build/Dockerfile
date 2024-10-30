FROM ubuntu:18.04

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser && \
    usermod root -s /usr/sbin/nologin

ENV LD_LIBRARY_PATH /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

COPY ./device-plugin /usr/local/bin/
COPY ./faultCode.json /usr/local/
COPY ./faultCustomization.json /usr/local/
COPY ./SwitchFaultCode.json /usr/local/

RUN chmod 550 /usr/local/bin/device-plugin &&\
    chmod 550 /usr/local/bin &&\
    chmod 440 /usr/local/faultCode.json &&\
    chmod 440 /usr/local/faultCustomization.json &&\
    chmod 440 /usr/local/SwitchFaultCode.json &&\
    chmod 750 /home/HwHiAiUser &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc