FROM ubuntu:22.04

RUN usermod root -s /usr/sbin/nologin

RUN apt-get update && apt-get install -y libmnl-dev

COPY sriov-cni /usr/local/bin/
COPY agreement.txt /usr/local/
COPY ../images/entrypoint.sh /usr/local

RUN chmod 550 /usr/local/bin/sriov-cni &&\
    chmod 440 /usr/local/agreement.txt &&\
    chmod 755 /usr/local/entrypoint.sh &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc

LABEL io.k8s.display-name="Sriov Cni"

ENTRYPOINT ["/bin/bash", "-c", "cat /usr/local/agreement.txt; exec /usr/local/entrypoint.sh; exec /bin/bash"]
