<table align="center"><tr><td align="center" width="9999">
<img src="logo.png" align="center" width="150" style="border-radius:60%;">

# The Misakey Backend Project.

[![pipeline](https://gitlab.misakey.dev/misakey/backend/badges/master/pipeline.svg)](https://gitlab.misakey.dev/misakey/backend/-/pipelines)
[![api doc](https://img.shields.io/badge/doc-api-blue)](https://backend.docs.misakey.dev)
[![License AGPLv3](https://img.shields.io/static/v1?label=License&message=AGPLv3&color=e32e72)](./LICENSE)

</td></tr></table>

## Introduction

[Misakey](https://misakey.com) is the user account solution for people and applications who
value privacy and simplicity.

You can find more info about Misakey in our [website](https://www.misakey.com) and our [about page](https://about.misakey.com/).

## Folder architecture

The project is composed of:
* `api`: the main API service, crafted with Golang. 
* `hydra`: configuration to run an instance of [Ory Hydra](https://github.com/ory/hydra) to manage the auth protocol.
* `gateway`: an nginx gateway configuration.
* `tools`: some script helpers and functional testing.
* `docs`: the concepts and endpoints documentation.

We host the platform on a Kubernetes cluster, so everything we build is done to work in Docker and k8s.

## Community

We don't have tools to welcome community for now. 

You want to talk with us, or contribute to the project? 
You can open an issue, or contact us by email at [love@misakey.com](mailto:love@misakey.com)!.

We will answer you quickly and would love to hear feedback from you!

## License

Most of the code is released under the AGPLv3. 
If subdirectories include a different license, that license applies instead.

## Source management disclaimer

Misakey uses GitLab for the development of its free softwares. Our Github repositories are only mirrors.

## Cryptography notice

This distribution and associated services include cryptographic software. 
The country in which you currently reside may have restrictions on the import, possession, use, 
and/or re-export to another country, of encryption software. BEFORE using any encryption software, 
please check your country's laws, regulations and policies concerning the import, possession, 
or use, and re-export of encryption software, to see if this is permitted. 
See http://www.wassenaar.org/ for more information.
