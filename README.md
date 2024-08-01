# config-keeper
Config-keeper will help you to store and manage your configuration files for applications!

It'll have it's own API and frontend where each can be deployed separately on different hosts or even k8s using Docker-image.

Every config will have:
- multiple tags which yu can use to search in UI or even search using API
- multiple versions of one configuration file(very usefull for canary deployment if you need different configurations)
- callback endpoints which will be called if files content was updated(like subscriptions)

Some of features will be added lately like CLI and others.
