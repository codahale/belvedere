# Belvedere

_A small lookout tower (usually square) on the roof of a house._

Belvedere is a small, opinionated tool for deploying HTTP2-based apps on GCP in an operationally friendly manner.
It handles load balancing, DNS, TLS certificates, autoscaling, deploys, logging, metrics, tracing, permissions, IAP access control, CDN setup, WAF rules, and more.

## Setup

First, create a GCP project and set it as the default using `gcloud config set core/project $BLAH`.
(Or pass the `--project` flag everywhere.)
Make sure it has a billing account associated with it.

Once the project is created, pick a DNS zone in which the apps will live.
Let's say you've recently registered the domain `cornbread.club`.

```
belvedere setup cornbread.club
```

This will enable all the required GCP APIs, grant Deployment Manager permissions to manage IAM roles, and create a Deployment Manager deployment with a managed zone for `cornbread.club` plus a few firewall rules for securing Belvedere-managed apps.
(Don't worry, they won't affect anything else in the project.)

Once `setup` has been run, the domain's DNS settings will need to be configured at the register.
To get a list of the DNS servers to use, run:

```
belvedere dns-servers
```

It's important this be done before creating any apps, as creating an application involves provisioning a Google-managed TLS certificate.
In order for that provisioning to be successful, the hostname associated with the application (e.g. `my-app.cornbread.club`) must resolve to the load balancer's IP address.
In order to do that, the domain registration needs to provide the servers listed via `dns-servers` as the DNS servers for that hostname.
If this isn't the case when an application is created, the certificate will take much longer to provision.

## Apps

Belvedere apps are HTTP2 apps packaged as containers in a registry, running on virtual machines with optional sidecars.
Belvedere requires that the application (or a sidecar reverse proxy) listen on port `8443` for HTTP2 requests.
The application can use self-signed certs, but it must use TLS.
The application must return a `200 OK` response to `GET` requests for `/healthz`.

### Building An App

This repository has an example "Hello, world!" application in the `examples` directory.
It consists of an HTTP/1.1 Go application and an Nginx-based frontend proxy.

To build the application using Google Cloud Build, run:

```
gcloud builds submit --config ./examples/helloworld/cloudbuild.yaml ./examples/helloworld/
```

To build the Nginx frontend using Google Cloud Build, run:

```
gcloud builds submit --config ./examples/nginx-frontend/cloudbuild.yaml ./examples/nginx-frontend/
```

This will result in images for the application and the frontend proxy being built and pushed to Google Container Registry in the GCP project.

### Configuration

Application configuration lives in a YAML file.
Check out `examples/helloworld.yaml` for an example.
Belvedere requires a Google Compute Engine machine type and a Docker image URL for the app's main container.

### Creating An App

To create an app, pick a GCE region (e.g. `us-central1`) and run:

```
belevedere apps create my-app us-central1 ./my-app.yaml
```

You can also pipe the config in via STDIN:

```shell script
cat ./my-app.yaml | belevedere apps create my-app us-central1
```

This will create a Deployment Manager deployment with a bunch of goodies:

* a global, HTTP2 load balancing stack, with support for QUIC
* a DNS A record for `my-app.cornbread.club` pointing to the load balancer
* a Let's Encrypt-provided, Google-managed TLS certificate for `my-app.cornbread.club`
* a Cloud Armor WAF with configurable rules for protecting your app
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

```
belvedere releases create my-app v1 $SHA256 ./my-app.yaml 
```

This will create a Deployment Manager deployment with some more goodies:

* an instance template for creating instances running the application inside Docker on Google Container-Optimized OS
* an instance group manager for manging those instances
* an autoscaler for scaling the number of application instances up or down based on load balancer utilization

Once this is done, the release has been created but is not in service.

### Enabling A Release

To direct traffic to the instances in a release, enable the release by running:

```
belvedere releases enable my-app v1
```

This will add the `v1` instance group to the load balancer and wait for the instances to register,
pass health checks, and go into service. If the instances aren't healthy after 5 minutes,
`belvedere` will exit with a non-zero status.

### Disabling A Release

To remove a release from service, disable it by running:

```
belvedere releases disable my-app v1
```

This will remove the application from service and drain any existing connections.

### Deleting A Release

To delete all the resources associated with a release, including the instances, run:

```
belvedere releases delete my-app v1
```

### Deleting An App

To delete all the resources associated with an app, run:

```
belvedere apps delete my-app
```

## Operational Amenities

### Listing Apps

To list all the apps in the project, run:

```
belvedere apps list
```

### Listing Releases

To list all the releases in the project, run:

```
belvedere releases list
belvedere releases list my-app
```

### Listing Instances

To list all the running instances in the project, run:

```
belvedere instances
belvedere instances my-app
belvedere instances my-app v43
```

### SSH Access

To SSH into a particular instance, run:

```
belvedere ssh my-app-v43-hxht
```

This will use `gcloud` to automatically configure an SSH key, inject it into the instance, and tunnel an SSH connection over GCP's Identity-Aware Proxy (IAP) to the instance.
This uses IAP tunneling because it allows for public SSH access to application instances to be disabled.
Belvedere's configuration only allows SSH over IAP tunnels, and IAP tunnels require that the initiator be an authenticated member of the GCP project.

You can also pass arguments to SSH:

```
belvedere ssh my-app-v43-hxht -- ls -al
```

### Viewing Logs

To view the logs for an application and its sidecar containers, run:

```
belvedere logs my-app
belvedere logs my-app v43
belvedere logs my-app v43 my-app-v43-hxht
belvedere logs my-app v43 my-app-v43-hxht --max-age=1h
belvedere logs my-app v43 my-app-v43-hxht --max-age=1h --filter="/login/"
```

### Secrets

Secrets (e.g. database passwords, API keys, etc.) are stored in [Google Secret Manager](https://cloud.google.com/secret-manager/docs).
This provides you with encryption at rest, encryption in flight, access control, and extensive audit logging.

#### Creating And Updating Secrets

You can create a secret with an initial value of a file's contents:

```
belvedere secrets create my-secret secret-value.txt
```

Or pipe the value in via STDIN:

```shell script
echo "super secret" | belvedere secrets create my-secret 
```

Updating a secret's value works similarly:

```
belvedere secrets update my-secret secret-value.txt
```

#### Listing And Deleting Secrets

It's about what you'd expect:

```
belvedere secrets list
belvedere secrets delete my-secret
```

#### Granting And Revoking Access

You can quickly grant or revoke an app's access to a secret:

```
belvedere secrets grant secret1 my-app
belvedere secrets revoke secret1 my-app
```

#### Accessing Secrets From Your Application

If you include [Berglas](https://github.com/GoogleCloudPlatform/berglas) in your application's Docker image, you can use it to convert environment variables of the form `sm://project-id/secret-id` into the secret's current value.
The resulting plaintext secrets will only ever be stored in memory.

## TODO

- [ ] Canary deploys
- [ ] Run containers as a non-root user
- [ ] GPU accelerator support
- [ ] Redirect HTTP to HTTPS on LB
- [ ] Traffic Director integration (https://github.com/google-cloud-sdk-unofficial/google-cloud-sdk/blob/e3c7770d324cedd5aeb4df6741de9a2c26235597/lib/surface/compute/instance_templates/create.py#L328)
