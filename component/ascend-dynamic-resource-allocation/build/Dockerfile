FROM ubuntu:22.04

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser && \
    usermod root -s /usr/sbin/nologin

ENV LD_LIBRARY_PATH /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

COPY ./ascend-dra /usr/local/bin/
COPY ./agreement.txt /usr/local/

RUN chmod 550 /usr/local/bin/ascend-dra && \
    chmod 550 /usr/local/bin && \
    chmod 440 /usr/local/agreement.txt && \
    chmod 750 /home/HwHiAiUser && \
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc

ENTRYPOINT ["/bin/bash", "-c", "cat /usr/local/agreement.txt; exec /bin/bash"]
