# ==============================================================================
# MindCluster Dev Container Dockerfile
# Base: Ubuntu 20.04 (glibc 2.31, satisfies VS Code Server >= 2.28)
# Covers: Go / C-C++ / Python / volcano(musl) / helm-deploy-tool
# Toolchain versions are aligned with the CI environment.
#   - gcc-7.5.0 via apt (focal's gcc-7 package) to match CI .comment=7.5.0
#   - musl 1.2.0 / cmake 3.20.5 / Python 3.9.11 built from source
# ==============================================================================
FROM ubuntu:20.04

# Avoid interactive prompts during apt operations
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai

# ------------------------------------------------------------------------------
# 1. Base system toolchain (apt)
#    Ubuntu 20.04 (focal) ships glibc 2.31 (satisfies VS Code Server) and
#    gcc-9 by default. CI requires gcc-7.5.0 (artifact .comment=7.5.0), so
#    gcc-7/g++-7 are installed and pinned as the default cc/c++ via
#    update-alternatives. cmake 3.20.5 and Python 3.9.11 are installed in
#    later steps to match CI.
# ------------------------------------------------------------------------------
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        gcc-7 \
        g++-7 \
        make \
        git \
        wget \
        curl \
        unzip \
        dos2unix \
        zip \
        ca-certificates \
        libnuma-dev \
        zlib1g-dev \
        libssl-dev \
        libffi-dev \
        libbz2-dev \
        libsqlite3-dev \
        libreadline-dev \
        libncursesw5-dev \
        libgdbm-dev \
        liblzma-dev \
        tk-dev \
        uuid-dev \
        python3 \
        pkg-config \
        xz-utils \
        file \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-7 100 \
    && update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-7 100 \
    && update-alternatives --install /usr/bin/cc cc /usr/bin/gcc-7 100 \
    && update-alternatives --install /usr/bin/c++ c++ /usr/bin/g++-7 100

# ------------------------------------------------------------------------------
# 2. Python 3.9.11 (aligned with CI; focal ships 3.8 which is too old)
#    Built from source and installed to /usr/local so that python3/pip3
#    resolve to 3.9.11 (PATH precedence: /usr/local/bin before /usr/bin).
#    --enable-optimizations is skipped to keep build time reasonable.
# ------------------------------------------------------------------------------
ARG PYTHON_VERSION=3.9.11
RUN cd /tmp \
    && wget -q https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz \
    && tar -xzf Python-${PYTHON_VERSION}.tgz \
    && cd Python-${PYTHON_VERSION} \
    && ./configure --prefix=/usr/local \
    && make -j$(nproc) \
    && make install \
    && cd / && rm -rf /tmp/Python-${PYTHON_VERSION}* \
    && python3 --version && pip3 --version

# ------------------------------------------------------------------------------
# 3. cmake 3.20.5 (aligned with CI; focal apt only provides 3.16.x)
#    mindio requires cmake >= 3.14.1. Use the official prebuilt binary.
# ------------------------------------------------------------------------------
ARG CMAKE_VERSION=3.20.5
RUN cd /tmp \
    && wget -q --tries=3 --timeout=60 https://github.com/Kitware/CMake/releases/download/v${CMAKE_VERSION}/cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz \
    && tar -xzf cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz -C /opt \
    && ln -sf /opt/cmake-${CMAKE_VERSION}-linux-x86_64/bin/cmake /usr/local/bin/cmake \
    && ln -sf /opt/cmake-${CMAKE_VERSION}-linux-x86_64/bin/ctest /usr/local/bin/ctest \
    && rm -f /tmp/cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz

# ------------------------------------------------------------------------------
# 4. Go 1.26 (official binary, statically linked, runs on focal)
#    Component go.mod files require go 1.26; focal apt only has golang-1.13.
# ------------------------------------------------------------------------------
ARG GO_VERSION=1.26.4
RUN cd /tmp \
    && wget -q https://mirrors.aliyun.com/golang/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm -f /tmp/go${GO_VERSION}.linux-amd64.tar.gz

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH="${GOROOT}/bin:${GOPATH}/bin:${PATH}"

# Go module proxy (CN mirror) / sumdb (CN gateway) / pin toolchain to local
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
ENV GOTOOLCHAIN=local
ENV GO111MODULE=on

# ------------------------------------------------------------------------------
# 5. musl-gcc (PIE-capable, used to build volcano vc-scheduler)
#    volcano build.sh hardcodes CC=/usr/local/musl/bin/musl-gcc.
#    The distro musl-tools package ships a libmusl.a built without -fPIC,
#    which breaks PIE linking with "R_X86_64_32 ... recompile with -fPIC".
#    Build musl 1.2.0 (aligned with CI) with -fPIC and install to /usr/local/musl.
#    The musl-gcc wrapper drives the system gcc-7.5.0 (pinned above), so the
#    resulting binary .comment reads "GCC: (Ubuntu 7.5.0-...) 7.5.0",
#    matching the CI artifact.
# ------------------------------------------------------------------------------
ARG MUSL_VERSION=1.2.0
RUN cd /tmp \
    && wget -q https://musl.libc.org/releases/musl-${MUSL_VERSION}.tar.gz \
    && tar -xzf musl-${MUSL_VERSION}.tar.gz \
    && cd musl-${MUSL_VERSION} \
    && CFLAGS="-fPIC -O2" ./configure --prefix=/usr/local/musl --disable-shared \
    && make -j$(nproc) \
    && make install \
    && cd / && rm -rf /tmp/musl-${MUSL_VERSION}*

