FROM ubuntu:18.04

RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&\
    usermod root -s /usr/sbin/nologin

WORKDIR /home/hwMindX

COPY ./noded .

RUN chown -R hwMindX:hwMindX /home/hwMindX &&\
    chmod 500 -R /home/hwMindX/* &&\
    chmod 750 /home/hwMindX &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

USER hwMindX