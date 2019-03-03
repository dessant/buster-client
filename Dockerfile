FROM golang:1.11.5-stretch

RUN set -ex \
    && apt-get update \
    && apt-get install -y libx11-dev xorg-dev libxtst-dev libpng++-dev xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev libxkbcommon-dev xsel xclip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/*

WORKDIR /src
