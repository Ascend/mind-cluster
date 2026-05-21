FROM ubuntu:22.04

RUN usermod root -s /usr/sbin/nologin

COPY ./k8s-rdma-shared-dp /usr/local/bin
COPY ./agreement.txt /usr/local/

RUN chmod 550 /usr/local/bin/k8s-rdma-shared-dp &&\
    chmod 550 /usr/local/bin &&\
    chmod 440 /usr/local/agreement.txt &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc

ENTRYPOINT ["/bin/bash", "-c", "cat /usr/local/agreement.txt; exec /usr/local/bin/k8s-rdma-shared-dp"]
