# Reproducible Central Index

> [reproducible-central](https://github.com/jvm-repo-rebuild/reproducible-central) is an effort to rebuild public releases published to Maven Central and check that Reproducible Build can be achieved.
> This projects aims to provide a machine-readable index for third party tools to look up the reproducibility status of a given artifact by group/artifact/version.

This index is not affiliated with the [Reproducible Builds](https://reproducible-builds.org/) project and a temporary solution until there is a official way to query the status of a given artifact.

## Status

The index is updated once per hour using GitHub Actions and published to GitHub Pages.

## Usage

The generated index files are hosted on GitHub Pages and can be accessed using the following URLs:

| URL                                                                                    | Description                    |
|----------------------------------------------------------------------------------------|--------------------------------|
| `https://philippheuer.me/reproducible-central-index/index.json`                        | All artifacts                  |
| `https://philippheuer.me/reproducible-central-index/{group}/{artifact}/index.json`     | By group and artifact          |
| `https://philippheuer.me/reproducible-central-index/{group}/{artifact}/{version}.json` | By group, artifact and version |

_Examples:_

com.fasterxml.jackson.core:jackson-databind

`curl https://philippheuer.me/reproducible-central-index/com/fasterxml/jackson/core/jackson-databind/index.json`

com.fasterxml.jackson.core:jackson-databind:2.17.0

`curl https://philippheuer.me/reproducible-central-index/com/fasterxml/jackson/core/jackson-databind/2.17.0.json`

## Badges

...

## License

The code is released under the [MIT license](./LICENSE).