# ------------------------------------------------------------------------------
# 6. Go test tools (gotestsum / gocov / gocov-html)
#    Component test.sh scripts invoke these from $GOPATH/bin.
# ------------------------------------------------------------------------------
# gotestsum installs directly (its deps already support Go 1.26).
RUN go install gotest.tools/gotestsum@latest

# gocov / gocov-html pin an old golang.org/x/tools in their go.mod which is
# incompatible with Go 1.26 (tokeninternal compile error). Clone the sources,
# upgrade x/tools, then install locally to bypass module graph pruning.
RUN mkdir -p /tmp/gotools && cd /tmp/gotools \
    && git clone --depth=1 https://github.com/axw/gocov.git \
    && cd gocov \
    && go get golang.org/x/tools@v0.36.0 \
    && go mod tidy \
    && go install ./gocov \
    && cd /tmp/gotools \
    && git clone --depth=1 https://github.com/matm/gocov-html.git \
    && cd gocov-html \
    && go get golang.org/x/tools@v0.36.0 \
    && go mod tidy \
    && go install $(go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./...) \
    && cd / && rm -rf /tmp/gotools

# ------------------------------------------------------------------------------
# 7. helm-deploy-tool deps: yq (mikefarah) + helm
#    Consumed by build_all.sh when building helm-deploy-tool.
# ------------------------------------------------------------------------------
RUN go install github.com/mikefarah/yq/v4@v4.44.3 \
    && go install helm.sh/helm/v3/cmd/helm@v3.15.4

# ------------------------------------------------------------------------------
# 8. swig >= 4.2.0 (hard dependency for mindio abi3 wheels)
#    focal apt provides swig 4.0.1, still below the 4.2 minimum; install a
#    standalone swig via pip.
# ------------------------------------------------------------------------------
RUN pip3 install "swig>=4.2,<4.4"

# ------------------------------------------------------------------------------
# 9. Python component deps (ascend-faultdiag / taskd)
#    Use the CN PyPI mirror for speed.
# ------------------------------------------------------------------------------
RUN pip3 install -i https://mirrors.aliyun.com/pypi/simple/ \
        scikit-learn \
        pandas \
        numpy \
        ply \
        paramiko \
        scp \
        cryptography \
        openpyxl \
        grpcio \
        protobuf \
        pyOpenSSL \
        pytest \
        pytest-cov \
        pytest-html \
        coverage

# ------------------------------------------------------------------------------
# 10. Pre-cache volcano sources (required by build_all.sh / build_each.sh)
#     Supported versions: v1.7.0, v1.9.0, v1.12.0 (build.sh default).
#     Fetched as release-branch tarballs from the upstream github repo.
#     A plain `git clone` is not usable here: focal's git is built against
#     gnutls, which fails the TLS 1.3 handshake with github ("The TLS
#     connection was non-properly terminated"). wget uses openssl and goes
#     through the github CDN, so it is stable. The build scripts only need
#     the source tree (no git history), so a tarball is equivalent.
# ------------------------------------------------------------------------------
RUN mkdir -p /opt/buildtools/volcano_opensource/volcano_1.12 \
    && mkdir -p /opt/buildtools/volcano_opensource/volcano_1.9 \
    && mkdir -p /opt/buildtools/volcano_opensource/volcano_1.7 \
    && wget -q --tries=3 --timeout=60 https://github.com/volcano-sh/volcano/archive/refs/heads/release-1.12.tar.gz -O /tmp/v1.12.tar.gz \
    && tar -xzf /tmp/v1.12.tar.gz -C /opt/buildtools/volcano_opensource/volcano_1.12 \
    && mv /opt/buildtools/volcano_opensource/volcano_1.12/volcano-release-1.12 \
          /opt/buildtools/volcano_opensource/volcano_1.12/volcano \
    && rm -f /tmp/v1.12.tar.gz \
    && wget -q --tries=3 --timeout=60 https://github.com/volcano-sh/volcano/archive/refs/heads/release-1.9.tar.gz -O /tmp/v1.9.tar.gz \
    && tar -xzf /tmp/v1.9.tar.gz -C /opt/buildtools/volcano_opensource/volcano_1.9 \
    && mv /opt/buildtools/volcano_opensource/volcano_1.9/volcano-release-1.9 \
          /opt/buildtools/volcano_opensource/volcano_1.9/volcano \
    && rm -f /tmp/v1.9.tar.gz \
    && wget -q --tries=3 --timeout=60 https://github.com/volcano-sh/volcano/archive/refs/heads/release-1.7.tar.gz -O /tmp/v1.7.tar.gz \
    && tar -xzf /tmp/v1.7.tar.gz -C /opt/buildtools/volcano_opensource/volcano_1.7 \
    && mv /opt/buildtools/volcano_opensource/volcano_1.7/volcano-release-1.7 \
          /opt/buildtools/volcano_opensource/volcano_1.7/volcano \
    && rm -f /tmp/v1.7.tar.gz

# ------------------------------------------------------------------------------
# 11. Timezone (test cases assert on UTC+8 timestamps)
# ------------------------------------------------------------------------------
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# ------------------------------------------------------------------------------
# 12. Working directory
# ------------------------------------------------------------------------------
ENV MIND_CLUSTER_ROOT=/workspace
WORKDIR ${MIND_CLUSTER_ROOT}

# Default shell is bash (build scripts rely on bash syntax)
SHELL ["/bin/bash", "-c"]
