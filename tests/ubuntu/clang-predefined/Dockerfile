ARG UBUNTU_VERSION
FROM ghcr.io/rse-radiuss/ubuntu:$UBUNTU_IMAGE

# Version of LLVM
# This is an example of a manual type of build arg
ARG LLVM_VERSION

# BUild args will persist in environment?
ENV llvmver=$LLVM_VERSION
ENV llvmtar=clang+llvm-${llvmver}-x86_64-linux-gnu-ubuntu-${UBUNTU_VERSION}
ENV tarext=.tar.xz

RUN wget -q --no-check-certificate http://releases.llvm.org/${llvmver}/${llvmtar}${tarext} && \
    tar xf ${llvmtar}${tarext} && \
    sudo cp -fR ${llvmtar}/* /usr && \
    rm -rf ${llvmtar} && \
    rm ${llvmtar}${tarext}
