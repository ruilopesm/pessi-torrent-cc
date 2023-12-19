FROM gh0st42/coreemu7:1.0.0
LABEL Description="CORE Docker Ubuntu Image"

WORKDIR /home/pessi-torrent

# install go

COPY --from=golang:1.21.1 /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

# copy project files

COPY ./go.mod .
COPY ./cmd ./cmd
COPY ./internal ./internal

# build project

RUN go get ./cmd/tracker
RUN go get ./cmd/node
RUN go build -o /usr/local/bin/pessi-tracker ./cmd/tracker
RUN go build -o /usr/local/bin/pessi-node ./cmd/node

COPY ./topologies/CC-Topo-2023.imn /usr/local/lib/core/CC-Topo-2023.imn
COPY ./config/config.coreemu.yml /config/config.yml

COPY ./topologies /topologies
COPY bin/dns /dns

RUN mkdir /downloads-1
RUN mkdir /downloads-2
RUN mkdir /downloads-3
RUN mkdir /downloads-4
RUN mkdir /downloads-5

RUN mkdir /files
RUN fallocate -l 10M /files/10M.file
RUN fallocate -l 100M /files/100M.file
RUN fallocate -l 500M /files/500M.file
RUN fallocate -l 1G /files/1G.file

# Install add-apt-repository command
RUN apt-get -qqqy update
RUN apt-get -qqqy dist-upgrade
RUN apt-get -qqqy install --no-install-recommends dnsutils apt-utils software-properties-common dctrl-tools gpg-agent

# Add the BIND 9 APT Repository
RUN add-apt-repository -y ppa:isc/bind-esv

ARG DEB_VERSION=1:9.16.45-1+ubuntu20.04.1+deb.sury.org+1

# Install BIND 9
RUN apt-get -qqqy update
RUN apt-get -qqqy dist-upgrade
RUN apt-get -qqqy install bind9=$DEB_VERSION bind9utils=$DEB_VERSION

# Now remove the pkexec that got pulled as dependency to software-properties-common
RUN apt-get --purge -y autoremove policykit-1

RUN mkdir -p /etc/bind && chown root:bind /etc/bind/ && chmod 755 /etc/bind
RUN mkdir -p /var/cache/bind && chown bind:bind /var/cache/bind && chmod 755 /var/cache/bind
RUN mkdir -p /var/lib/bind && chown bind:bind /var/lib/bind && chmod 755 /var/lib/bind
RUN mkdir -p /var/log/bind && chown bind:bind /var/log/bind && chmod 755 /var/log/bind
RUN mkdir -p /run/named && chown bind:bind /run/named && chmod 755 /run/named

RUN echo 'core-daemon > /var/log/core-daemon.log 2>&1 & \n\
sleep 1 \n\
core-gui' > /root/entryPoint.sh

ENTRYPOINT ["/bin/sh", "-c", "\"/root/entryPoint.sh\""]
