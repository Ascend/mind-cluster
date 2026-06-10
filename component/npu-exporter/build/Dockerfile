FROM ubuntu:22.04

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser &&\
    usermod root -s /usr/sbin/nologin

COPY ./npu-exporter  /usr/local/bin/
COPY ./metricConfiguration.json  /user/mind-cluster/npu-exporter-config/metricConfiguration.json
COPY ./pluginConfiguration.json  /user/mind-cluster/npu-exporter-config/pluginConfiguration.json
COPY ./agreement.txt /usr/local/

RUN chown root:root /usr/local/bin/npu-exporter  &&\
    chmod 750 -R /home/HwHiAiUser &&\
    chmod 550 /usr/local/bin/ &&\
    chmod 500 /usr/local/bin/npu-exporter &&\
    chmod 440 /user/mind-cluster/npu-exporter-config/metricConfiguration.json &&\
    chmod 440 /user/mind-cluster/npu-exporter-config/pluginConfiguration.json &&\
    chmod 440 /usr/local/agreement.txt &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc
ENV LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/add-ons:/usr/local/Ascend/driver/lib64:/usr/local/dcmi

ENTRYPOINT ["/bin/bash", "-c", "cat /usr/local/agreement.txt; exec /bin/bash"]
