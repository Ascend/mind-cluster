FROM ubuntu:22.04

RUN usermod root -s /usr/sbin/nologin

RUN apt-get update && apt-get install -y libmnl-dev

COPY sriov-network-device-plugin /usr/local/bin/
COPY agreement.txt /usr/local/

RUN chmod 550 /usr/local/bin/sriov-network-device-plugin &&\
    chmod 440 /usr/local/agreement.txt &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc

WORKDIR /etc/rdma-plugin

LABEL io.k8s.display-name="Sriov Network Device Plugin"

ENTRYPOINT ["/bin/bash", "-c", "cat /usr/local/agreement.txt; exec /bin/bash"]
