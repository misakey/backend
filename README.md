Misakey
=======

## What is Misakey ?

[Misakey](https://misakey.com) is a platform to secure data portability. Users can accustom websites 
to send their data back to them avoiding the big tech. Websites can easily return their users' data using a secure path.

You can find more info about Misakey in our [about page (french only)](https://docs.misakey.com/).

## Technical overview

The project is composed of:
* The API service, crafted with Golang. 
* An instance of **hydra** ([repo](https://github.com/ory/hydra)) to manage our auth protocol

We host the platform on a Kubernetes cluster, so everything we build is done to work in Docker and K8s.

## Misakey's source code

### Backend

####  API

- [API](./api/README.md)

#### Jobs

- [notification-job](./notification-job/README.md)

#### SDK

- [msk-sdk-go](https://gitlab.misakey.dev/misakey/msk-sdk-go/README.md)

### Frontend

#### Webapp & webextension

- [frontend](https://gitlab.misakey.dev/misakey/frontend/README.md)

#### Management tool for applications

- [misadmin-frontend](https://gitlab.misakey.dev/misakey/misadmin-frontend/README.md)


## Community

We don't have tools to welcome community for now. 

You want to talk with us, or contribute to the project? 
[Send us an email](mailto:question.perso@misakey.com)!
We will answer you rapidly and would love to hear what community tools you would like!

## License

Most code is released under the AGPLv3. 
If subdirectories include a different license, that license applies instead.

## Cryptography notice

This distribution and associated services include cryptographic software. 
The country in which you currently reside may have restrictions on the import, possession, use, 
and/or re-export to another country, of encryption software. BEFORE using any encryption software, 
please check your country's laws, regulations and policies concerning the import, possession, 
or use, and re-export of encryption software, to see if this is permitted. 
See http://www.wassenaar.org/ for more information.
