# Belvedere

_A small lookout tower (usually square) on the roof of a house._

Belvedere is a small, opinionated tool for deploying HTTP2-based apps on GCP in an operationally friendly manner.
It handles load balancing, DNS, TLS certificates, autoscaling, deploys, logging, metrics, tracing, permissions, IAP access control, CDN setup, and more.

## Setup

First, create a GCP project and set it as the default using `gcloud config set core/project $BLAH`.
(Or pass the `--project` flag everywhere.)
Make sure it has a billing account associated with it.

Once the project is created, pick a DNS zone in which the apps will live.
Let's say you've recently registered the domain `cornbread.club`.

```shell script
belvedere setup cornbread.club
```

This will enable all the required GCP APIs, grant Deployment Manager permissions to manage IAM roles, and create a Deployment Manager deployment with a managed zone for `cornbread.club` plus a few firewall rules for securing Belvedere-managed apps.
(Don't worry, they won't affect anything else in the project.)

Once `setup` has been run, the domain's DNS settings will need to be configured at the register.
To get a list of the DNS servers to use, run:

```shell script
belvedere dns-servers
```

It's important that this be done before creating any apps, as creating an app involves provisioning a Google-managed TLS certificate.
In order for that provisioning to be successful, the hostname associated with the app (e.g. `my-app.cornbread.club`) must resolve to the load balancer's IP address.
In order to do that, the domain registration needs to provide the servers listed via `dns-servers` as the DNS servers for that hostname.
If this isn't the case when an app is created, the certificate will take much longer to provision.

## Apps

Belvedere apps are HTTP2 apps packaged as containers in a registry, running on virtual machines with optional sidecars.
Belvedere requires that the app (or a sidecar reverse proxy) listen on port `8443` for HTTP2 requests.
The app can use self-signed certs, but it must use TLS.
The app must return a `200 OK` response to `GET` requests for `/healthz`

### Building An App

This repository has an example "Hello, world!" app in the `examples` directory.
It consists of an HTTP/1.1 Go app and an Nginx-based frontend proxy.

To build the app using Google Cloud Build, run:

```shell script
gcloud builds submit --config ./examples/helloworld/cloudbuild.yaml ./examples/helloworld/
```

To build the Nginx frontend using Google Cloud Build, run:

```shell script
gcloud builds submit --config ./examples/nginx-frontend/cloudbuild.yaml ./examples/nginx-frontend/
```

This will result in images for the app and the frontend proxy being built and pushed to Google Container Registry in the GCP project.

### Configuration

App configuration lives in a YAML file.
Check out `examples/helloworld.yaml` for an example.
Belvedere requires a Google Compute Engine machine type and a Docker image URL for the app's main container.

### Creating An App

To create an app, pick a GCE region (e.g. `us-central1`) and run:

```shell script
belevedere apps create my-app us-central1 ./my-app.yaml
```

This will create a Deployment Manager deployment with a bunch of goodies:

* a global, HTTP2 load balancing stack, with support for QUIC
* a DNS A record for `my-app.cornbread.club` pointing to the load balancer
* a Let's Encrypt-provided, Google-managed TLS certificate for `my-app.cornbread.club`
* a service account with the specified IAM roles, plus baked in access to key services:
  - Stackdriver Metrics
  - Stackdriver Logging
  - Google Container Registry
  - Cloud Debugger
  - Cloud Profiler
  - Stackdriver Error Reporting

The load balancer and DNS stuff will take 10-30 minutes to fully provision.

### Creating A Release

To create a release for an app, get the SHA256 hash of the container image and run:

```shell script
belvedere releases create my-app v1 ./my-app.yaml $SHA256
```

This will create a Deployment Manager deployment with some more goodies:

* an instance template for creating instances running the app inside Docker on Google Container-Optimized OS
* an instance group manager for manging those instances
* an autoscaler for scaling the number of app instances up or down based on load balancer utilization

Once this is done, the release has been created but is not in service.

### Enabling A Release

To direct traffic to the instances in a release, enable the release by running:

```shell script
belvedere releases enable my-app v1
```

This will add the `v1` instance group to the load balancer and wait for the instances to register,
pass health checks, and go into service. If the instances aren't healthy after 5 minutes,
`belvedere` will exit with a non-zero status.

### Disabling A Release

To remove a release from service, disable it by running:

```shell script
belvedere releases disable my-app v1
```

This will remove the app from service and drain any existing connections.

### Deleting A Release

To delete all of the resources associated with a release, including the instances, run:

```shell script
belvedere releases delete my-app v1
```

### Deleting An App

To delete all of the resources associated with an app, run:

```shell script
belvedere apps delete my-app
```

## Operational Amenities

### Listing Apps

To list all the apps in the project, run:

```shell script
belvedere apps list
```

### Listing Releases

To list all the releases in the project, run:

```shell script
belvedere releases list
belvedere releases list my-app
```

### Listing Instances

To list all the running instances in the project, run:

```shell script
belvedere instances
belvedere instance my-app
belvedere instance my-app v43
```

### SSH Access

To SSH into a particular instance, run:

```shell script
belvedere ssh my-app-v43-hxht
```

This will use `gcloud` to automatically configure an SSH key, inject it into the instance, and tunnel an SSH connection over GCP's Identity-Aware Proxy (IAP) to the instance.
IAP tunneling is used because it allows for public SSH access to app instances to be disabled.
Only IAP tunnels are allowed, and IAP tunnels require that the initiator be an authenticated member of the GCP project.

You can also pass arguments to SSH:

```shell script
belvedere ssh my-app-v43-hxht -- ls -al
```

### Viewing Logs

To view the logs for an app and its sidecar containers, run:

```shell script
belvedere logs my-app
belvedere logs my-app v43
belvedere logs my-app v43 --freshness=1h
belvedere logs my-app v43 --freshness=1h --filter="/login/"
```

## TODO

- [ ] Block external access to `/healthz`
- [ ] Canary deploys
- [ ] Run containers as a non-root user
- [ ] Session affinity
- [ ] GPU accelerator support
- [ ] Redirect HTTP to HTTPS on LB
