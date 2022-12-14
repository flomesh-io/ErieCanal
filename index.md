# ErieCanal Helm Chart Repo 

![GitHub](https://img.shields.io/github/license/flomesh-io/ErieCanal)

## Usage

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```console
helm repo add erie-canal https://flomesh-io.github.io/ErieCanal
```

Then you're good to install ErieCanal:

```console
helm install erie-canal erie-canal/erie-canal --namespace erie-canal --create-namespace
```

## License
[Apache-2.0 License](https://github.com/flomesh-io/ErieCanal/blob/main/LICENSE).
