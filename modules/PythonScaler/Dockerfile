FROM ubuntu:latest

RUN apt-get update -qq -y && \
    apt-get install -y \
    build-essential \
    snapcraft \
    snapd \
    python3-pip \
    golang-go \
    python-mysqldb

RUN apt-get install -y software-properties-common
RUN add-apt-repository -yu ppa:juju/stable
RUN apt install -y juju

RUN  mkdir app
COPY . /app
COPY configs/ /root/.local/share/juju/.
COPY cookies/empty_cookie.json /root/.go-cookies
WORKDIR /app

RUN pip3 install --upgrade pip
RUN pip3 install -r requirements.txt

CMD ["python3","-u","main.py"]