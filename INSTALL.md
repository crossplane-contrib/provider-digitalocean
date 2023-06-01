# Building and Installing the Crossplane DigitalOcean Provider

`provider-digitalocean` is composed of a golang project and can be built directly with standard `golang` tools. We currently support two different platforms for building:

* Linux: most modern distros should work although most testing has been done on Ubuntu
* Mac: macOS 10.6+ is supported

## Build Requirements

An Intel-based machine (recommend 2+ cores, 2+ GB of memory and 128GB of SSD). Inside your build environment (Docker for Mac or a VM), 6+ GB memory is also recommended.

The following tools are need on the host:

* curl
* docker (1.12+) or Docker for Mac (17+)
* git
* make
* golang (v1.18)
* rsync (if you're using the build container on mac)
* helm (v2.8.2+)
* kubebuilder (v1.0.4+)

## Build
You can build the Crossplane DigitalOcean Provider for the host platform by simply running the command below.
Building in parallel with the `-j` option is recommended.

```console
make -j4
```

The first time `make` is run, the build submodule will be synced and
updated. After initial setup, it can be updated by running `make submodules`.

Run `make help` for more options.

## Building inside the cross container

Official Crossplane builds are done inside a build container. This ensures that we get a consistent build, test and release environment. To run the build inside the cross container run:

```console
build/run make -j4
```

The first run of `build/run` will build the container itself and could take a few minutes to complete, but subsequent builds should go much faster.

## Install Crossplane in Your Cluster
Once your Kind Cluster is up and running, you'll need to install Crossplane. 

We recommend using Helm to install Crossplane. You can find the [official documentation here](https://crossplane.io/docs/v1.5/getting-started/install-configure.html#install-crossplane). These are the commands: 

```console
kubectl create namespace crossplane-system

helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

helm install crossplane --namespace crossplane-system crossplane-stable/crossplane
```

## Install the DigitalOcean Crossplane Provider

1. Find the provider installation file at [examples/provider/install.yaml](./examples/provider/install.yaml)
2. Run the installation:

```console
kubectl apply -f https://raw.githubusercontent.com/crossplane-contrib/provider-digitalocean/main/examples/provider/install.yaml
```

## Configure the DigitalOcean Crossplane Provider 
1. Find the provider config file at [examples/provider/config.yaml](./examples/provider/config.yaml)
1. [Create a DigitalOcean personal access token](https://docs.digitalocean.com/reference/api/create-personal-access-token/)
1. Encode that token using base64, and in the `config.yaml` file, replace `BASE64ENCODED_PROVIDER_CREDS` with your encoded token
1. Create a new Secret and ProviderConfig with 
```console
kubectl apply -f config.yaml
```
1. Check that the Provider has been created by running 

```console
kubectl get ProviderConfig
```

You should see output similar to this: 

```console
NAME      AGE
example   34s
```

## Provision DigitalOcean Resources 
Once you have Crossplane installed in your cluster, and you've created created the DigitalOcean `ProviderConfig` resource, you can start spinning up DigitalOcean resources like Droplets, Managed Databases, and other DOKS clusters. 

Go to the [examples](./examples) directory in this repo, find the DigitalOcean product you'd like to spin up, make any needed changes to the yaml file, and then create the resource. 

For example, if you'd like to spin up a DigitalOcean Droplet, run the command

```console 
kubectl apply -f examples/compute
```
and check your DigitalOcean account to see if the Droplet has been created.
