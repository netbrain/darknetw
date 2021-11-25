ARG CUDA_IMAGE=nvidia/cuda:11.2.2-cudnn8-devel
FROM $CUDA_IMAGE as base

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
	apt-get install -y --no-install-recommends \
            git build-essential cmake pkg-config unzip libgtk2.0-dev \
            curl ca-certificates libcurl4-openssl-dev libssl-dev \
            libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev \
            libjpeg-dev libpng-dev libtiff-dev libdc1394-22-dev \
            software-properties-common && \
            rm -rf /var/lib/apt/lists/*

FROM base as builder

ARG OPENCV_VERSION="4.2.0"
ARG GOVERSION="1.15"
ARG COMMIT_HASH=eb0272f27acda1982fe4d30acd838fca427785a9
ARG DARKNET_REPO=https://github.com/AlexeyAB/darknet

ENV GOVERSION $GOVERSION
ENV OPENCV_VERSION $OPENCV_VERSION

RUN curl -Lo opencv.zip https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip && \
            unzip -q opencv.zip && \
            curl -Lo opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip && \
            unzip -q opencv_contrib.zip && \
            rm opencv.zip opencv_contrib.zip && \
            cd opencv-${OPENCV_VERSION} && \
            mkdir build && cd build && \
            cmake -D CMAKE_BUILD_TYPE=RELEASE \
                  -D CMAKE_INSTALL_PREFIX=/usr/local \
                  -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib-${OPENCV_VERSION}/modules \
                  -D WITH_JASPER=OFF \
                  -D BUILD_DOCS=OFF \
                  -D BUILD_EXAMPLES=OFF \
                  -D BUILD_TESTS=OFF \
                  -D BUILD_PERF_TESTS=OFF \
                  -D BUILD_opencv_java=NO \
                  -D BUILD_opencv_python=NO \
                  -D BUILD_opencv_python2=NO \
                  -D BUILD_opencv_python3=NO \
                  -D OPENCV_GENERATE_PKGCONFIG=ON .. && \
            make -j $(nproc --all) && \
            #make preinstall && make install && ldconfig
            make preinstall && make install && ldconfig && \
            cd / && rm -rf opencv*


#fix libcuda.so.1 missing
RUN cp /usr/local/cuda/compat/* /usr/local/cuda/targets/x86_64-linux/lib/
ENV LIBRARY_PATH=$LIBRARY_PATH:/usr/local/cuda/compat/

WORKDIR /usr/src
RUN git clone ${DARKNET_REPO} darknet
WORKDIR darknet
RUN git checkout ${COMMIT_HASH}
COPY docker/Makefile .
RUN make

#GO
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /usr/src
RUN curl -Lo go${GOVERSION}.linux-amd64.tar.gz https://dl.google.com/go/go${GOVERSION}.linux-amd64.tar.gz && \
            tar -C /usr/local -xzf go${GOVERSION}.linux-amd64.tar.gz && \
            rm go${GOVERSION}.linux-amd64.tar.gz && \
            mkdir -p "$GOPATH/src" "$GOPATH/bin" && \
            chmod -R 777 "$GOPATH"

COPY . dn
WORKDIR dn
RUN cp /usr/src/darknet/libdarknet.so ./lib/ && \
    cp /usr/src/darknet/include/darknet.h ./include/ && \
    go build


FROM base
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/app/lib

COPY --from=builder /usr/src/dn/example /app/example/
COPY --from=builder /usr/src/dn/darknetw /app/
COPY --from=builder /usr/src/dn/lib /app/lib/
COPY --from=builder /usr/local/lib/libopencv* /app/lib/

WORKDIR /app
ENTRYPOINT /app/darknetw
