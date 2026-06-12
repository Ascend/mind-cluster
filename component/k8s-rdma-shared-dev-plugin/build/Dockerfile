FROM ubuntu:22.04

RUN usermod root -s /usr/sbin/nologin

COPY k8s-rdma-shared-dp /usr/local/bin/
COPY agreement.txt /usr/local/
COPY fault_code.json /etc/rdma-plugin/
COPY fault_detection.sh /etc/rdma-plugin/

RUN chmod 550 /usr/local/bin/k8s-rdma-shared-dp &&\
    chmod 440 /usr/local/agreement.txt &&\
    chmod 550 /etc/rdma-plugin/fault_detection.sh &&\
    sed -i 's/\r$//' /etc/rdma-plugin/fault_detection.sh &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc

WORKDIR /etc/rdma-plugin

LABEL io.k8s.display-name="RDMA Shared Device Plugin"

ENTRYPOINT ["/bin/bash", "-c", "cat /usr/local/agreement.txt; exec /bin/bash"]
