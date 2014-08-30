Dockertie
==========

Tiny Docker coordinator

Usage
-----

```
$ export AWS_ACCESS_KEY_ID=xxxxxxxxxxx
$ export AWS_SECRET_KEY=xxxxxxxxxxx
$ dockertie --discoverer ec2 --containerizer docker
```

### GET /hosts

List all hosts

### GET /hosts/HOST_ID/container

List containers on the host

### GET /containers

List all containers

### POST /containers

Create a container on available host.

Discoverer
----------

Discoverer is what finds hosts.

Currently, `ec2` and `json` discoverers are supported.

Containerizer
-------------

Containerizer is what create containers.

Currently, `docker` containerizer is supported.


